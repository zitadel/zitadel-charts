package smoke_test_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/zitadel/zitadel-charts/charts/zitadel/smoke_test/support"
)

type serviceExpected struct {
	zitadelEnabled bool
	loginEnabled   bool

	zitadelSpec        corev1.ServiceSpec
	zitadelAnnotations map[string]string

	loginSpec        corev1.ServiceSpec
	loginAnnotations map[string]string
}

func TestServiceMatrix(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)

	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	commonSetValues := map[string]string{
		"zitadel.masterkey":                                         "x123456789012345678901234567891y",
		"zitadel.configmapConfig.ExternalDomain":                    "pg-insecure.127.0.0.1.sslip.io",
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

	testCases := []struct {
		name      string
		setValues map[string]string
		expected  serviceExpected
	}{
		{
			name: "both-enabled-default-clusterip",
			setValues: map[string]string{
				"login.enabled": "true",
			},
			expected: serviceExpected{
				zitadelEnabled: true,
				loginEnabled:   true,
				zitadelSpec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeClusterIP,
					Ports: []corev1.ServicePort{
						{
							Port:       8080,
							Protocol:   corev1.ProtocolTCP,
							TargetPort: intstr.FromInt32(8080),
						},
					},
					Selector: map[string]string{
						"app.kubernetes.io/component": "start",
						"app.kubernetes.io/instance":  "",
						"app.kubernetes.io/name":      "zitadel",
					},
				},
				zitadelAnnotations: map[string]string{},
				loginSpec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeClusterIP,
					Ports: []corev1.ServicePort{
						{
							Port:       9091,
							Protocol:   corev1.ProtocolTCP,
							TargetPort: intstr.FromInt32(3000),
						},
					},
					Selector: map[string]string{
						"app.kubernetes.io/component": "login",
						"app.kubernetes.io/instance":  "",
						"app.kubernetes.io/name":      "zitadel-login",
					},
				},
				loginAnnotations: map[string]string{},
			},
		},
		{
			name: "both-enabled-custom-ports",
			setValues: map[string]string{
				"service.port":       "9090",
				"login.enabled":      "true",
				"login.service.port": "9091",
			},
			expected: serviceExpected{
				zitadelEnabled: true,
				loginEnabled:   true,
				zitadelSpec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeClusterIP,
					Ports: []corev1.ServicePort{
						{
							Port:       9090,
							Protocol:   corev1.ProtocolTCP,
							TargetPort: intstr.FromInt32(8080),
						},
					},
					Selector: map[string]string{
						"app.kubernetes.io/component": "start",
						"app.kubernetes.io/instance":  "",
						"app.kubernetes.io/name":      "zitadel",
					},
				},
				zitadelAnnotations: map[string]string{},
				loginSpec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeClusterIP,
					Ports: []corev1.ServicePort{
						{
							Port:       9091,
							Protocol:   corev1.ProtocolTCP,
							TargetPort: intstr.FromInt32(3000),
						},
					},
					Selector: map[string]string{
						"app.kubernetes.io/component": "login",
						"app.kubernetes.io/instance":  "",
						"app.kubernetes.io/name":      "zitadel-login",
					},
				},
				loginAnnotations: map[string]string{},
			},
		},
		{
			name: "both-enabled-with-annotations",
			setValues: map[string]string{
				"service.annotations.cloud\\.google\\.com/load-balancer-type": "Internal",
				"service.annotations.owner":                                   "platform-team",
				"login.enabled":                                               "true",
				"login.service.annotations.service\\.beta\\.kubernetes\\.io/aws-load-balancer-internal": "yes",
			},
			expected: serviceExpected{
				zitadelEnabled: true,
				loginEnabled:   true,
				zitadelSpec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeClusterIP,
					Ports: []corev1.ServicePort{
						{
							Port:       8080,
							Protocol:   corev1.ProtocolTCP,
							TargetPort: intstr.FromInt32(8080),
						},
					},
					Selector: map[string]string{
						"app.kubernetes.io/component": "start",
						"app.kubernetes.io/instance":  "",
						"app.kubernetes.io/name":      "zitadel",
					},
				},
				zitadelAnnotations: map[string]string{
					"cloud.google.com/load-balancer-type": "Internal",
					"owner":                               "platform-team",
				},
				loginSpec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeClusterIP,
					Ports: []corev1.ServicePort{
						{
							Port:       9091,
							Protocol:   corev1.ProtocolTCP,
							TargetPort: intstr.FromInt32(3000),
						},
					},
					Selector: map[string]string{
						"app.kubernetes.io/component": "login",
						"app.kubernetes.io/instance":  "",
						"app.kubernetes.io/name":      "zitadel-login",
					},
				},
				loginAnnotations: map[string]string{
					"service.beta.kubernetes.io/aws-load-balancer-internal": "yes",
				},
			},
		},
		{
			name: "zitadel-only-login-disabled",
			setValues: map[string]string{
				"service.type":  "ClusterIP",
				"service.port":  "8888",
				"login.enabled": "false",
			},
			expected: serviceExpected{
				zitadelEnabled: true,
				loginEnabled:   false,
				zitadelSpec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeClusterIP,
					Ports: []corev1.ServicePort{
						{
							Port:       8888,
							Protocol:   corev1.ProtocolTCP,
							TargetPort: intstr.FromInt32(8080),
						},
					},
					Selector: map[string]string{
						"app.kubernetes.io/component": "start",
						"app.kubernetes.io/instance":  "",
						"app.kubernetes.io/name":      "zitadel",
					},
				},
				zitadelAnnotations: map[string]string{},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			support.WithNamespace(t, cluster, func(env *support.Env) {
				env.Logger.Logf(t, "namespace %q created; installing PostgreSQL…", env.Namespace)
				support.WithPostgres(t, env)

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

				var expectedZitadelService *corev1.Service
				if testCase.expected.zitadelEnabled {
					zitadelAnnotations := map[string]string{
						"meta.helm.sh/release-name":                           releaseName,
						"meta.helm.sh/release-namespace":                      env.Namespace,
						"traefik.ingress.kubernetes.io/service.serversscheme": "h2c",
					}
					for k, v := range testCase.expected.zitadelAnnotations {
						zitadelAnnotations[k] = v
					}

					zitadelSpec := testCase.expected.zitadelSpec
					zitadelSpec.Selector["app.kubernetes.io/instance"] = releaseName

					expectedZitadelService = &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:        releaseName,
							Annotations: zitadelAnnotations,
						},
						Spec: zitadelSpec,
					}
				}
				assertService(t, env, expectedZitadelService)

				var expectedLoginService *corev1.Service
				if testCase.expected.loginEnabled {
					loginAnnotations := map[string]string{
						"meta.helm.sh/release-name":      releaseName,
						"meta.helm.sh/release-namespace": env.Namespace,
					}
					for k, v := range testCase.expected.loginAnnotations {
						loginAnnotations[k] = v
					}

					loginSpec := testCase.expected.loginSpec
					loginSpec.Selector["app.kubernetes.io/instance"] = releaseName

					expectedLoginService = &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:        releaseName + "-login",
							Annotations: loginAnnotations,
						},
						Spec: loginSpec,
					}
				}
				assertService(t, env, expectedLoginService)
			})
		})
	}
}

func assertService(t *testing.T, env *support.Env, expected *corev1.Service) {
	t.Helper()

	if expected == nil {
		return
	}

	actualService, err := env.Client.
		CoreV1().
		Services(env.Kube.Namespace).
		Get(context.Background(), expected.Name, metav1.GetOptions{})

	require.NoError(t, err, "failed to get Service %s", expected.Name)
	require.Equal(t, expected.Spec.Type, actualService.Spec.Type, "Service type mismatch for %s", expected.Name)
	require.Equal(t, expected.Spec.Selector, actualService.Spec.Selector, "Service selector mismatch for %s", expected.Name)
	require.GreaterOrEqual(t, len(actualService.Spec.Ports), 1, "Service should have at least one port for %s", expected.Name)
	require.Equal(t, expected.Spec.Ports[0].Port, actualService.Spec.Ports[0].Port, "Service port mismatch for %s", expected.Name)
	require.Equal(t, expected.Spec.Ports[0].TargetPort, actualService.Spec.Ports[0].TargetPort, "Service targetPort mismatch for %s", expected.Name)
	require.Equal(t, expected.Annotations, actualService.Annotations, "Service annotations mismatch for %s", expected.Name)

	env.Logger.Logf(t, "✓ Verified Service configuration for %s", expected.Name)
}
