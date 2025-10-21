package acceptance_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/zitadel/oidc/pkg/oidc"
	mgmtapi "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/management"
)

func testAuthenticatedAPI(secretName, secretKey string) func(test *ConfigurationTest) {
	return func(cfg *ConfigurationTest) {
		t := cfg.T()
		secret := k8s.GetSecret(t, cfg.KubeOptions, secretName)
		key := secret.Data[secretKey]
		if key == nil {
			t.Fatalf("key %s in secret %s is nil", secretKey, secretName)
		}
		jwta, err := oidc.NewJWTProfileAssertionFromFileData(key, []string{cfg.ApiBaseUrl})
		if err != nil {
			t.Fatal(err)
		}
		jwt, err := oidc.GenerateJWTProfileToken(jwta)
		if err != nil {
			t.Fatal(err)
		}
		awaitCtx, cancel := context.WithTimeoutCause(context.Background(), 5*time.Minute, fmt.Errorf("testing authenticated APIs timed out after 5 minutes"))
		defer cancel()
		var token string
		Awaitf(awaitCtx, t, 1*time.Minute, func(ctx context.Context) error {
			var tokenErr error
			token, tokenErr = getToken(ctx, t, jwt, cfg.ApiBaseUrl)
			return tokenErr
		}, "getting token failed for a minute")
		Awaitf(awaitCtx, t, 1*time.Minute, func(ctx context.Context) error {
			if httpErr := callAuthenticatedHTTPEndpoint(ctx, token, cfg.ApiBaseUrl); httpErr != nil {
				return httpErr
			}
			return callAuthenticatedGRPCEndpoint(ctx, cfg, key)
		}, "calling authenticated endpoints failed for a minute")
	}
}

func getToken(ctx context.Context, t *testing.T, jwt, apiBaseURL string) (string, error) {
	form := url.Values{}
	form.Add("grant_type", string(oidc.GrantTypeBearer))
	form.Add("scope", fmt.Sprintf("%s %s %s urn:zitadel:iam:org:project:id:zitadel:aud", oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeEmail))
	form.Add("assertion", jwt)
	//nolint:bodyclose
	resp, tokenBody, err := HttpPost(ctx, fmt.Sprintf("%s/oauth/v2/token", apiBaseURL), func(req *http.Request) {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("expected token response 200, but got %d", resp.StatusCode)
	}
	token := struct {
		AccessToken string `json:"access_token"`
	}{}
	if err = json.Unmarshal(tokenBody, &token); err != nil {
		t.Fatal(err)
	}
	return token.AccessToken, nil
}

func callAuthenticatedHTTPEndpoint(ctx context.Context, token, apiBaseURL string) error {
	checkCtx, checkCancel := context.WithTimeoutCause(ctx, 5*time.Second, fmt.Errorf("calling authenticated HTTP endpoint timed out after 5 seconds"))
	defer checkCancel()
	//nolint:bodyclose
	resp, _, err := HttpGet(checkCtx, fmt.Sprintf("%s/management/v1/languages", apiBaseURL), func(req *http.Request) {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	})
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("expected status 200 at an authenticated endpoint, but got %d", resp.StatusCode)
	}
	return nil
}

func callAuthenticatedGRPCEndpoint(ctx context.Context, cfg *ConfigurationTest, key []byte) error {
	t := cfg.T()
	cleanup := withInsecureDefaultHttpClient()
	defer cleanup()
	conn, err := OpenGRPCConnection(ctx, cfg, key)
	if err != nil {
		return fmt.Errorf("couldn't open gRPC connection: %v", err)
	}
	_, err = conn.GetSupportedLanguages(ctx, &mgmtapi.GetSupportedLanguagesRequest{})
	if err != nil {
		t.Fatalf("couldn't call authenticated gRPC endpoint: %v", err)
	}
	return nil
}
