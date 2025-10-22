package smoke_test_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zitadel/zitadel-charts/charts/zitadel/smoke_test/support"
)

type configMapExpected struct {
	zitadelEnabled bool
	loginEnabled   bool

	zitadelAnnotations map[string]string
	zitadelData        map[string]string

	loginAnnotations map[string]string
	loginData        map[string]string
}

func TestConfigMapMatrix(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)

	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	testCases := []struct {
		name      string
		setValues map[string]string
		expected  configMapExpected
	}{
		{
			name: "both-enabled-default",
			setValues: map[string]string{
				"login.enabled": "true",
			},
			expected: configMapExpected{
				zitadelEnabled: true,
				loginEnabled:   true,
				zitadelAnnotations: map[string]string{
					"helm.sh/hook":               "pre-install,pre-upgrade",
					"helm.sh/hook-delete-policy": "before-hook-creation",
					"helm.sh/hook-weight":        "0",
				},
				zitadelData: map[string]string{
					"zitadel-config-yaml": "",
				},
				loginAnnotations: map[string]string{
					"helm.sh/hook":               "pre-install,pre-upgrade",
					"helm.sh/hook-delete-policy": "before-hook-creation",
					"helm.sh/hook-weight":        "0",
				},
				loginData: map[string]string{
					".env": "",
				},
			},
		},
		{
			name: "both-enabled-with-annotations",
			setValues: map[string]string{
				"configMap.annotations.owner":      "platform-team",
				"login.enabled":                    "true",
				"login.configMap.annotations.team": "frontend",
			},
			expected: configMapExpected{
				zitadelEnabled: true,
				loginEnabled:   true,
				zitadelAnnotations: map[string]string{
					"owner":                      "platform-team",
					"helm.sh/hook":               "pre-install,pre-upgrade",
					"helm.sh/hook-delete-policy": "before-hook-creation",
					"helm.sh/hook-weight":        "0",
				},
				zitadelData: map[string]string{
					"zitadel-config-yaml": "",
				},
				loginAnnotations: map[string]string{
					"team":                       "frontend",
					"helm.sh/hook":               "pre-install,pre-upgrade",
					"helm.sh/hook-delete-policy": "before-hook-creation",
					"helm.sh/hook-weight":        "0",
				},
				loginData: map[string]string{
					".env": "",
				},
			},
		},
		{
			name: "zitadel-only-login-disabled",
			setValues: map[string]string{
				"configMap.annotations.config-version": "v2",
				"login.enabled":                        "false",
			},
			expected: configMapExpected{
				zitadelEnabled: true,
				loginEnabled:   false,
				zitadelAnnotations: map[string]string{
					"config-version":             "v2",
					"helm.sh/hook":               "pre-install,pre-upgrade",
					"helm.sh/hook-delete-policy": "before-hook-creation",
					"helm.sh/hook-weight":        "0",
				},
				zitadelData: map[string]string{
					"zitadel-config-yaml": "",
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
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
					"ingress.enabled":       "true",
					"login.ingress.enabled": "true",
				}

				releaseName := env.MakeRelease("zitadel-test", testCase.name)

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

				if err := helm.UpgradeE(t, helmOptions, chartPath, releaseName); err != nil {
					//dumpSetupAndInitJobLogs(t, env, releaseName)
					require.NoError(t, err)
				}

				var expectedZitadelConfigMap *corev1.ConfigMap
				if testCase.expected.zitadelEnabled {
					expectedZitadelConfigMap = &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:        releaseName + "-config-yaml",
							Annotations: testCase.expected.zitadelAnnotations,
						},
						Data: testCase.expected.zitadelData,
					}
				}
				assertConfigMap(t, env, expectedZitadelConfigMap)

				var expectedLoginConfigMap *corev1.ConfigMap
				if testCase.expected.loginEnabled {
					expectedLoginConfigMap = &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:        releaseName + "-login-config-dotenv",
							Annotations: testCase.expected.loginAnnotations,
						},
						Data: testCase.expected.loginData,
					}
				}
				assertConfigMap(t, env, expectedLoginConfigMap)
			})
		})
	}
}

func assertConfigMap(t *testing.T, env *support.Env, expected *corev1.ConfigMap) {
	t.Helper()

	if expected == nil {
		return
	}

	actualConfigMap, err := env.Client.
		CoreV1().
		ConfigMaps(env.Kube.Namespace).
		Get(context.Background(), expected.Name, metav1.GetOptions{})

	require.NoError(t, err, "failed to get ConfigMap %s", expected.Name)
	require.Equal(t, expected.Annotations, actualConfigMap.Annotations, "ConfigMap annotations mismatch for %s", expected.Name)

	for key := range expected.Data {
		_, exists := actualConfigMap.Data[key]
		require.True(t, exists, "ConfigMap data key %s missing for %s", key, expected.Name)
	}

	env.Logger.Logf(t, "✓ Verified ConfigMap configuration for %s", expected.Name)
}
