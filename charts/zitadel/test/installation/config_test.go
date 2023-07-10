package installation_test

import (
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/suite"
	"github.com/zitadel/oidc/pkg/oidc"
	"github.com/zitadel/zitadel-charts/charts/zitadel/test/installation"
)

func TestPostgresPwAuth(t *testing.T) {
	t.Parallel()
	example := "2-postgres-pw-auth"
	_, values := workingDirectory(example)
	suite.Run(t, installation.Configure(
		t,
		newNamespaceIdentifier(example),
		installation.WithValues(installation.Postgres, map[string]string{
			"auth.postgresPassword": "abc",
		}),
		[]string{values},
		nil,
		nil,
	))
}

func TestPostgresCertAuth(t *testing.T) {
	t.Parallel()
	example := "3-postgres-cert-auth"
	_, values := workingDirectory(example)
	suite.Run(t, installation.Configure(
		t,
		newNamespaceIdentifier(example),
		installation.WithValues(installation.Postgres, map[string]string{
			"auth.postgresPassword": "abc",
		}),
		[]string{values},
		nil,
		nil,
	))
}

func TestCockroachPwAuth(t *testing.T) {
	t.Parallel()
	example := "5-cockroach-pw-auth"
	_, values := workingDirectory(example)
	suite.Run(t, installation.Configure(
		t,
		newNamespaceIdentifier(example),
		installation.Cockroach,
		[]string{values},
		nil,
		nil,
	))
}

func TestCockroachCertAuth(t *testing.T) {
	t.Parallel()
	example := "6-cockroach-cert-auth"
	_, values := workingDirectory(example)
	suite.Run(t, installation.Configure(
		t,
		newNamespaceIdentifier(example),
		installation.Cockroach,
		[]string{values},
		func(cfg *installation.ConfigurationTest) {
			k8s.RunKubectl()
		},
		nil,
	))
}

func TestReferencedSecrets(t *testing.T) {
	t.Parallel()
	example := "7-referenced-secrets"
	workDir, values := workingDirectory(example)
	suite.Run(t, installation.Configure(
		t,
		newNamespaceIdentifier(example),
		installation.Cockroach,
		[]string{values},
		func(cfg *installation.ConfigurationTest) {
			k8s.KubectlApply(t, cfg.KubeOptions, filepath.Join(workDir, "zitadel.yaml"))
			k8s.KubectlApply(t, cfg.KubeOptions, filepath.Join(workDir, "masterkey.yaml"))
		},
		nil,
	))
}

func TestMachineUser(t *testing.T) {
	t.Parallel()
	example := "8-machine-user"
	_, values := workingDirectory(example)
	saUserame := readValues(t, values).Zitadel.ConfigmapConfig.FirstInstance.Org.Machine.Machine.Username
	suite.Run(t, installation.Configure(
		t,
		newNamespaceIdentifier(example),
		installation.Cockroach,
		[]string{values},
		nil,
		testJWTProfileKey("http://localhost:8080", saUserame, fmt.Sprintf("%s.json", saUserame))),
	)
}

func readValues(t *testing.T, valuesFilePath string) (values struct {
	Zitadel struct {
		MasterkeySecretName string `yaml:"masterkeySecretName"`
		ConfigSecretName    string `yaml:"configSecretName"`
		ConfigmapConfig     struct {
			FirstInstance struct {
				Org struct {
					Machine struct {
						Machine struct {
							Username string `yaml:"Username"`
						} `yaml:"Machine"`
					} `yaml:"Machine"`
				} `yaml:"Org"`
			} `yaml:"FirstInstance"`
		} `yaml:"configmapConfig"`
	} `yaml:"zitadel"`
}) {
	valuesBytes, err := os.ReadFile(valuesFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(valuesBytes, &values); err != nil {
		t.Fatal(err)
	}
	return values
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

func workingDirectory(exampleDir string) (workingDir, valuesFile string) {
	_, filename, _, _ := runtime.Caller(0)
	workingDir = filepath.Join(filename, "..", "..", "..", "..", "..", "examples", exampleDir)
	valuesFile = filepath.Join(workingDir, "values.yaml")
	return workingDir, valuesFile
}
