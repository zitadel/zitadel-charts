package smoke_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zitadel/zitadel-charts/charts/zitadel/smoke_test/support"
)

// masterkeyExpected describes the expected state for masterkey configuration test cases.
type masterkeyExpected struct {
	secretCreated      bool
	secretName         string
	masterkeyGenerated bool
	masterkeyValue     string
	immutable          bool
	shouldFail         bool
}

// TestMasterkeySecretLogic validates masterkey configuration behavior across various
// scenarios including auto-generation, explicit values, and external secrets.
func TestMasterkeySecretLogic(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)

	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	commonSetValues := map[string]string{
		"zitadel.configmapConfig.ExternalDomain":                   "masterkey-test.127.0.0.1.sslip.io",
		"zitadel.configmapConfig.ExternalPort":                     "443",
		"zitadel.configmapConfig.TLS.Enabled":                      "false",
		"zitadel.configmapConfig.Database.Postgres.Host":           "db-postgresql",
		"zitadel.configmapConfig.Database.Postgres.Port":           "5432",
		"zitadel.configmapConfig.Database.Postgres.Database":       "zitadel",
		"zitadel.configmapConfig.Database.Postgres.User.Username":  "postgres",
		"zitadel.configmapConfig.Database.Postgres.User.SSL.Mode":  "disable",
		"zitadel.configmapConfig.Database.Postgres.Admin.Username": "postgres",
		"zitadel.configmapConfig.Database.Postgres.Admin.SSL.Mode": "disable",
	}

	testCases := []struct {
		name      string
		setValues map[string]string
		expected  masterkeyExpected
	}{
		{
			name:      "auto-generate-masterkey",
			setValues: map[string]string{},
			expected: masterkeyExpected{
				secretCreated:      true,
				secretName:         "",
				masterkeyGenerated: true,
				immutable:          true,
			},
		},
		{
			name: "explicit-masterkey-value",
			setValues: map[string]string{
				"zitadel.masterkey": "abcd1234efgh5678ijkl9012mnop3456",
			},
			expected: masterkeyExpected{
				secretCreated:      true,
				secretName:         "",
				masterkeyValue:     "abcd1234efgh5678ijkl9012mnop3456",
				masterkeyGenerated: false,
				immutable:          true,
			},
		},
		{
			name: "external-secret-reference",
			setValues: map[string]string{
				"zitadel.masterkeySecretName": "my-external-masterkey",
			},
			expected: masterkeyExpected{
				secretCreated:      false,
				secretName:         "my-external-masterkey",
				masterkeyGenerated: false,
			},
		},
		{
			name: "both-set-should-fail",
			setValues: map[string]string{
				"zitadel.masterkey":           "abcd1234efgh5678ijkl9012mnop3456",
				"zitadel.masterkeySecretName": "my-external-masterkey",
			},
			expected: masterkeyExpected{
				secretCreated: false,
				shouldFail:    true,
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.name == "external-secret-reference" {
				t.Skip("Skipping external secret test - requires pre-created secret")
				return
			}

			support.WithNamespace(t, cluster, func(env *support.Env) {
				env.Logger.Logf(t, "namespace %q created; installing PostgreSQL…", env.Namespace)
				support.WithPostgres(t, env)

				releaseName := env.MakeRelease("zitadel-masterkey-test", testCase.name)

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
						"upgrade": {"--install", "--timeout", "10m"},
					},
				}

				if testCase.name == "both-set-should-fail" {
					err := helm.UpgradeE(t, helmOptions, chartPath, releaseName)
					require.Error(t, err, "helm install should fail when both masterkey and masterkeySecretName are set")
					require.Contains(t, err.Error(), "set either .Values.zitadel.masterkey xor .Values.zitadel.masterkeySecretName")
					return
				}

				require.NoError(t, helm.UpgradeE(t, helmOptions, chartPath, releaseName))

				secretName := releaseName + "-masterkey"
				if testCase.expected.secretName != "" {
					secretName = testCase.expected.secretName
				}

				assertSecret(t, env, secretName, testCase.expected)
			})
		})
	}
}

// getMasterkeySecret retrieves a masterkey secret by name, returning nil if not found.
func getMasterkeySecret(t *testing.T, env *support.Env, secretName string) *corev1.Secret {
	secret, err := env.Client.CoreV1().Secrets(env.Kube.Namespace).Get(
		context.Background(),
		secretName,
		metav1.GetOptions{},
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		require.NoError(t, err, "unexpected error retrieving secret %q", secretName)
	}
	return secret
}

// assertSecret verifies secret existence and content based on expected configuration.
func assertSecret(t *testing.T, env *support.Env, secretName string, expected masterkeyExpected) {
	secret := getMasterkeySecret(t, env, secretName)

	if expected.secretCreated {
		require.NotNil(t, secret, "masterkey secret should be created")

		masterkeyData, exists := secret.Data["masterkey"]
		require.True(t, exists, "secret should contain 'masterkey' key")
		require.NotEmpty(t, masterkeyData, "masterkey should not be empty")

		masterkeyValue := string(masterkeyData)

		if expected.masterkeyGenerated {
			require.Len(t, masterkeyValue, 32, "generated masterkey should be 32 characters")
			require.Regexp(t, "^[A-Za-z0-9]+$", masterkeyValue, "generated masterkey should be alphanumeric")
		} else if expected.masterkeyValue != "" {
			require.Equal(t, expected.masterkeyValue, masterkeyValue, "masterkey should match provided value")
		}

		if expected.immutable {
			require.True(t, secret.Immutable != nil && *secret.Immutable, "secret should be marked as immutable")
		}

		hookAnnotation, hasHook := secret.Annotations["helm.sh/hook"]
		require.True(t, hasHook, "secret should have helm hook annotation")
		require.Equal(t, "pre-install", hookAnnotation, "hook should be pre-install only, not pre-upgrade")
	} else {
		require.Nil(t, secret, "secret %q should not exist", secretName)
	}
}
