package smoke_test_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zitadel/zitadel-charts/charts/zitadel/smoke_test/support"
)

type securityContextsExpected struct {
	zitadelPodSecurityContext *corev1.PodSecurityContext
	zitadelSecurityContext    *corev1.SecurityContext
	loginPodSecurityContext   *corev1.PodSecurityContext
	loginSecurityContext      *corev1.SecurityContext
}

func TestSecurityContexts(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)

	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	int64Ptr := func(value int64) *int64 { return &value }
	boolPtr := func(value bool) *bool { return &value }

	testCases := []struct {
		name      string
		setValues map[string]string
		expected  securityContextsExpected
	}{
		{
			name: "defaults-use-global",
			setValues: map[string]string{
				"login.enabled":         "true",
				"login.ingress.enabled": "true",
			},
			expected: securityContextsExpected{
				zitadelPodSecurityContext: &corev1.PodSecurityContext{
					RunAsNonRoot: boolPtr(true),
					RunAsUser:    int64Ptr(1000),
					FSGroup:      int64Ptr(1000),
				},
				zitadelSecurityContext: &corev1.SecurityContext{
					RunAsNonRoot:             boolPtr(true),
					RunAsUser:                int64Ptr(1000),
					ReadOnlyRootFilesystem:   boolPtr(true),
					Privileged:               boolPtr(false),
					AllowPrivilegeEscalation: nil,
					Capabilities:             nil,
				},
				loginPodSecurityContext: &corev1.PodSecurityContext{
					RunAsNonRoot: boolPtr(true),
					RunAsUser:    int64Ptr(1000),
					FSGroup:      int64Ptr(1000),
				},
				loginSecurityContext: &corev1.SecurityContext{
					RunAsNonRoot:             boolPtr(true),
					RunAsUser:                int64Ptr(1000),
					ReadOnlyRootFilesystem:   boolPtr(true),
					Privileged:               boolPtr(false),
					AllowPrivilegeEscalation: nil,
					Capabilities:             nil,
				},
			},
		},
		{
			name: "component-overrides",
			setValues: map[string]string{
				"login.enabled":         "true",
				"login.ingress.enabled": "true",

				"zitadel.podSecurityContext.runAsNonRoot":          "true",
				"zitadel.podSecurityContext.runAsUser":             "2000",
				"zitadel.podSecurityContext.fsGroup":               "2000",
				"zitadel.podSecurityContext.seccompProfile.type":   "RuntimeDefault",
				"zitadel.securityContext.runAsNonRoot":             "true",
				"zitadel.securityContext.runAsUser":                "2000",
				"zitadel.securityContext.readOnlyRootFilesystem":   "true",
				"zitadel.securityContext.privileged":               "false",
				"zitadel.securityContext.allowPrivilegeEscalation": "false",
				"zitadel.securityContext.capabilities.drop[0]":     "ALL",
				"login.podSecurityContext.runAsNonRoot":            "true",
				"login.podSecurityContext.runAsUser":               "3000",
				"login.podSecurityContext.fsGroup":                 "3000",
				"login.podSecurityContext.seccompProfile.type":     "RuntimeDefault",
				"login.securityContext.runAsNonRoot":               "true",
				"login.securityContext.runAsUser":                  "3000",
				"login.securityContext.readOnlyRootFilesystem":     "true",
				"login.securityContext.privileged":                 "false",
				"login.securityContext.allowPrivilegeEscalation":   "false",
				"login.securityContext.capabilities.drop[0]":       "NET_RAW",
			},
			expected: securityContextsExpected{
				zitadelPodSecurityContext: &corev1.PodSecurityContext{
					RunAsNonRoot: boolPtr(true),
					RunAsUser:    int64Ptr(2000),
					FSGroup:      int64Ptr(2000),
					SeccompProfile: &corev1.SeccompProfile{
						Type: corev1.SeccompProfileTypeRuntimeDefault,
					},
				},
				zitadelSecurityContext: &corev1.SecurityContext{
					RunAsNonRoot:             boolPtr(true),
					RunAsUser:                int64Ptr(2000),
					ReadOnlyRootFilesystem:   boolPtr(true),
					Privileged:               boolPtr(false),
					AllowPrivilegeEscalation: boolPtr(false),
					Capabilities: &corev1.Capabilities{
						Drop: []corev1.Capability{"ALL"},
					},
				},
				loginPodSecurityContext: &corev1.PodSecurityContext{
					RunAsNonRoot: boolPtr(true),
					RunAsUser:    int64Ptr(3000),
					FSGroup:      int64Ptr(3000),
					SeccompProfile: &corev1.SeccompProfile{
						Type: corev1.SeccompProfileTypeRuntimeDefault,
					},
				},
				loginSecurityContext: &corev1.SecurityContext{
					RunAsNonRoot:             boolPtr(true),
					RunAsUser:                int64Ptr(3000),
					ReadOnlyRootFilesystem:   boolPtr(true),
					Privileged:               boolPtr(false),
					AllowPrivilegeEscalation: boolPtr(false),
					Capabilities: &corev1.Capabilities{
						Drop: []corev1.Capability{"NET_RAW"},
					},
				},
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
					dumpSetupAndInitJobLogs(t, env, releaseName)
					require.NoError(t, err)
				}

				ctx := context.Background()

				zitadelDeployment, err := env.Client.
					AppsV1().
					Deployments(env.Kube.Namespace).
					Get(ctx, releaseName, metav1.GetOptions{})
				require.NoError(t, err, "failed to get zitadel deployment")
				assertDeploymentSecurity(t, zitadelDeployment, testCase.expected.zitadelPodSecurityContext, testCase.expected.zitadelSecurityContext, "zitadel")

				loginDeployment, err := env.Client.
					AppsV1().
					Deployments(env.Kube.Namespace).
					Get(ctx, releaseName+"-login", metav1.GetOptions{})
				require.NoError(t, err, "failed to get login deployment")
				assertDeploymentSecurity(t, loginDeployment, testCase.expected.loginPodSecurityContext, testCase.expected.loginSecurityContext, "zitadel-login")
			})
		})
	}
}

func assertDeploymentSecurity(t *testing.T, deployment *appsv1.Deployment, expectedPodSC *corev1.PodSecurityContext, expectedContainerSC *corev1.SecurityContext, containerName string) {
	t.Helper()

	require.NotNil(t, deployment.Spec.Template.Spec.SecurityContext, "pod securityContext missing for deployment %s", deployment.Name)
	require.Equal(t, expectedPodSC, deployment.Spec.Template.Spec.SecurityContext, "pod securityContext mismatch for deployment %s", deployment.Name)

	container := findContainer(deployment.Spec.Template.Spec.Containers, containerName)
	require.NotNil(t, container, "container %s not found in deployment %s", containerName, deployment.Name)
	require.NotNil(t, container.SecurityContext, "container securityContext missing for %s", containerName)
	require.Equal(t, expectedContainerSC, container.SecurityContext, "container securityContext mismatch for %s", containerName)

	for _, c := range deployment.Spec.Template.Spec.InitContainers {
		require.NotNil(t, c.SecurityContext, "init container securityContext missing for %s in deployment %s", c.Name, deployment.Name)
		require.Equal(t, expectedContainerSC, c.SecurityContext, "init container securityContext mismatch for %s in deployment %s", c.Name, deployment.Name)
	}
	for _, c := range deployment.Spec.Template.Spec.Containers {
		require.NotNil(t, c.SecurityContext, "container securityContext missing for %s in deployment %s", c.Name, deployment.Name)
		require.Equal(t, expectedContainerSC, c.SecurityContext, "container securityContext mismatch for %s in deployment %s", c.Name, deployment.Name)
	}
}

func findContainer(containers []corev1.Container, name string) *corev1.Container {
	for i := range containers {
		if containers[i].Name == name {
			return &containers[i]
		}
	}
	return nil
}
