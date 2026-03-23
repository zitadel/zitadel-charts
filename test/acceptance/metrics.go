package acceptance_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckMetrics verifies that Prometheus-compatible metrics endpoints are
// serving valid metric data for both the ZITADEL server and the Login UI.
//
// Both endpoints are validated via the Kubernetes API server pod proxy,
// which routes requests directly to the pod without requiring ingress
// exposure, port-forwarding, or tools inside the container.
//
//   - ZITADEL: /debug/metrics on port 8080
//   - Login:   /metrics on port 9464
//
// When useTLS is true, the ZITADEL proxy request uses HTTPS (required when
// the server has internal TLS enabled via selfSignedCert).
func CheckMetrics(ctx context.Context, t *testing.T, k *k8s.KubectlOptions, useTLS bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, k)
	if err != nil {
		t.Fatalf("failed to create k8s client: %v", err)
	}

	t.Run("zitadel", func(t *testing.T) {
		awaitCheck(ctx, t, 1*time.Minute, func(ctx context.Context) error {
			pods := k8s.ListPods(t, k, metav1.ListOptions{
				LabelSelector: "app.kubernetes.io/component=start",
			})
			if len(pods) == 0 {
				return fmt.Errorf("no zitadel pods found")
			}

			scheme := ""
			if useTLS {
				scheme = "https"
			}

			body, err := clientset.CoreV1().Pods(k.Namespace).
				ProxyGet(scheme, pods[0].Name, "8080", "/debug/metrics", nil).
				DoRaw(ctx)
			if err != nil {
				return fmt.Errorf("proxy request failed: %w", err)
			}

			if !strings.Contains(string(body), "go_goroutines") {
				return fmt.Errorf("metrics response does not contain expected metric 'go_goroutines'")
			}
			return nil
		}, "zitadel metrics check failed")
	})

	t.Run("login", func(t *testing.T) {
		awaitCheck(ctx, t, 1*time.Minute, func(ctx context.Context) error {
			pods := k8s.ListPods(t, k, metav1.ListOptions{
				LabelSelector: "app.kubernetes.io/component=login",
			})
			if len(pods) == 0 {
				return fmt.Errorf("no login pods found")
			}

			body, err := clientset.CoreV1().Pods(k.Namespace).
				ProxyGet("", pods[0].Name, "9464", "/metrics", nil).
				DoRaw(ctx)
			if err != nil {
				return fmt.Errorf("proxy request failed: %w", err)
			}

			content := string(body)
			if !strings.Contains(content, "# HELP") && !strings.Contains(content, "# TYPE") {
				return fmt.Errorf("login metrics response does not contain Prometheus format headers")
			}
			return nil
		}, "login metrics check failed")
	})
}
