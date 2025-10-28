package acceptance_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	oidcclient "github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/management"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// assertGRPCWorks validates that authenticated GRPC connectivity is working
// by retrieving the JWT key from the specified Kubernetes secret and calling
// an authenticated endpoint. The secret is expected to contain a key named
// "{secretName}.json" with the JWT key file data.
func assertGRPCWorks(ctx context.Context, t *testing.T, cfg *IntegrationSuite, secretName string) {
	t.Helper()

	secret := k8s.GetSecret(t, cfg.KubeOptions, secretName)
	secretKey := fmt.Sprintf("%s.json", secretName)
	key := secret.Data[secretKey]
	require.NotNil(t, key, "key %s in secret %s is nil", secretKey, secretName)

	cleanup := withInsecureDefaultHttpClient()
	defer cleanup()

	conn, err := openGRPCConnection(ctx, cfg, key)
	require.NoError(t, err, "failed to create gRPC management client")

	_, err = conn.GetSupportedLanguages(ctx, &management.GetSupportedLanguagesRequest{})
	require.NoError(t, err, "authenticated gRPC call failed")

	t.Log("✓ Authenticated GRPC connectivity verified")
}

// openGRPCConnection creates a GRPC connection to the ZITADEL instance. If
// key is nil, the connection is unauthenticated. Otherwise, the key should
// contain JWT key file data for authenticated connections.
func openGRPCConnection(ctx context.Context, cfg *IntegrationSuite, key []byte) (management.ManagementServiceClient, error) {
	clientOptions := []client.Option{
		client.WithGRPCDialOptions(grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}))),
	}

	if key != nil {
		keyFile, err := oidcclient.ConfigFromKeyFileData(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key file: %w", err)
		}
		clientOptions = append(clientOptions, client.WithAuth(client.JWTAuthentication(keyFile, client.ScopeZitadelAPI())))
	}

	apiBaseUrl, err := url.Parse(cfg.ApiBaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API base URL: %w", err)
	}

	c, err := client.New(ctx, zitadel.New(apiBaseUrl.Hostname(), zitadel.WithInsecureSkipVerifyTLS()), clientOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZITADEL client: %w", err)
	}

	return c.ManagementService(), nil
}

// withInsecureDefaultHttpClient configures the default HTTP client to skip
// TLS verification. This is needed for OIDC discovery in test environments
// with self-signed certificates. Returns a cleanup function to restore the
// original transport.
func withInsecureDefaultHttpClient() func() {
	http.DefaultClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	return func() {
		http.DefaultClient.Transport = nil
	}
}
