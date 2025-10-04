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
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zitadel/zitadel-charts/charts/zitadel/smoke_test/support"
)

// hpaExpected describes the expected state for a single autoscaling test case.
type hpaExpected struct {
	zitadelEnabled bool
	loginEnabled   bool

	zitadelCPU    *int32
	zitadelMemory *int32
	loginCPU      *int32
	loginMemory   *int32

	zitadelAnnKey string
	zitadelAnnVal string
	loginAnnKey   string
	loginAnnVal   string

	zitadelHasBehavior bool
	loginHasBehavior   bool

	zitadelReplicas *int32
	loginReplicas   *int32

	zitadelSkipMetrics bool
	loginSkipMetrics   bool
}

// TestAutoscalingMatrix validates HPA behavior across a matrix of values. It
// connects to the cluster once, then runs each subtest in its own ephemeral
// namespace with a fresh PostgreSQL instance and unique Helm release.
func TestAutoscalingMatrix(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)

	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	int32Ptr := func(value int32) *int32 { return &value }

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
				zitadelCPU:     int32Ptr(60),
				loginCPU:       int32Ptr(55),
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
				zitadelMemory:  int32Ptr(70),
				loginReplicas:  int32Ptr(2),
			},
		},
		{
			name: "both-enabled-no-metrics-block",
			setValues: map[string]string{
				"zitadel.enabled":             "true",
				"zitadel.autoscaling.enabled": "true",

				"login.enabled":             "true",
				"login.autoscaling.enabled": "true",
			},
			expected: hpaExpected{
				zitadelEnabled:     true,
				loginEnabled:       true,
				zitadelSkipMetrics: true,
				loginSkipMetrics:   true,
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
				zitadelEnabled:     true,
				loginEnabled:       true,
				zitadelCPU:         int32Ptr(60),
				loginCPU:           int32Ptr(50),
				zitadelAnnKey:      "team",
				zitadelAnnVal:      "platform",
				loginAnnKey:        "owner",
				loginAnnVal:        "iam",
				zitadelHasBehavior: true,
				loginHasBehavior:   true,
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
					dumpSetupAndInitJobLogs(t, env, releaseName)
					require.NoError(t, err)
				}

				zitadelDeployment := mustGetDeployment(t, env, releaseName)
				loginDeployment := mustGetDeployment(t, env, releaseName+"-login")

				if testCase.expected.zitadelEnabled {
					zitadelHPA := mustGetHPA(t, env, releaseName)
					assertHPA(t, zitadelHPA, hpaExpectedNormalized{
						TargetKind:  "Deployment",
						TargetName:  zitadelDeployment.Name,
						CPU:         testCase.expected.zitadelCPU,
						Memory:      testCase.expected.zitadelMemory,
						AnnKey:      testCase.expected.zitadelAnnKey,
						AnnVal:      testCase.expected.zitadelAnnVal,
						HasBehavior: testCase.expected.zitadelHasBehavior,
						SkipMetrics: testCase.expected.zitadelSkipMetrics,
					})
				} else {
					assertHPANotFound(t, env, releaseName)
					require.NotNil(t, zitadelDeployment.Spec.Replicas)
					if testCase.expected.zitadelReplicas != nil {
						require.Equal(t, *testCase.expected.zitadelReplicas, *zitadelDeployment.Spec.Replicas)
					}
				}

				if testCase.expected.loginEnabled {
					loginHPA := mustGetHPA(t, env, releaseName+"-login")
					assertHPA(t, loginHPA, hpaExpectedNormalized{
						TargetKind:  "Deployment",
						TargetName:  loginDeployment.Name,
						CPU:         testCase.expected.loginCPU,
						Memory:      testCase.expected.loginMemory,
						AnnKey:      testCase.expected.loginAnnKey,
						AnnVal:      testCase.expected.loginAnnVal,
						HasBehavior: testCase.expected.loginHasBehavior,
						SkipMetrics: testCase.expected.loginSkipMetrics,
					})
				} else {
					assertHPANotFound(t, env, releaseName+"-login")
					require.NotNil(t, loginDeployment.Spec.Replicas)
					if testCase.expected.loginReplicas != nil {
						require.Equal(t, *testCase.expected.loginReplicas, *loginDeployment.Spec.Replicas)
					}
				}
			})
		})
	}
}

// mustGetDeployment polls until the named Deployment exists and returns it.
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

// mustGetHPA polls until the named HPA exists and returns it.
func mustGetHPA(t *testing.T, env *support.Env, hpaName string) *autoscalingv2.HorizontalPodAutoscaler {
	var horizontalPodAutoscaler *autoscalingv2.HorizontalPodAutoscaler
	require.Eventually(
		t,
		func() bool {
			retrievedHPA, err := env.Client.
				AutoscalingV2().
				HorizontalPodAutoscalers(env.Kube.Namespace).
				Get(context.Background(), hpaName, metav1.GetOptions{})
			if err != nil {
				return false
			}
			horizontalPodAutoscaler = retrievedHPA
			return true
		},
		2*time.Minute,
		2*time.Second,
	)
	return horizontalPodAutoscaler
}

// assertHPANotFound asserts that the named HPA is not present.
func assertHPANotFound(t *testing.T, env *support.Env, hpaName string) {
	_, err := env.Client.
		AutoscalingV2().
		HorizontalPodAutoscalers(env.Kube.Namespace).
		Get(context.Background(), hpaName, metav1.GetOptions{})
	require.True(t, apierrors.IsNotFound(err), "expected HPA %q to be NotFound, got: %v", hpaName, err)
}

// dumpSetupAndInitJobLogs prints logs from setup/init jobs when a Helm action
// fails, helping with debugging deployment issues.
func dumpSetupAndInitJobLogs(t *testing.T, env *support.Env, releaseName string) {
	namespace := env.Kube.Namespace
	jobNames := []string{fmt.Sprintf("%s-setup", releaseName), fmt.Sprintf("%s-init", releaseName)}

	for _, jobName := range jobNames {
		labelSelector := fmt.Sprintf("job-name=%s", jobName)
		pods := listPodsE(t, env, labelSelector)

		for _, pod := range pods {
			for _, container := range pod.Spec.Containers {
				logOutput, _ := k8s.RunKubectlAndGetOutputE(
					t,
					env.Kube,
					"logs",
					pod.Name, "-n", namespace, "-c", container.Name, "--tail=500",
				)
				env.Logger.Logf(t, "---- logs: pod=%s container=%s ----\n%s\n---- end logs ----", pod.Name, container.Name, logOutput)
			}
			for _, initContainer := range pod.Spec.InitContainers {
				logOutput, _ := k8s.RunKubectlAndGetOutputE(
					t,
					env.Kube,
					"logs",
					pod.Name, "-n", namespace, "-c", initContainer.Name, "--tail=500",
				)
				env.Logger.Logf(t, "---- logs: pod=%s initContainer=%s ----\n%s\n---- end logs ----", pod.Name, initContainer.Name, logOutput)
			}
		}
	}
}

// listPodsE lists pods matching a label selector and returns an empty slice on
// error to avoid breaking test execution.
func listPodsE(t *testing.T, env *support.Env, labelSelector string) []corev1.Pod {
	podList, err := env.Client.CoreV1().Pods(env.Kube.Namespace).List(
		context.Background(),
		metav1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		env.Logger.Logf(t, "warn: list pods selector=%q: %v", labelSelector, err)
		return []corev1.Pod{}
	}
	return podList.Items
}

// hpaExpectedNormalized defines the normalized set of HPA fields we assert as
// a complete validation check.
type hpaExpectedNormalized struct {
	TargetKind  string
	TargetName  string
	CPU         *int32
	Memory      *int32
	AnnKey      string
	AnnVal      string
	HasBehavior bool
	SkipMetrics bool
}

// normalizeHPA converts an HPA into a stable, comparable view of target,
// metrics, selected annotation, and whether a Behavior block is present.
func normalizeHPA(horizontalPodAutoscaler *autoscalingv2.HorizontalPodAutoscaler, annotationKey string, skipMetrics bool) hpaExpectedNormalized {
	var cpuUtilization, memoryUtilization *int32
	if !skipMetrics {
		for _, metric := range horizontalPodAutoscaler.Spec.Metrics {
			if metric.Type != autoscalingv2.ResourceMetricSourceType {
				continue
			}
			if metric.Resource.Target.Type == autoscalingv2.UtilizationMetricType {
				if metric.Resource.Name == "cpu" && metric.Resource.Target.AverageUtilization != nil {
					value := *metric.Resource.Target.AverageUtilization
					cpuUtilization = &value
				}
				if metric.Resource.Name == "memory" && metric.Resource.Target.AverageUtilization != nil {
					value := *metric.Resource.Target.AverageUtilization
					memoryUtilization = &value
				}
			}
		}
	}
	var annotationValue string
	if annotationKey != "" {
		annotationValue = horizontalPodAutoscaler.Annotations[annotationKey]
	}
	return hpaExpectedNormalized{
		TargetKind:  horizontalPodAutoscaler.Spec.ScaleTargetRef.Kind,
		TargetName:  horizontalPodAutoscaler.Spec.ScaleTargetRef.Name,
		CPU:         cpuUtilization,
		Memory:      memoryUtilization,
		AnnKey:      annotationKey,
		AnnVal:      annotationValue,
		HasBehavior: horizontalPodAutoscaler.Spec.Behavior != nil,
		SkipMetrics: skipMetrics,
	}
}

// assertHPA compares an actual HPA against a normalized expected view and
// validates all relevant fields match the expected configuration.
func assertHPA(t *testing.T, actualHPA *autoscalingv2.HorizontalPodAutoscaler, expectedHPA hpaExpectedNormalized) {
	actualNormalized := normalizeHPA(actualHPA, expectedHPA.AnnKey, expectedHPA.SkipMetrics)
	require.Equal(t, expectedHPA.TargetKind, actualNormalized.TargetKind, "HPA target kind mismatch")
	require.Equal(t, expectedHPA.TargetName, actualNormalized.TargetName, "HPA target name mismatch")

	if expectedHPA.SkipMetrics {
		require.Nil(t, actualNormalized.CPU, "expected no CPU metric in HPA, but found one")
		require.Nil(t, actualNormalized.Memory, "expected no Memory metric in HPA, but found one")
	} else {
		if expectedHPA.CPU == nil {
			require.Nil(t, actualNormalized.CPU, "expected no CPU metric in HPA, but found one")
		} else {
			require.NotNil(t, actualNormalized.CPU, "expected CPU metric not found")
			require.Equal(t, *expectedHPA.CPU, *actualNormalized.CPU, "CPU AverageUtilization mismatch")
		}
		if expectedHPA.Memory == nil {
			require.Nil(t, actualNormalized.Memory, "expected no Memory metric in HPA, but found one")
		} else {
			require.NotNil(t, actualNormalized.Memory, "expected Memory metric not found")
			require.Equal(t, *expectedHPA.Memory, *actualNormalized.Memory, "Memory AverageUtilization mismatch")
		}
	}

	if expectedHPA.AnnKey != "" {
		require.Equal(t, expectedHPA.AnnVal, actualNormalized.AnnVal, "annotation %q mismatch", expectedHPA.AnnKey)
	}

	if expectedHPA.HasBehavior {
		require.True(t, actualNormalized.HasBehavior, "expected HPA Behavior to be set")
	} else {
		require.False(t, actualNormalized.HasBehavior, "expected HPA Behavior to be nil")
	}
}
