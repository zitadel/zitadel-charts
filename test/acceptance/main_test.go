// Package acceptance_test contains end-to-end acceptance tests that validate
// full ZITADEL deployments including HTTP endpoints, gRPC APIs, browser-based
// login flows, and authenticated machine-to-machine communication.
//
// A K3s cluster is started automatically via testcontainers before any test
// runs. K3s ships with a bundled Traefik ingress controller which is customized
// via a HelmChartConfig manifest to use NodePort with dynamically mapped ports.
// The cluster is torn down when the suite completes. No external cluster or
// manual setup is required.
package acceptance_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/zitadel/zitadel-charts/test/internal/testcluster"
)

const k3sStartupTimeout = 5 * time.Minute

// httpsPort holds the dynamically mapped host port for the Traefik HTTPS
// NodePort (30443). It is set during TestMain and used by all test functions
// to construct API base URLs and configure ZITADEL's ExternalPort.
var httpsPort string

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

// run starts a K3s cluster with its bundled Traefik ingress controller,
// extracts the kubeconfig, discovers the dynamically mapped HTTPS port, and
// then executes the test suite. It returns the exit code from m.Run.
func run(m *testing.M) int {
	ctx, cancel := context.WithTimeout(context.Background(), k3sStartupTimeout)
	defer cancel()

	cluster, err := testcluster.Start(ctx)
	if err != nil {
		log.Printf("failed to start K3s cluster: %v", err)
		return 1
	}
	defer cluster.Cleanup()

	httpsPort = cluster.HTTPSPort

	return m.Run()
}
