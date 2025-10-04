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
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/zitadel/zitadel-charts/charts/zitadel/smoke_test/support"
)

type pdbExpected struct {
	zitadelEnabled bool
	loginEnabled   bool

	zitadelMinAvailable   *intstr.IntOrString
	zitadelMaxUnavailable *intstr.IntOrString
	zitadelAnnKey         string
	zitadelAnnVal         string

	loginMinAvailable   *intstr.IntOrString
	loginMaxUnavailable *intstr.IntOrString
	loginAnnKey         string
	loginAnnVal         string
}

func TestPodDisruptionBudgetMatrix(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)

	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	intOrStringPtr := func(value intstr.IntOrString) *intstr.IntOrString { return &value }

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
		expected  pdbExpected
	}{
		{
			name: "both-enabled-minAvailable-int",
			setValues: map[string]string{
				"pdb.enabled":      "true",
				"pdb.minAvailable": "2",

				"login.enabled":          "true",
				"login.pdb.enabled":      "true",
				"login.pdb.minAvailable": "1",
			},
			expected: pdbExpected{
				zitadelEnabled:      true,
				loginEnabled:        true,
				zitadelMinAvailable: intOrStringPtr(intstr.FromInt32(2)),
				loginMinAvailable:   intOrStringPtr(intstr.FromInt32(1)),
			},
		},
		{
			name: "both-enabled-minAvailable-percentage",
			setValues: map[string]string{
				"pdb.enabled":      "true",
				"pdb.minAvailable": "50%",

				"login.enabled":          "true",
				"login.pdb.enabled":      "true",
				"login.pdb.minAvailable": "75%",
			},
			expected: pdbExpected{
				zitadelEnabled:      true,
				loginEnabled:        true,
				zitadelMinAvailable: intOrStringPtr(intstr.FromString("50%")),
				loginMinAvailable:   intOrStringPtr(intstr.FromString("75%")),
			},
		},
		{
			name: "both-enabled-maxUnavailable-int",
			setValues: map[string]string{
				"pdb.enabled":        "true",
				"pdb.maxUnavailable": "1",

				"login.enabled":            "true",
				"login.pdb.enabled":        "true",
				"login.pdb.maxUnavailable": "2",
			},
			expected: pdbExpected{
				zitadelEnabled:        true,
				loginEnabled:          true,
				zitadelMaxUnavailable: intOrStringPtr(intstr.FromInt32(1)),
				loginMaxUnavailable:   intOrStringPtr(intstr.FromInt32(2)),
			},
		},
		{
			name: "both-enabled-maxUnavailable-percentage",
			setValues: map[string]string{
				"pdb.enabled":        "true",
				"pdb.maxUnavailable": "25%",

				"login.enabled":            "true",
				"login.pdb.enabled":        "true",
				"login.pdb.maxUnavailable": "33%",
			},
			expected: pdbExpected{
				zitadelEnabled:        true,
				loginEnabled:          true,
				zitadelMaxUnavailable: intOrStringPtr(intstr.FromString("25%")),
				loginMaxUnavailable:   intOrStringPtr(intstr.FromString("33%")),
			},
		},
		{
			name: "zitadel-enabled-login-disabled",
			setValues: map[string]string{
				"pdb.enabled":      "true",
				"pdb.minAvailable": "1",

				"login.enabled":     "true",
				"login.pdb.enabled": "false",
			},
			expected: pdbExpected{
				zitadelEnabled:      true,
				loginEnabled:        false,
				zitadelMinAvailable: intOrStringPtr(intstr.FromInt32(1)),
			},
		},
		{
			name: "both-enabled-with-annotations",
			setValues: map[string]string{
				"pdb.enabled":           "true",
				"pdb.minAvailable":      "1",
				"pdb.annotations.team":  "platform",
				"pdb.annotations.owner": "sre",

				"login.enabled":              "true",
				"login.pdb.enabled":          "true",
				"login.pdb.minAvailable":     "1",
				"login.pdb.annotations.team": "frontend",
			},
			expected: pdbExpected{
				zitadelEnabled:      true,
				loginEnabled:        true,
				zitadelMinAvailable: intOrStringPtr(intstr.FromInt32(1)),
				zitadelAnnKey:       "team",
				zitadelAnnVal:       "platform",
				loginMinAvailable:   intOrStringPtr(intstr.FromInt32(1)),
				loginAnnKey:         "team",
				loginAnnVal:         "frontend",
			},
		},
		{
			name: "both-disabled",
			setValues: map[string]string{
				"pdb.enabled": "false",

				"login.enabled":     "true",
				"login.pdb.enabled": "false",
			},
			expected: pdbExpected{
				zitadelEnabled: false,
				loginEnabled:   false,
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

				var expectedZitadelPDB *policyv1.PodDisruptionBudget
				if testCase.expected.zitadelEnabled {
					pdb := &policyv1.PodDisruptionBudget{
						Spec: policyv1.PodDisruptionBudgetSpec{
							MinAvailable:   testCase.expected.zitadelMinAvailable,
							MaxUnavailable: testCase.expected.zitadelMaxUnavailable,
						},
					}
					if testCase.expected.zitadelAnnKey != "" {
						pdb.Annotations = map[string]string{
							testCase.expected.zitadelAnnKey: testCase.expected.zitadelAnnVal,
						}
					}
					expectedZitadelPDB = pdb
				}
				assertPDB(t, env, releaseName, expectedZitadelPDB)

				var expectedLoginPDB *policyv1.PodDisruptionBudget
				if testCase.expected.loginEnabled {
					pdb := &policyv1.PodDisruptionBudget{
						Spec: policyv1.PodDisruptionBudgetSpec{
							MinAvailable:   testCase.expected.loginMinAvailable,
							MaxUnavailable: testCase.expected.loginMaxUnavailable,
						},
					}
					if testCase.expected.loginAnnKey != "" {
						pdb.Annotations = map[string]string{
							testCase.expected.loginAnnKey: testCase.expected.loginAnnVal,
						}
					}
					expectedLoginPDB = pdb
				}
				assertPDB(t, env, releaseName+"-login", expectedLoginPDB)
			})
		})
	}
}

func mustGetPDB(t *testing.T, env *support.Env, pdbName string) *policyv1.PodDisruptionBudget {
	var pdb *policyv1.PodDisruptionBudget
	require.Eventually(
		t,
		func() bool {
			retrievedPDB, err := env.Client.
				PolicyV1().
				PodDisruptionBudgets(env.Kube.Namespace).
				Get(context.Background(), pdbName, metav1.GetOptions{})
			if err != nil {
				return false
			}
			pdb = retrievedPDB
			return true
		},
		2*time.Minute,
		2*time.Second,
	)
	return pdb
}

func assertPDB(t *testing.T, env *support.Env, pdbName string, expected *policyv1.PodDisruptionBudget) {
	t.Helper()

	actualPDB, err := env.Client.
		PolicyV1().
		PodDisruptionBudgets(env.Kube.Namespace).
		Get(context.Background(), pdbName, metav1.GetOptions{})

	if expected == nil {
		require.True(t, apierrors.IsNotFound(err), "expected PDB %q to be NotFound, got: %v", pdbName, err)
		env.Logger.Logf(t, "✓ Verified PDB %s does not exist", pdbName)
		return
	}

	require.NoError(t, err, "failed to get PDB %s", pdbName)
	require.Equal(t, *expected, *actualPDB, "PDB mismatch for %s", pdbName)

	env.Logger.Logf(t, "✓ Verified PDB configuration for %s", pdbName)
}

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
