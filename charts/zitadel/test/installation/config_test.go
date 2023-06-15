//go:build integration
// +build integration

package installation_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/suite"
	"github.com/zitadel/oidc/pkg/oidc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/zitadel/zitadel-charts/charts/zitadel/test/installation"
)

func TestWithInlineSecrets(t *testing.T) {
	t.Parallel()
	suite.Run(t, installation.Configure(t, newNamespaceIdentifier("1-inline-secrets"), map[string]string{
		"zitadel.configmapConfig.ExternalSecure":                "false",
		"zitadel.configmapConfig.TLS.Enabled":                   "false",
		"zitadel.secretConfig.Database.cockroach.User.Password": "xy",
		"pdb.enabled":       "true",
		"ingress.enabled":   "true",
		"zitadel.masterkey": "x123456789012345678901234567891y",
	}, nil, nil))
}

func TestWithReferencedSecrets(t *testing.T) {
	t.Parallel()
	masterKeySecretName := "existing-zitadel-masterkey"
	masterKeySecretKey := "masterkey"
	zitadelConfigSecretName := "existing-zitadel-secrets"
	zitadelConfigSecretKey := "config-yaml"
	suite.Run(t, installation.Configure(t, newNamespaceIdentifier("2-ref-secrets"), map[string]string{
		"zitadel.configmapConfig.ExternalSecure":                "false",
		"zitadel.configmapConfig.TLS.Enabled":                   "false",
		"zitadel.secretConfig.Database.cockroach.User.Password": "xy",
		"pdb.enabled":                 "true",
		"ingress.enabled":             "true",
		"zitadel.masterkeySecretName": masterKeySecretName,
		"zitadel.configSecretName":    zitadelConfigSecretName,
	}, func(cfg *installation.ConfigurationTest) {
		if err := createSecret(cfg.Ctx, cfg.KubeOptions.Namespace, cfg.KubeClient, masterKeySecretName, masterKeySecretKey, "x123456789012345678901234567891y"); err != nil {
			t.Fatal(err)
		}
		if err := createSecret(cfg.Ctx, cfg.KubeOptions.Namespace, cfg.KubeClient, zitadelConfigSecretName, zitadelConfigSecretKey, "ExternalSecure: false\n"); err != nil {
			t.Fatal(err)
		}
	}, nil))
}

func TestWithMachineKey(t *testing.T) {
	t.Parallel()
	saUserame := "zitadel-admin-sa"
	suite.Run(t, installation.Configure(t, newNamespaceIdentifier("3-machine-key"), map[string]string{
		"zitadel.configmapConfig.ExternalSecure":                "false",
		"zitadel.configmapConfig.TLS.Enabled":                   "false",
		"zitadel.secretConfig.Database.cockroach.User.Password": "xy",
		"pdb.enabled":       "true",
		"ingress.enabled":   "true",
		"zitadel.masterkey": "x123456789012345678901234567891y",
		"zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Username": saUserame,
		"zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Name":     "Admin",
		"zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.Type":  "1",
	}, nil, testJWTProfileKey("http://localhost:8080", saUserame, fmt.Sprintf("%s.json", saUserame))))
}

func createSecret(ctx context.Context, namespace string, k8sClient *kubernetes.Clientset, name, key, value string) error {
	_, err := k8sClient.CoreV1().Secrets(namespace).Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		StringData: map[string]string{key: value},
	}, metav1.CreateOptions{})
	return err
}

func testJWTProfileKey(audience, secretName, secretKey string) func(test *installation.ConfigurationTest) {
	return func(cfg *installation.ConfigurationTest) {
		t := cfg.T()
		secret := k8s.GetSecret(t, cfg.KubeOptions, secretName)
		key := secret.Data[secretKey]
		if key == nil {
			t.Fatalf("key %s in secret %s is nil", secretKey, secretName)
		}
		jwta, err := oidc.NewJWTProfileAssertionFromFileData(key, []string{audience})
		if err != nil {
			t.Fatal(err)
		}
		jwt, err := oidc.GenerateJWTProfileToken(jwta)
		if err != nil {
			t.Fatal(err)
		}
		closeTunnel := installation.ServiceTunnel(cfg)
		defer closeTunnel()
		var token string
		installation.Await(cfg.Ctx, t, nil, 60, func(ctx context.Context) error {
			var tokenErr error
			token, tokenErr = getToken(ctx, t, audience, jwt)
			return tokenErr
		})
		installation.Await(cfg.Ctx, t, nil, 60, func(ctx context.Context) error {
			return callAuthenticatedEndpoint(ctx, audience, token)
		})
	}
}

func getToken(ctx context.Context, t *testing.T, audience, jwt string) (string, error) {
	form := url.Values{}
	form.Add("grant_type", string(oidc.GrantTypeBearer))
	form.Add("scope", fmt.Sprintf("%s %s %s urn:zitadel:iam:org:project:id:zitadel:aud", oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeEmail))
	form.Add("assertion", jwt)
	//nolint:bodyclose
	resp, tokenBody, err := installation.HttpPost(ctx, fmt.Sprintf("%s/oauth/v2/token", audience), func(req *http.Request) {
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

func callAuthenticatedEndpoint(ctx context.Context, audience, token string) error {
	checkCtx, checkCancel := context.WithTimeout(ctx, 5*time.Second)
	defer checkCancel()
	//nolint:bodyclose
	resp, _, err := installation.HttpGet(checkCtx, fmt.Sprintf("%s/management/v1/languages", audience), func(req *http.Request) {
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

func newNamespaceIdentifier(testcase string) string {
	// if triggered by a github action the environment variable is set
	// we use it to better identify the test
	commitSHA, exist := os.LookupEnv("GITHUB_SHA")
	namespace := fmt.Sprintf("zitadel-test-%s-%s", testcase, strings.ToLower(random.UniqueId()))
	if exist {
		namespace += "-" + commitSHA
	}
	// max namespace length is 63 characters
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
	return truncateString(namespace, 63)
}

func truncateString(str string, num int) string {
	shortenStr := str
	if len(str) > num {
		shortenStr = str[0:num]
	}
	return shortenStr
}
