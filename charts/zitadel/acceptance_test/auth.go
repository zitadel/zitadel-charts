package acceptance_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zitadel/oidc/pkg/oidc"
	"google.golang.org/protobuf/types/known/emptypb"

	grpchelper "github.com/zitadel/zitadel-charts/charts/zitadel/acceptance_test/helpers/grpc"
	httphelper "github.com/zitadel/zitadel-charts/charts/zitadel/acceptance_test/helpers/http"
)

// CheckAuthenticatedAPI verifies that both HTTP and gRPC authenticated API
// endpoints are functioning correctly using a machine user's service account
// credentials. This check validates the complete machine-to-machine auth flow.
//
// The function retrieves the service account key from the specified Kubernetes
// secret, generates a JWT assertion, exchanges it for an access token via the
// OAuth token endpoint, and then makes authenticated calls to both the HTTP
// management API (/management/v1/languages) and gRPC management API.
//
// This check validates:
//   - Service account key provisioning in Kubernetes secrets
//   - JWT profile assertion generation and signing
//   - OAuth token endpoint functionality
//   - Bearer token authentication on HTTP and gRPC endpoints
//   - Management API availability and authorization
//
// The check uses eventual consistency with a 1-minute timeout to account for
// potential startup delays in token issuance and API readiness.
func CheckAuthenticatedAPI(ctx context.Context, t *testing.T, k *k8s.KubectlOptions, apiBaseURL, secretName, secretKey string) {
	t.Helper()

	secret := k8s.GetSecret(t, k, secretName)
	key := secret.Data[secretKey]
	require.NotNil(t, key, "key %s in secret %s is nil", secretKey, secretName)

	jwta, err := oidc.NewJWTProfileAssertionFromFileData(key, []string{apiBaseURL})
	require.NoError(t, err)

	jwt, err := oidc.GenerateJWTProfileToken(jwta)
	require.NoError(t, err)

	authCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	var token string
	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		var tokenErr error
		token, tokenErr = getAccessToken(authCtx, jwt, apiBaseURL)
		assert.NoError(collect, tokenErr)
	}, 1*time.Minute, time.Second, "getting token failed for a minute")

	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		httpErr := callAuthenticatedHTTP(authCtx, token, apiBaseURL)
		if !assert.NoError(collect, httpErr) {
			return
		}
		grpcErr := callAuthenticatedGRPC(authCtx, token, apiBaseURL)
		assert.NoError(collect, grpcErr)
	}, 1*time.Minute, time.Second, "calling authenticated endpoints failed for a minute")
}

func getAccessToken(ctx context.Context, jwt, apiBaseURL string) (string, error) {
	form := url.Values{}
	form.Add("grant_type", string(oidc.GrantTypeBearer))
	form.Add("scope", fmt.Sprintf("%s %s %s urn:zitadel:iam:org:project:id:zitadel:aud", oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeEmail))
	form.Add("assertion", jwt)

	status, body, err := httphelper.Post(ctx, fmt.Sprintf("%s/oauth/v2/token", apiBaseURL),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	if status != 200 {
		return "", fmt.Errorf("expected token response 200, but got %d", status)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err = json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}
	return tokenResp.AccessToken, nil
}

func callAuthenticatedHTTP(ctx context.Context, token, apiBaseURL string) error {
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statusCode, _, err := httphelper.Get(checkCtx, fmt.Sprintf("%s/management/v1/languages", apiBaseURL),
		map[string]string{"Authorization": fmt.Sprintf("Bearer %s", token)})
	if err != nil {
		return err
	}
	if statusCode != 200 {
		return fmt.Errorf("expected status 200 at authenticated endpoint, but got %d", statusCode)
	}
	return nil
}

func callAuthenticatedGRPC(ctx context.Context, token, apiBaseURL string) error {
	conn, err := grpchelper.Dial(ctx, apiBaseURL)
	if err != nil {
		return fmt.Errorf("couldn't create gRPC connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	authCtx := grpchelper.WithBearerToken(ctx, token)
	var resp emptypb.Empty
	err = grpchelper.Invoke(authCtx, conn, "/zitadel.management.v1.ManagementService/GetSupportedLanguages", &emptypb.Empty{}, &resp)
	if err != nil {
		return fmt.Errorf("authenticated gRPC call failed: %w", err)
	}
	return nil
}

