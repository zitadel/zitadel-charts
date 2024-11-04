package acceptance_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/zitadel/oidc/pkg/oidc"
	"github.com/zitadel/zitadel-charts/charts/zitadel/acceptance"
	mgmt_api "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/management"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func testAuthenticatedAPI(secretName, secretKey string) func(test *acceptance.ConfigurationTest) {
	return func(cfg *acceptance.ConfigurationTest) {
		t := cfg.T()
		apiBaseURL := cfg.APIBaseURL()
		secret := k8s.GetSecret(t, cfg.KubeOptions, secretName)
		key := secret.Data[secretKey]
		if key == nil {
			t.Fatalf("key %s in secret %s is nil", secretKey, secretName)
		}
		jwta, err := oidc.NewJWTProfileAssertionFromFileData(key, []string{apiBaseURL})
		if err != nil {
			t.Fatal(err)
		}
		jwt, err := oidc.GenerateJWTProfileToken(jwta)
		if err != nil {
			t.Fatal(err)
		}
		closeTunnel := acceptance.ServiceTunnel(cfg)
		defer closeTunnel()
		var token string
		acceptance.Await(cfg.Ctx, t, nil, 60, func(ctx context.Context) error {
			var tokenErr error
			token, tokenErr = getToken(ctx, t, jwt, apiBaseURL)
			return tokenErr
		})
		acceptance.Await(cfg.Ctx, t, nil, 60, func(ctx context.Context) error {
			if httpErr := callAuthenticatedHTTPEndpoint(ctx, token, apiBaseURL); httpErr != nil {
				return httpErr
			}
			return callAuthenticatedGRPCEndpoint(cfg, key)
		})
	}
}

func getToken(ctx context.Context, t *testing.T, jwt, apiBaseURL string) (string, error) {
	form := url.Values{}
	form.Add("grant_type", string(oidc.GrantTypeBearer))
	form.Add("scope", fmt.Sprintf("%s %s %s urn:zitadel:iam:org:project:id:zitadel:aud", oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeEmail))
	form.Add("assertion", jwt)
	//nolint:bodyclose
	resp, tokenBody, err := acceptance.HttpPost(ctx, fmt.Sprintf("%s/oauth/v2/token", apiBaseURL), func(req *http.Request) {
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
	checkCtx, checkCancel := context.WithTimeout(ctx, 5*time.Second)
	defer checkCancel()
	//nolint:bodyclose
	resp, _, err := acceptance.HttpGet(checkCtx, fmt.Sprintf("%s/management/v1/languages", apiBaseURL), func(req *http.Request) {
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

func callAuthenticatedGRPCEndpoint(cfg *acceptance.ConfigurationTest, key []byte) error {
	t := cfg.T()
	conn, err := acceptance.OpenGRPCConnection(cfg, key)
	if err != nil {
		return fmt.Errorf("couldn't open gRPC connection: %v", err)
	}
	_, err = conn.GetSupportedLanguages(cfg.Ctx, &mgmt_api.GetSupportedLanguagesRequest{})
	if err != nil {
		t.Fatalf("couldn't call authenticated gRPC endpoint: %v", err)
	}
	return nil
}
