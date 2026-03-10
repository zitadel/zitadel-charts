// Package smoke_test_test contains smoke tests that validate rendered Kubernetes
// resources produced by the Zitadel Helm chart. Tests assert on object metadata,
// spec fields, and structural correctness without routing live traffic.
//
// A K3s cluster is started automatically via testcontainers before any test
// runs. The cluster is torn down when the suite completes. No external cluster
// or manual setup is required.
package smoke_test_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/k3s"
)

// k3sStartupTimeout is the maximum time allowed for the K3s container to start
// and become ready. If the container does not start within this duration, the
// test suite aborts with an error.
const k3sStartupTimeout = 3 * time.Minute

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

// run starts a K3s cluster, extracts its kubeconfig into a temporary file, sets
// the KUBECONFIG environment variable so that all tests discover it, executes
// the test suite, and tears everything down. It returns the exit code from
// m.Run for use with os.Exit.
func run(m *testing.M) int {
	ctx, cancel := context.WithTimeout(context.Background(), k3sStartupTimeout)
	defer cancel()

	container, err := k3s.Run(ctx, "rancher/k3s:v1.31.6-k3s1")
	if err != nil {
		log.Printf("failed to start K3s container: %v", err)
		return 1
	}
	defer func() {
		if err := container.Terminate(context.Background()); err != nil {
			log.Printf("failed to terminate K3s container: %v", err)
		}
	}()

	kubeconfig, err := container.GetKubeConfig(ctx)
	if err != nil {
		log.Printf("failed to get kubeconfig from K3s: %v", err)
		return 1
	}

	kubeconfigFile, err := os.CreateTemp("", "k3s-kubeconfig-*.yaml")
	if err != nil {
		log.Printf("failed to create kubeconfig temp file: %v", err)
		return 1
	}
	defer func() {
		if err := os.Remove(kubeconfigFile.Name()); err != nil {
			log.Printf("failed to remove kubeconfig temp file: %v", err)
		}
	}()

	if _, err := kubeconfigFile.Write(kubeconfig); err != nil {
		log.Printf("failed to write kubeconfig: %v", err)
		return 1
	}
	if err := kubeconfigFile.Close(); err != nil {
		log.Printf("failed to close kubeconfig file: %v", err)
		return 1
	}

	if err := os.Setenv("KUBECONFIG", kubeconfigFile.Name()); err != nil {
		log.Printf("failed to set KUBECONFIG: %v", err)
		return 1
	}

	return m.Run()
}
