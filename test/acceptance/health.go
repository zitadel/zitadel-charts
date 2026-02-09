package acceptance_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	grpchelper "github.com/zitadel/zitadel-charts/test/acceptance/helpers/grpc"
	httphelper "github.com/zitadel/zitadel-charts/test/acceptance/helpers/http"
)

// CheckAccessibility performs comprehensive endpoint accessibility verification
// for a ZITADEL instance. It validates that all critical HTTP and gRPC endpoints
// are reachable and responding correctly through the ingress controller.
//
// The function runs the following checks with a 2-minute overall timeout:
//
//   - debug/validate: Internal validation endpoint returns HTTP 200
//   - debug/healthz: Kubernetes liveness probe endpoint returns HTTP 200
//   - debug/ready: Kubernetes readiness probe endpoint returns HTTP 200
//   - well-known/openid-configuration: OIDC discovery endpoint returns HTTP 200
//   - login page: Login UI (/ui/v2/login) loads without 5xx errors
//   - environment.json: Console config contains correct API and issuer URLs
//   - grpc healthz: gRPC management API responds to health check requests
//
// Each check uses eventual consistency with a 1-minute retry window and 1-second
// intervals to account for ingress routing propagation and pod startup delays.
// This is particularly important in Kind clusters where Traefik may take time
// to recognize new ingress routes.
//
// The environment.json check is critical as it verifies that the ZITADEL
// console will be able to communicate with the backend APIs. Misconfigured
// external domains are a common source of deployment issues.
func CheckAccessibility(ctx context.Context, t *testing.T, _ *k8s.KubectlOptions, apiBaseURL string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	checks := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"debug/validate", func(ctx context.Context) error {
			return checkHTTPEndpoint(ctx, apiBaseURL+"/debug/validate", 200)
		}},
		{"debug/healthz", func(ctx context.Context) error {
			return checkHTTPEndpoint(ctx, apiBaseURL+"/debug/healthz", 200)
		}},
		{"debug/ready", func(ctx context.Context) error {
			return checkHTTPEndpoint(ctx, apiBaseURL+"/debug/ready", 200)
		}},
		{"well-known/openid-configuration", func(ctx context.Context) error {
			return checkOIDCDiscovery(ctx, apiBaseURL)
		}},
		{"login page", func(ctx context.Context) error {
			return checkHTTPEndpointNot500(ctx, apiBaseURL+"/ui/v2/login")
		}},
		{"environment.json", func(ctx context.Context) error {
			return checkEnvironmentJSON(ctx, apiBaseURL)
		}},
		{"grpc healthz", func(ctx context.Context) error {
			return checkGRPCHealth(ctx, apiBaseURL)
		}},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			awaitCheck(ctx, t, 1*time.Minute, check.fn, "check %s failed", check.name)
		})
	}
}

func checkHTTPEndpoint(ctx context.Context, endpointURL string, expectedStatus int) error {
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statusCode, _, err := httphelper.Get(checkCtx, endpointURL, nil)
	if err != nil {
		return err
	}
	if statusCode != expectedStatus {
		return fmt.Errorf("expected status %d but got %d", expectedStatus, statusCode)
	}
	return nil
}

func checkHTTPEndpointNot500(ctx context.Context, endpointURL string) error {
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statusCode, _, err := httphelper.Get(checkCtx, endpointURL, nil)
	if err != nil {
		return err
	}
	if statusCode >= 500 {
		return fmt.Errorf("expected status < 500 but got %d", statusCode)
	}
	return nil
}

func checkEnvironmentJSON(ctx context.Context, apiBaseURL string) error {
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statusCode, body, err := httphelper.Get(checkCtx, apiBaseURL+"/ui/console/assets/environment.json", nil)
	if err != nil {
		return err
	}
	if statusCode != 200 {
		return fmt.Errorf("expected status 200 but got %d", statusCode)
	}

	bodyStr := string(body)
	parsedURL, _ := url.Parse(apiBaseURL)
	hostWithPort := parsedURL.Host

	if !containsAny(bodyStr, []string{
		fmt.Sprintf(`"api":"http://%s"`, hostWithPort),
		fmt.Sprintf(`"api":"https://%s"`, hostWithPort),
	}) {
		return fmt.Errorf("couldn't find api endpoint for host %s in environment.json", hostWithPort)
	}

	if !containsAny(bodyStr, []string{
		fmt.Sprintf(`"issuer":"http://%s"`, hostWithPort),
		fmt.Sprintf(`"issuer":"https://%s"`, hostWithPort),
	}) {
		return fmt.Errorf("couldn't find issuer endpoint for host %s in environment.json", hostWithPort)
	}

	return nil
}

func checkOIDCDiscovery(ctx context.Context, apiBaseURL string) error {
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statusCode, body, err := httphelper.Get(checkCtx, apiBaseURL+"/.well-known/openid-configuration", nil)
	if err != nil {
		return err
	}
	if statusCode != 200 {
		return fmt.Errorf("expected status 200 but got %d", statusCode)
	}

	bodyStr := string(body)
	parsedURL, _ := url.Parse(apiBaseURL)
	hostWithPort := parsedURL.Host

	if !containsAny(bodyStr, []string{
		fmt.Sprintf(`"issuer":"http://%s"`, hostWithPort),
		fmt.Sprintf(`"issuer":"https://%s"`, hostWithPort),
	}) {
		return fmt.Errorf("couldn't find issuer for host %s in OIDC discovery", hostWithPort)
	}

	return nil
}

func checkGRPCHealth(ctx context.Context, apiBaseURL string) error {
	conn, err := grpchelper.Dial(ctx, apiBaseURL)
	if err != nil {
		return fmt.Errorf("couldn't create gRPC connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		// If health check is unimplemented, gRPC connectivity is still verified
		if status.Code(err) == codes.Unimplemented {
			return nil
		}
		return err
	}
	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("gRPC health check returned non-serving status: %v", resp.Status)
	}
	return nil
}

func containsAny(body string, targets []string) bool {
	for _, target := range targets {
		if strings.Contains(body, target) {
			return true
		}
	}
	return false
}

func awaitCheck(ctx context.Context, t *testing.T, waitFor time.Duration, cb func(ctx context.Context) error, msg string, args ...any) {
	t.Helper()
	require.EventuallyWithTf(t, func(collect *assert.CollectT) {
		if !assert.NoError(collect, cb(ctx)) {
			t.Logf("retrying in a second")
		}
	}, waitFor, time.Second, msg, args...)
}
