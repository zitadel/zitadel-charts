// file: charts/zitadel/smoke_test/servacc_test.go
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

type serviceAccountExpected struct {
	zitadelEnabled bool
	loginEnabled   bool

	zitadelAnnotations map[string]string
	loginAnnotations   map[string]string
}

func TestServiceAccountMatrix(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)

	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	testCases := []struct {
		name      string
		setValues map[string]string
		expected  serviceAccountExpected
	}{
		{
			name: "both-enabled-default",
			setValues: map[string]string{
				"serviceAccount.create":       "true",
				"login.enabled":               "true",
				"login.serviceAccount.create": "true",
			},
			expected: serviceAccountExpected{
				zitadelEnabled: true,
				loginEnabled:   true,
				zitadelAnnotations: map[string]string{
					"helm.sh/hook":               "pre-install,pre-upgrade",
					"helm.sh/hook-delete-policy": "before-hook-creation",
					"helm.sh/hook-weight":        "0",
				},
				loginAnnotations: map[string]string{
					"helm.sh/hook":               "pre-install,pre-upgrade",
					"helm.sh/hook-delete-policy": "before-hook-creation",
					"helm.sh/hook-weight":        "0",
				},
			},
		},
		{
			name: "both-enabled-with-annotations",
			setValues: map[string]string{
				"serviceAccount.create": "true",
				"serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn":       "arn:aws:iam::123456789012:role/zitadel-role",
				"serviceAccount.annotations.owner":                                "platform-team",
				"login.enabled":                                                   "true",
				"login.serviceAccount.create":                                     "true",
				"login.serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn": "arn:aws:iam::123456789012:role/login-role",
				"login.serviceAccount.annotations.team":                           "frontend",
			},
			expected: serviceAccountExpected{
				zitadelEnabled: true,
				loginEnabled:   true,
				zitadelAnnotations: map[string]string{
					"eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/zitadel-role",
					"owner":                      "platform-team",
					"helm.sh/hook":               "pre-install,pre-upgrade",
					"helm.sh/hook-delete-policy": "before-hook-creation",
					"helm.sh/hook-weight":        "0",
				},
				loginAnnotations: map[string]string{
					"eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/login-role",
					"team":                       "frontend",
					"helm.sh/hook":               "pre-install,pre-upgrade",
					"helm.sh/hook-delete-policy": "before-hook-creation",
					"helm.sh/hook-weight":        "0",
				},
			},
		},
		{
			name: "zitadel-enabled-login-disabled",
			setValues: map[string]string{
				"serviceAccount.create":                        "true",
				"serviceAccount.annotations.workload-identity": "enabled",
				"login.enabled":                                "true",
				"login.serviceAccount.create":                  "false",
			},
			expected: serviceAccountExpected{
				zitadelEnabled: true,
				loginEnabled:   false,
				zitadelAnnotations: map[string]string{
					"workload-identity":          "enabled",
					"helm.sh/hook":               "pre-install,pre-upgrade",
					"helm.sh/hook-delete-policy": "before-hook-creation",
					"helm.sh/hook-weight":        "0",
				},
			},
		},
		{
			name: "both-disabled",
			setValues: map[string]string{
				"serviceAccount.create":       "false",
				"login.enabled":               "true",
				"login.serviceAccount.create": "false",
			},
			expected: serviceAccountExpected{
				zitadelEnabled: false,
				loginEnabled:   false,
			},
		},
		{
			name: "zitadel-only-with-gcp-workload-identity",
			setValues: map[string]string{
				"serviceAccount.create": "true",
				"serviceAccount.annotations.iam\\.gke\\.io/gcp-service-account": "zitadel@project.iam.gserviceaccount.com",
				"login.enabled": "false",
			},
			expected: serviceAccountExpected{
				zitadelEnabled: true,
				loginEnabled:   false,
				zitadelAnnotations: map[string]string{
					"iam.gke.io/gcp-service-account": "zitadel@project.iam.gserviceaccount.com",
					"helm.sh/hook":                   "pre-install,pre-upgrade",
					"helm.sh/hook-delete-policy":     "before-hook-creation",
					"helm.sh/hook-weight":            "0",
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

				var expectedZitadelServiceAccount *corev1.ServiceAccount
				if testCase.expected.zitadelEnabled {
					expectedZitadelServiceAccount = &corev1.ServiceAccount{
						ObjectMeta: metav1.ObjectMeta{
							Name:        releaseName,
							Annotations: testCase.expected.zitadelAnnotations,
						},
					}
				}
				assertServiceAccount(t, env, expectedZitadelServiceAccount)

				var expectedLoginServiceAccount *corev1.ServiceAccount
				if testCase.expected.loginEnabled {
					expectedLoginServiceAccount = &corev1.ServiceAccount{
						ObjectMeta: metav1.ObjectMeta{
							Name:        releaseName + "-login",
							Annotations: testCase.expected.loginAnnotations,
						},
					}
				}
				assertServiceAccount(t, env, expectedLoginServiceAccount)
			})
		})
	}
}

func assertServiceAccount(t *testing.T, env *support.Env, expected *corev1.ServiceAccount) {
	t.Helper()

	if expected == nil {
		return
	}

	actualServiceAccount, err := env.Client.
		CoreV1().
		ServiceAccounts(env.Kube.Namespace).
		Get(context.Background(), expected.Name, metav1.GetOptions{})

	require.NoError(t, err, "failed to get ServiceAccount %s", expected.Name)
	require.Equal(t, expected.Annotations, actualServiceAccount.Annotations, "ServiceAccount annotations mismatch for %s", expected.Name)

	env.Logger.Logf(t, "✓ Verified ServiceAccount configuration for %s", expected.Name)
}
