package smoke_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zitadel/zitadel-charts/test/smoke/support"
)

type secretExpected struct {
	machineKeySecret   bool
	machinePatSecret   bool
	loginClientSecret  bool
	machineKeyName     string
	machinePatName     string
	loginClientName    string
	machineKeyContent  string
	machinePatContent  string
	loginClientContent string
}

// TestSecretsMatrix validates secret creation behavior across various configuration
// combinations. It verifies that the setup job creates the expected Kubernetes
// secrets for machine keys, PATs, and login client authentication tokens.
func TestSecretsMatrix(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)

	chartPath := support.ChartPath(t)

	testCases := []struct {
		name      string
		setValues map[string]string
		expected  secretExpected
	}{
		{
			name: "default-all-enabled",
			setValues: map[string]string{
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Username":          "iam-admin",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Name":              "Admin Machine",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.ExpirationDate": "2029-01-01T00:00:00Z",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.Type":           "1",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Pat.ExpirationDate":        "2029-01-01T00:00:00Z",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Username":      "login-client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Name":          "Login Client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Pat.ExpirationDate":    "2029-01-01T00:00:00Z",
			},
			expected: secretExpected{
				machineKeySecret:   true,
				machinePatSecret:   true,
				loginClientSecret:  true,
				machineKeyName:     "iam-admin",
				machinePatName:     "iam-admin-pat",
				loginClientName:    "login-client",
				machineKeyContent:  "iam-admin.json",
				machinePatContent:  "pat",
				loginClientContent: "pat",
			},
		},
		{
			name: "machine-only-no-pat",
			setValues: map[string]string{
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Username":          "my-machine",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Name":              "My Machine",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.ExpirationDate": "2029-01-01T00:00:00Z",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.Type":           "1",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Username":      "login-client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Name":          "Login Client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Pat.ExpirationDate":    "2029-01-01T00:00:00Z",
			},
			expected: secretExpected{
				machineKeySecret:   true,
				machinePatSecret:   false,
				loginClientSecret:  true,
				machineKeyName:     "my-machine",
				machinePatName:     "",
				loginClientName:    "login-client",
				machineKeyContent:  "my-machine.json",
				machinePatContent:  "",
				loginClientContent: "pat",
			},
		},
		{
			name: "login-client-only",
			setValues: map[string]string{
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Username":   "login-client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Name":       "Login Client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Pat.ExpirationDate": "2029-01-01T00:00:00Z",
			},
			expected: secretExpected{
				machineKeySecret:   false,
				machinePatSecret:   false,
				loginClientSecret:  true,
				machineKeyName:     "",
				machinePatName:     "",
				loginClientName:    "login-client",
				machineKeyContent:  "",
				machinePatContent:  "",
				loginClientContent: "pat",
			},
		},
		{
			name: "custom-names-with-prefix",
			setValues: map[string]string{
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Username":          "custom-admin",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Name":              "Custom Admin",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.ExpirationDate": "2029-01-01T00:00:00Z",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.Type":           "1",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Pat.ExpirationDate":        "2029-01-01T00:00:00Z",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Username":      "login-client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Name":          "Login Client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Pat.ExpirationDate":    "2029-01-01T00:00:00Z",
				"login.loginClientSecretPrefix":                                               "myapp-",
			},
			expected: secretExpected{
				machineKeySecret:   true,
				machinePatSecret:   true,
				loginClientSecret:  true,
				machineKeyName:     "custom-admin",
				machinePatName:     "custom-admin-pat",
				loginClientName:    "myapp-login-client",
				machineKeyContent:  "custom-admin.json",
				machinePatContent:  "pat",
				loginClientContent: "pat",
			},
		},
		{
			name:      "minimal-no-setup",
			setValues: map[string]string{},
			expected: secretExpected{
				machineKeySecret:   false,
				machinePatSecret:   false,
				loginClientSecret:  false,
				machineKeyName:     "",
				machinePatName:     "",
				loginClientName:    "",
				machineKeyContent:  "",
				machinePatContent:  "",
				loginClientContent: "",
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

				uniqueDomain := fmt.Sprintf("%s.test.local", env.Namespace)
				commonSetValues := map[string]string{
					"zitadel.masterkey":                                         "x123456789012345678901234567891y",
					"zitadel.configmapConfig.ExternalDomain":                    uniqueDomain,
					"zitadel.configmapConfig.ExternalPort":                      "443",
					"zitadel.configmapConfig.TLS.Enabled":                       "false",
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
					"image.tag": support.DigestTag,
				}

				releaseName := env.MakeRelease("zitadel-secrets-test", testCase.name)

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
					ExtraArgs: map[string][]string{
						"upgrade": {"--install", "--wait", "--timeout", "30m"},
					},
				}

				require.NoError(t, helm.UpgradeE(t, helmOptions, chartPath, releaseName))

				assertSecret(t, env, releaseName+"-masterkey", "masterkey", true, support.ExpectedLabels(
					releaseName,
					"zitadel",
					support.ExpectedVersion,
					"",
					nil,
				))

				if testCase.expected.machineKeySecret {
					assertSecret(t, env, testCase.expected.machineKeyName, testCase.expected.machineKeyContent, true, nil)
				} else if testCase.expected.machineKeyName != "" {
					assertSecret(t, env, testCase.expected.machineKeyName, "", false, nil)
				}

				if testCase.expected.machinePatSecret {
					assertSecret(t, env, testCase.expected.machinePatName, testCase.expected.machinePatContent, true, nil)
				} else if testCase.expected.machinePatName != "" {
					assertSecret(t, env, testCase.expected.machinePatName, "", false, nil)
				}

				if testCase.expected.loginClientSecret {
					assertSecret(t, env, testCase.expected.loginClientName, testCase.expected.loginClientContent, true, nil)
				} else if testCase.expected.loginClientName != "" {
					assertSecret(t, env, testCase.expected.loginClientName, "", false, nil)
				}
			})
		})
	}
}

// assertSecret verifies secret existence and content based on expectation flags.
// When shouldExist is true, validates the secret exists with the expected key.
// When shouldExist is false, validates the secret does not exist.
func assertSecret(t *testing.T, env *support.Env, secretName, expectedKey string, shouldExist bool, expectedLabels map[string]string) {
	secret, err := env.Client.CoreV1().Secrets(env.Kube.Namespace).Get(
		context.Background(),
		secretName,
		metav1.GetOptions{},
	)

	if shouldExist {
		require.NoError(t, err, "secret %q should exist", secretName)
		if expectedLabels != nil {
			support.AssertLabels(t, secret.Labels, expectedLabels)
		}
		require.Contains(t, secret.Data, expectedKey, "secret %q should contain key %q", secretName, expectedKey)
		require.NotEmpty(t, secret.Data[expectedKey], "secret %q key %q should not be empty", secretName, expectedKey)
	} else {
		require.True(t, apierrors.IsNotFound(err), "secret %q should not exist, got error: %v", secretName, err)
	}
}
