// Package testcluster provides a self-contained K3s cluster for integration
// tests. It starts K3s via testcontainers with the bundled Traefik ingress
// controller, extracts the kubeconfig, writes it to a temporary file, and
// sets the KUBECONFIG environment variable so that Helm and kubectl operations
// discover the cluster automatically.
//
// Traefik is configured via a HelmChartConfig manifest to use NodePort with
// ports 30080 (HTTP) and 30443 (HTTPS).
// The Gateway API provider is enabled and Traefik automatically creates a
// GatewayClass and Gateway. The dynamically mapped host ports are available
// via Cluster.HTTPSPort and Cluster.HTTPPort.
package testcluster

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
)

//go:embed *.yaml
var manifests embed.FS

// k3sImage is the K3s container image used by all test suites.
const k3sImage = "rancher/k3s:v1.34.1-k3s1"

// Cluster holds a running K3s container and the temporary kubeconfig file
// that points to it.
type Cluster struct {
	// HTTPSPort is the dynamically mapped host port for the Traefik HTTPS
	// NodePort (30443).
	HTTPSPort string
	// HTTPPort is the dynamically mapped host port for the Traefik HTTP
	// NodePort (30080). This is also the port used by the Gateway API
	// listener ("web" entrypoint).
	HTTPPort  string
	container      *k3s.K3sContainer
	kubeconfigPath string
}

// Start creates a K3s cluster with Traefik enabled, extracts its kubeconfig
// into a temporary file, and sets the KUBECONFIG environment variable. The
// bundled Traefik ingress controller is customized via a HelmChartConfig
// manifest to use NodePort with dynamically mapped ports. The Gateway API
// provider is enabled and Traefik creates a GatewayClass ("traefik") and
// Gateway ("traefik-gateway") automatically. Call Cleanup when the cluster
// is no longer needed.
func Start(ctx context.Context) (*Cluster, error) {
	traefikPath, err := writeEmbeddedFile("traefik-config.yaml")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.Remove(traefikPath) }()

	container, err := k3s.Run(ctx, k3sImage,
		testcontainers.WithCmd("server", "--tls-san=localhost"),
		testcontainers.WithExposedPorts("30080/tcp", "30443/tcp"),
		k3s.WithManifest(traefikPath),
	)
	if err != nil {
		return nil, err
	}

	kubeconfig, err := container.GetKubeConfig(ctx)
	if err != nil {
		_ = container.Terminate(context.Background())
		return nil, err
	}

	kubeconfigFile, err := os.CreateTemp("", "k3s-kubeconfig-*.yaml")
	if err != nil {
		_ = container.Terminate(context.Background())
		return nil, err
	}

	if _, err := kubeconfigFile.Write(kubeconfig); err != nil {
		_ = kubeconfigFile.Close()
		_ = os.Remove(kubeconfigFile.Name())
		_ = container.Terminate(context.Background())
		return nil, err
	}
	if err := kubeconfigFile.Close(); err != nil {
		_ = os.Remove(kubeconfigFile.Name())
		_ = container.Terminate(context.Background())
		return nil, err
	}

	if err := os.Setenv("KUBECONFIG", kubeconfigFile.Name()); err != nil {
		_ = os.Remove(kubeconfigFile.Name())
		_ = container.Terminate(context.Background())
		return nil, err
	}

	httpsMapped, err := container.MappedPort(ctx, nat.Port("30443/tcp"))
	if err != nil {
		_ = os.Remove(kubeconfigFile.Name())
		_ = container.Terminate(context.Background())
		return nil, err
	}

	httpMapped, err := container.MappedPort(ctx, nat.Port("30080/tcp"))
	if err != nil {
		_ = os.Remove(kubeconfigFile.Name())
		_ = container.Terminate(context.Background())
		return nil, err
	}

	return &Cluster{
		HTTPSPort:      httpsMapped.Port(),
		HTTPPort:       httpMapped.Port(),
		container:      container,
		kubeconfigPath: kubeconfigFile.Name(),
	}, nil
}

// ApplyGatewayCRDs registers the Gateway API CRDs with the cluster's API
// server using kubectl apply. The K3s bundled traefik-crd chart already
// installs these CRDs, so this call is idempotent. It is kept for test
// suites (e.g. smoke tests) that call it explicitly.
func (c *Cluster) ApplyGatewayCRDs(ctx context.Context) error {
	path, err := writeEmbeddedFile("gateway-crds.yaml")
	if err != nil {
		return fmt.Errorf("extracting gateway CRDs: %w", err)
	}
	defer func() { _ = os.Remove(path) }()

	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", path)
	cmd.Env = append(os.Environ(), "KUBECONFIG="+c.kubeconfigPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("kubectl apply gateway CRDs: %w\n%s", err, out)
	}
	return nil
}

// ApplyServiceMonitorCRD registers the ServiceMonitor CRD with the cluster's
// API server using kubectl apply. This must be called after Start and only by
// test suites that need ServiceMonitor resources (e.g. smoke tests).
func (c *Cluster) ApplyServiceMonitorCRD(ctx context.Context) error {
	path, err := writeEmbeddedFile("servicemonitor-crd.yaml")
	if err != nil {
		return fmt.Errorf("extracting ServiceMonitor CRD: %w", err)
	}
	defer func() { _ = os.Remove(path) }()

	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", path)
	cmd.Env = append(os.Environ(), "KUBECONFIG="+c.kubeconfigPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("kubectl apply ServiceMonitor CRD: %w\n%s", err, out)
	}
	return nil
}

// Cleanup terminates the K3s container and removes the temporary kubeconfig
// file. Errors are silently ignored since this is typically called in a defer
// and the container will be reaped by Ryuk regardless.
func (c *Cluster) Cleanup() {
	_ = c.container.Terminate(context.Background())
	_ = os.Remove(c.kubeconfigPath)
	// Only unset KUBECONFIG if it still points to this cluster's kubeconfig.
	if os.Getenv("KUBECONFIG") == c.kubeconfigPath {
		_ = os.Unsetenv("KUBECONFIG")
	}
}

// writeEmbeddedFile extracts an embedded file to a temporary file so it can
// be passed to k3s.WithManifest.
func writeEmbeddedFile(name string) (string, error) {
	data, err := manifests.ReadFile(name)
	if err != nil {
		return "", err
	}

	tmpFile, err := os.CreateTemp("", "testcluster-*.yaml")
	if err != nil {
		return "", err
	}

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return "", err
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}
