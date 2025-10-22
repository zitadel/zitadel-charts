package smoke_test_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zitadel/zitadel-charts/charts/zitadel/smoke_test/support"
)

type hpaExpected struct {
	zitadelEnabled bool
	loginEnabled   bool

	zitadelSpec        autoscalingv2.HorizontalPodAutoscalerSpec
	zitadelAnnotations map[string]string

	loginSpec        autoscalingv2.HorizontalPodAutoscalerSpec
	loginAnnotations map[string]string

	zitadelReplicas *int32
	loginReplicas   *int32
}

func TestAutoscalingMatrix(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)

	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	int32Ptr := func(value int32) *int32 { return &value }
	selectPolicyPtr := func(v autoscalingv2.ScalingPolicySelect) *autoscalingv2.ScalingPolicySelect { return &v }

	testCases := []struct {
		name      string
		setValues map[string]string
		expected  hpaExpected
	}{
		{
			name: "both-enabled-cpu-only",
			setValues: map[string]string{
				"zitadel.enabled":               "true",
				"zitadel.autoscaling.enabled":   "true",
				"zitadel.autoscaling.targetCPU": "60",

				"login.enabled":               "true",
				"login.autoscaling.enabled":   "true",
				"login.autoscaling.targetCPU": "55",
			},
			expected: hpaExpected{
				zitadelEnabled: true,
				loginEnabled:   true,
				zitadelSpec: autoscalingv2.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
						Kind:       "Deployment",
						Name:       "",
						APIVersion: "apps/v1",
					},
					MinReplicas: int32Ptr(3),
					MaxReplicas: 10,
					Metrics: []autoscalingv2.MetricSpec{
						{
							Type: autoscalingv2.ResourceMetricSourceType,
							Resource: &autoscalingv2.ResourceMetricSource{
								Name: "cpu",
								Target: autoscalingv2.MetricTarget{
									Type:               autoscalingv2.UtilizationMetricType,
									AverageUtilization: int32Ptr(60),
								},
							},
						},
					},
				},
				zitadelAnnotations: map[string]string{},
				loginSpec: autoscalingv2.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
						Kind:       "Deployment",
						Name:       "",
						APIVersion: "apps/v1",
					},
					MinReplicas: int32Ptr(3),
					MaxReplicas: 10,
					Metrics: []autoscalingv2.MetricSpec{
						{
							Type: autoscalingv2.ResourceMetricSourceType,
							Resource: &autoscalingv2.ResourceMetricSource{
								Name: "cpu",
								Target: autoscalingv2.MetricTarget{
									Type:               autoscalingv2.UtilizationMetricType,
									AverageUtilization: int32Ptr(55),
								},
							},
						},
					},
				},
				loginAnnotations: map[string]string{},
			},
		},
		{
			name: "zitadel-enabled-mem-only-login-disabled",
			setValues: map[string]string{
				"zitadel.enabled":                  "true",
				"zitadel.autoscaling.enabled":      "true",
				"zitadel.autoscaling.targetMemory": "70",

				"login.enabled":             "true",
				"login.autoscaling.enabled": "false",
				"login.replicaCount":        "2",
			},
			expected: hpaExpected{
				zitadelEnabled: true,
				loginEnabled:   false,
				zitadelSpec: autoscalingv2.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
						Kind:       "Deployment",
						Name:       "",
						APIVersion: "apps/v1",
					},
					MinReplicas: int32Ptr(3),
					MaxReplicas: 10,
					Metrics: []autoscalingv2.MetricSpec{
						{
							Type: autoscalingv2.ResourceMetricSourceType,
							Resource: &autoscalingv2.ResourceMetricSource{
								Name: "memory",
								Target: autoscalingv2.MetricTarget{
									Type:               autoscalingv2.UtilizationMetricType,
									AverageUtilization: int32Ptr(70),
								},
							},
						},
					},
				},
				zitadelAnnotations: map[string]string{},
				loginReplicas:      int32Ptr(2),
			},
		},
		{
			name: "both-enabled-with-annotations-and-behavior",
			setValues: map[string]string{
				"zitadel.enabled":                                                   "true",
				"zitadel.autoscaling.enabled":                                       "true",
				"zitadel.autoscaling.targetCPU":                                     "60",
				"zitadel.autoscaling.annotations.team":                              "platform",
				"zitadel.autoscaling.behavior.scaleDown.stabilizationWindowSeconds": "300",

				"login.enabled":                       "true",
				"login.autoscaling.enabled":           "true",
				"login.autoscaling.targetCPU":         "50",
				"login.autoscaling.annotations.owner": "iam",
				"login.autoscaling.behavior.scaleDown.stabilizationWindowSeconds": "300",
			},
			expected: hpaExpected{
				zitadelEnabled: true,
				loginEnabled:   true,
				zitadelSpec: autoscalingv2.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
						Kind:       "Deployment",
						Name:       "",
						APIVersion: "apps/v1",
					},
					MinReplicas: int32Ptr(3),
					MaxReplicas: 10,
					Metrics: []autoscalingv2.MetricSpec{
						{
							Type: autoscalingv2.ResourceMetricSourceType,
							Resource: &autoscalingv2.ResourceMetricSource{
								Name: "cpu",
								Target: autoscalingv2.MetricTarget{
									Type:               autoscalingv2.UtilizationMetricType,
									AverageUtilization: int32Ptr(60),
								},
							},
						},
					},
					Behavior: &autoscalingv2.HorizontalPodAutoscalerBehavior{
						ScaleUp: &autoscalingv2.HPAScalingRules{
							StabilizationWindowSeconds: int32Ptr(0),
							SelectPolicy:               selectPolicyPtr(autoscalingv2.MaxChangePolicySelect),
							Policies: []autoscalingv2.HPAScalingPolicy{
								{
									Type:          autoscalingv2.PodsScalingPolicy,
									Value:         4,
									PeriodSeconds: 15,
								},
								{
									Type:          autoscalingv2.PercentScalingPolicy,
									Value:         100,
									PeriodSeconds: 15,
								},
							},
						},
						ScaleDown: &autoscalingv2.HPAScalingRules{
							StabilizationWindowSeconds: int32Ptr(300),
							SelectPolicy:               selectPolicyPtr(autoscalingv2.MaxChangePolicySelect),
							Policies: []autoscalingv2.HPAScalingPolicy{
								{
									Type:          autoscalingv2.PercentScalingPolicy,
									Value:         100,
									PeriodSeconds: 15,
								},
							},
						},
					},
				},
				zitadelAnnotations: map[string]string{
					"team": "platform",
				},
				loginSpec: autoscalingv2.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
						Kind:       "Deployment",
						Name:       "",
						APIVersion: "apps/v1",
					},
					MinReplicas: int32Ptr(3),
					MaxReplicas: 10,
					Metrics: []autoscalingv2.MetricSpec{
						{
							Type: autoscalingv2.ResourceMetricSourceType,
							Resource: &autoscalingv2.ResourceMetricSource{
								Name: "cpu",
								Target: autoscalingv2.MetricTarget{
									Type:               autoscalingv2.UtilizationMetricType,
									AverageUtilization: int32Ptr(50),
								},
							},
						},
					},
					Behavior: &autoscalingv2.HorizontalPodAutoscalerBehavior{
						ScaleUp: &autoscalingv2.HPAScalingRules{
							StabilizationWindowSeconds: int32Ptr(0),
							SelectPolicy:               selectPolicyPtr(autoscalingv2.MaxChangePolicySelect),
							Policies: []autoscalingv2.HPAScalingPolicy{
								{
									Type:          autoscalingv2.PodsScalingPolicy,
									Value:         4,
									PeriodSeconds: 15,
								},
								{
									Type:          autoscalingv2.PercentScalingPolicy,
									Value:         100,
									PeriodSeconds: 15,
								},
							},
						},
						ScaleDown: &autoscalingv2.HPAScalingRules{
							StabilizationWindowSeconds: int32Ptr(300),
							SelectPolicy:               selectPolicyPtr(autoscalingv2.MaxChangePolicySelect),
							Policies: []autoscalingv2.HPAScalingPolicy{
								{
									Type:          autoscalingv2.PercentScalingPolicy,
									Value:         100,
									PeriodSeconds: 15,
								},
							},
						},
					},
				},
				loginAnnotations: map[string]string{
					"owner": "iam",
				},
			},
		},
		{
			name: "both-disabled-replicas-set",
			setValues: map[string]string{
				"zitadel.enabled":             "true",
				"zitadel.autoscaling.enabled": "false",
				"replicaCount":                "2",

				"login.enabled":             "true",
				"login.autoscaling.enabled": "false",
				"login.replicaCount":        "2",
			},
			expected: hpaExpected{
				zitadelEnabled:  false,
				loginEnabled:    false,
				zitadelReplicas: int32Ptr(2),
				loginReplicas:   int32Ptr(2),
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
					//dumpSetupAndInitJobLogs(t, env, releaseName)
					require.NoError(t, err)
				}

				zitadelDeployment := mustGetDeployment(t, env, releaseName)
				loginDeployment := mustGetDeployment(t, env, releaseName+"-login")

				var expectedZitadelHPA *autoscalingv2.HorizontalPodAutoscaler
				if testCase.expected.zitadelEnabled {
					zitadelAnnotations := map[string]string{
						"meta.helm.sh/release-name":      releaseName,
						"meta.helm.sh/release-namespace": env.Namespace,
					}
					for k, v := range testCase.expected.zitadelAnnotations {
						zitadelAnnotations[k] = v
					}

					zitadelSpec := testCase.expected.zitadelSpec
					zitadelSpec.ScaleTargetRef.Name = zitadelDeployment.Name

					expectedZitadelHPA = &autoscalingv2.HorizontalPodAutoscaler{
						ObjectMeta: metav1.ObjectMeta{
							Name:        releaseName,
							Annotations: zitadelAnnotations,
						},
						Spec: zitadelSpec,
					}
				}
				assertHPA(t, env, expectedZitadelHPA)

				if !testCase.expected.zitadelEnabled {
					require.NotNil(t, zitadelDeployment.Spec.Replicas)
					if testCase.expected.zitadelReplicas != nil {
						require.Equal(t, *testCase.expected.zitadelReplicas, *zitadelDeployment.Spec.Replicas)
					}
				}

				var expectedLoginHPA *autoscalingv2.HorizontalPodAutoscaler
				if testCase.expected.loginEnabled {
					loginAnnotations := map[string]string{
						"meta.helm.sh/release-name":      releaseName,
						"meta.helm.sh/release-namespace": env.Namespace,
					}
					for k, v := range testCase.expected.loginAnnotations {
						loginAnnotations[k] = v
					}

					loginSpec := testCase.expected.loginSpec
					loginSpec.ScaleTargetRef.Name = loginDeployment.Name

					expectedLoginHPA = &autoscalingv2.HorizontalPodAutoscaler{
						ObjectMeta: metav1.ObjectMeta{
							Name:        releaseName + "-login",
							Annotations: loginAnnotations,
						},
						Spec: loginSpec,
					}
				}
				assertHPA(t, env, expectedLoginHPA)

				if !testCase.expected.loginEnabled {
					require.NotNil(t, loginDeployment.Spec.Replicas)
					if testCase.expected.loginReplicas != nil {
						require.Equal(t, *testCase.expected.loginReplicas, *loginDeployment.Spec.Replicas)
					}
				}
			})
		})
	}
}

func mustGetDeployment(t *testing.T, env *support.Env, deploymentName string) *appsv1.Deployment {
	var deployment *appsv1.Deployment
	require.Eventually(
		t,
		func() bool {
			retrievedDeployment, err := k8s.GetDeploymentE(t, env.Kube, deploymentName)
			if err != nil {
				return false
			}
			deployment = retrievedDeployment
			return true
		},
		2*time.Minute,
		2*time.Second,
	)
	return deployment
}

func assertHPA(t *testing.T, env *support.Env, expected *autoscalingv2.HorizontalPodAutoscaler) {
	t.Helper()

	if expected == nil {
		return
	}

	actualHPA, err := env.Client.
		AutoscalingV2().
		HorizontalPodAutoscalers(env.Kube.Namespace).
		Get(context.Background(), expected.Name, metav1.GetOptions{})

	require.NoError(t, err, "failed to get HPA %s", expected.Name)
	require.Equal(t, expected.Spec, actualHPA.Spec, "HPA spec mismatch for %s", expected.Name)
	require.Equal(t, expected.Annotations, actualHPA.Annotations, "HPA annotations mismatch for %s", expected.Name)

	env.Logger.Logf(t, "✓ Verified HPA configuration for %s", expected.Name)
}
