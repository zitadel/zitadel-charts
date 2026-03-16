package testcluster

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
)

// WithNamespace creates a unique namespace for a test, runs the provided
// function, and cleans up the namespace afterwards. The namespace name is
// generated using the current Unix nanosecond timestamp to ensure uniqueness
// across concurrent test runs.
//
// A 30-minute context timeout is provided to the callback. If the test fails,
// the namespace is preserved for debugging purposes.
func WithNamespace(t *testing.T, fn func(ctx context.Context, k *k8s.KubectlOptions)) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	namespace := "zitadel-test-" + strings.ToLower(random.UniqueId())
	k := k8s.NewKubectlOptions("", os.Getenv("KUBECONFIG"), namespace)

	k8s.CreateNamespace(t, k, namespace)
	defer func() {
		if !t.Failed() {
			k8s.DeleteNamespace(t, k, namespace)
		} else {
			t.Logf("Test failed, keeping namespace %s for debugging", namespace)
		}
	}()

	fn(ctx, k)
}
