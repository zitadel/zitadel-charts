package smoke_test

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"

	"github.com/zitadel/zitadel-charts/charts/zitadel/smoke_test/support"
)

func TestTLSMatrix(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)

	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	commonSetValues := map[string]string{
		"zitadel.masterkey":                                         "x123456789012345678901234567891y",
		"zitadel.configmapConfig.ExternalDomain":                    "zitadel.dev.mrida.ng",
		"zitadel.configmapConfig.ExternalPort":                      "443",
		"zitadel.configmapConfig.ExternalSecure":                    "true",
		"zitadel.configmapConfig.Database.Postgres.Host":            "db-postgresql",
		"zitadel.configmapConfig.Database.Postgres.Port":            "5432",
		"zitadel.configmapConfig.Database.Postgres.Database":        "zitadel",
		"zitadel.configmapConfig.Database.Postgres.MaxOpenConns":    "20",
		"zitadel.configmapConfig.Database.Postgres.MaxIdleConns":    "10",
		"zitadel.configmapConfig.Database.Postgres.MaxConnLifetime": "30m",
		"zitadel.configmapConfig.Database.Postgres.MaxConnIdleTime": "5m",
		"zitadel.configmapConfig.Database.Postgres.User.Username":   "postgres",
		"zitadel.configmapConfig.Database.Postgres.User.SSL.Mode":   "disable",
		"zitadel.configmapConfig.Database.Postgres.Admin.Username":  "postgres",
		"zitadel.configmapConfig.Database.Postgres.Admin.SSL.Mode":  "disable",
		"ingress.enabled":       "false",
		"login.ingress.enabled": "false",
		"service.appProtocol":   "kubernetes.io/https",
		"service.port":          "443",
	}

	testCases := []struct {
		name          string
		setupTLS      func(t *testing.T, env *support.Env)
		setValues     map[string]string
		setJsonValues map[string]string
	}{
		{
			name: "user-provided-certificate",
			setupTLS: func(t *testing.T, env *support.Env) {
				support.WithTLSSecret(t, env, "wildcard-tls", "*.dev.mrida.ng", "dev.mrida.ng")
			},
			setValues: map[string]string{
				"zitadel.serverSslCrtSecret": "wildcard-tls",
				"login.serverSslCrtSecret":   "wildcard-tls",
			},
			setJsonValues: map[string]string{
				"login.env": `[{"name":"NODE_TLS_REJECT_UNAUTHORIZED","value":"0"}]`,
			},
		},
		{
			name: "chart-generated-self-signed",
			setupTLS: func(t *testing.T, env *support.Env) {
			},
			setValues: map[string]string{
				"zitadel.selfSignedCert.enabled":           "true",
				"zitadel.selfSignedCert.additionalDnsName": "zitadel.dev.mrida.ng",
				"login.selfSignedCert.enabled":             "true",
				"login.selfSignedCert.additionalDnsName":   "zitadel.dev.mrida.ng",
			},
			setJsonValues: map[string]string{
				"login.env": `[{"name":"NODE_TLS_REJECT_UNAUTHORIZED","value":"0"}]`,
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			support.WithNamespace(t, cluster, func(env *support.Env) {
				env.Logger.Logf(t, "namespace %q created; installing PostgreSQL…", env.Namespace)
				support.WithPostgres(t, env)

				releaseName := env.MakeRelease("zitadel-tls", testCase.name)

				testCase.setupTLS(t, env)

				mergedSetValues := make(map[string]string)
				for key, value := range commonSetValues {
					mergedSetValues[key] = value
				}
				for key, value := range testCase.setValues {
					mergedSetValues[key] = value
				}

				helmOptions := &helm.Options{
					KubectlOptions: env.Kube,
					SetValues:      mergedSetValues,
					SetJsonValues:  testCase.setJsonValues,
					ExtraArgs: map[string][]string{
						"upgrade": {"--install", "--wait", "--timeout", "30m"},
					},
				}

				require.NoError(t, helm.UpgradeE(t, helmOptions, chartPath, releaseName))

				waitForDeploymentReady(t, env, releaseName)
				waitForDeploymentReady(t, env, releaseName+"-login")

				assertZitadelConsoleAccessible(t, env, releaseName, 443)
				assertLoginUIAccessible(t, env, releaseName+"-login", 443)
			})
		})
	}
}

func waitForDeploymentReady(t *testing.T, env *support.Env, deploymentName string) {
	t.Helper()

	require.Eventually(t, func() bool {
		deployment, err := k8s.GetDeploymentE(t, env.Kube, deploymentName)
		if err != nil {
			return false
		}
		return deployment.Status.ReadyReplicas > 0 &&
			deployment.Status.ReadyReplicas == deployment.Status.Replicas
	}, 5*time.Minute, 5*time.Second)
}

func assertZitadelConsoleAccessible(t *testing.T, env *support.Env, serviceName string, port int) {
	t.Helper()

	tunnel := k8s.NewTunnel(env.Kube, k8s.ResourceTypeService, serviceName, 0, port)
	defer tunnel.Close()
	tunnel.ForwardPort(t)

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	require.Eventually(t, func() bool {
		resp, err := client.Get(fmt.Sprintf("https://%s/ui/console", tunnel.Endpoint()))
		if err != nil {
			env.Logger.Logf(t, "console access failed for %s: %v", serviceName, err)
			return false
		}
		defer resp.Body.Close()

		env.Logger.Logf(t, "---- %s console response ----", serviceName)
		env.Logger.Logf(t, "Status: %s", resp.Status)
		env.Logger.Logf(t, "Headers:")
		for key, values := range resp.Header {
			for _, value := range values {
				env.Logger.Logf(t, "  %s: %s", key, value)
			}
		}

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if len(bodyStr) > 1000 {
			bodyStr = bodyStr[:1000] + "... (truncated)"
		}
		env.Logger.Logf(t, "Body: %s", bodyStr)
		env.Logger.Logf(t, "---- end console response ----")

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
			if resp.StatusCode >= 300 && resp.StatusCode < 400 {
				location := resp.Header.Get("Location")
				env.Logger.Logf(t, "Console redirects to: %s", location)
				return strings.Contains(location, "/ui/login") || strings.Contains(location, "/oauth")
			}
			return true
		}

		return false
	}, 2*time.Minute, 5*time.Second)
}

func assertLoginUIAccessible(t *testing.T, env *support.Env, serviceName string, port int) {
	t.Helper()

	tunnel := k8s.NewTunnel(env.Kube, k8s.ResourceTypeService, serviceName, 0, port)
	defer tunnel.Close()
	tunnel.ForwardPort(t)

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	require.Eventually(t, func() bool {
		resp, err := client.Get(fmt.Sprintf("https://%s/ui/v2/login", tunnel.Endpoint()))
		if err != nil {
			env.Logger.Logf(t, "login UI access failed for %s: %v", serviceName, err)
			return false
		}
		defer resp.Body.Close()

		env.Logger.Logf(t, "---- %s login UI response ----", serviceName)
		env.Logger.Logf(t, "Status: %s", resp.Status)
		env.Logger.Logf(t, "Headers:")
		for key, values := range resp.Header {
			for _, value := range values {
				env.Logger.Logf(t, "  %s: %s", key, value)
			}
		}

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if len(bodyStr) > 1000 {
			bodyStr = bodyStr[:1000] + "... (truncated)"
		}
		env.Logger.Logf(t, "Body: %s", bodyStr)
		env.Logger.Logf(t, "---- end login UI response ----")

		return resp.StatusCode == http.StatusOK
	}, 2*time.Minute, 5*time.Second)
}
