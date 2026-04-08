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

	"github.com/zitadel/zitadel-charts/test/internal/testcluster"
)

// k3sStartupTimeout is the maximum time allowed for the K3s container to start
// and become ready. If the container does not start within this duration, the
// test suite aborts with an error.
const k3sStartupTimeout = 3 * time.Minute

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

// run starts a K3s cluster, sets up the KUBECONFIG environment variable,
// executes the test suite, and tears everything down. It returns the exit
// code from m.Run for use with os.Exit.
func run(m *testing.M) int {
	ctx, cancel := context.WithTimeout(context.Background(), k3sStartupTimeout)
	defer cancel()

	cluster, err := testcluster.Start(ctx)
	if err != nil {
		log.Printf("failed to start K3s cluster: %v", err)
		return 1
	}
	defer cluster.Cleanup()

	if err := cluster.ApplyGatewayCRDs(ctx); err != nil {
		log.Printf("failed to apply Gateway API CRDs: %v", err)
		return 1
	}

	if err := cluster.ApplyServiceMonitorCRD(ctx); err != nil {
		log.Printf("failed to apply ServiceMonitor CRD: %v", err)
		return 1
	}

	return m.Run()
}
