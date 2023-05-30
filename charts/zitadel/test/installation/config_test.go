//go:build integration
// +build integration

package installation_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/suite"
	"github.com/zitadel/oidc/pkg/oidc"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/zitadel/zitadel-charts/charts/zitadel/test/installation"
)

func TestWithInlineSecrets(t *testing.T) {
	suite.Run(t, installation.Configure(t, randomNewNamespace(), map[string]string{
		"zitadel.configmapConfig.ExternalSecure":                "false",
		"zitadel.configmapConfig.TLS.Enabled":                   "false",
		"zitadel.secretConfig.Database.cockroach.User.Password": "xy",
		"pdb.enabled":       "true",
		"ingress.enabled":   "true",
		"zitadel.masterkey": "x123456789012345678901234567891y",
	}, nil, nil))
}

func TestWithReferencedSecrets(t *testing.T) {
	masterKeySecretName := "existing-zitadel-masterkey"
	masterKeySecretKey := "masterkey"
	zitadelConfigSecretName := "existing-zitadel-secrets"
	zitadelConfigSecretKey := "config-yaml"
	suite.Run(t, installation.Configure(t, randomNewNamespace(), map[string]string{
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
	saUserame := "zitadel-admin-sa"
	suite.Run(t, installation.Configure(t, randomNewNamespace(), map[string]string{
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
		form := url.Values{}
		form.Add("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
		form.Add("scope", "openid profile email urn:zitadel:iam:org:project:id:zitadel:aud")
		form.Add("assertion", jwt)
		tokenResp, tokenBody, err := installation.HttpPost(cfg.Ctx, fmt.Sprintf("%s/oauth/v2/token", audience), func(req *http.Request) {
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		}, strings.NewReader(form.Encode()))
		token := struct {
			AccessToken string `json:"access_token"`
		}{}
		if tokenResp.StatusCode != 200 {
			t.Fatalf("expected token response 200, but got %d", tokenResp.StatusCode)
		}
		if err = json.Unmarshal(tokenBody, &token); err != nil {
			t.Fatal(err)
		}
		langResp, _, err := installation.HttpGet(cfg.Ctx, fmt.Sprintf("%s/management/v1/languages", audience), func(req *http.Request) {
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
		})
		if err != nil {
			t.Fatal(err)
		}
		if langResp.StatusCode != 200 {
			t.Fatalf("Expected status 200 at an authenticated endpoint, but got %d", langResp.StatusCode)
		}
	}
}

func randomNewNamespace() string {
	// if triggered by a github action the environment variable is set
	// we use it to better identify the test
	commitSHA, exist := os.LookupEnv("GITHUB_SHA")
	namespace := "zitadel-helm-" + strings.ToLower(random.UniqueId())
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
