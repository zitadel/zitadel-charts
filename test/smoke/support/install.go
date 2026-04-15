package support

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	testsupport "github.com/zitadel/zitadel-charts/test/support"
)

// ChartPath returns the absolute path to the Zitadel Helm chart.
func ChartPath(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine caller info for chart path resolution")
	}
	absPath, err := filepath.Abs(filepath.Join(filepath.Dir(filename), "..", "..", "..", "charts", "zitadel"))
	require.NoError(t, err)
	return absPath
}

// InstallZitadel installs the Zitadel chart with PostgreSQL and standard
// configuration. It handles: WithPostgres, commonSetValues, MakeRelease,
// mergedSetValues, helm.Options, helm.UpgradeE, and log dumping on failure.
// Returns the generated release name.
func InstallZitadel(t *testing.T, env *testsupport.Env, testName string, setValues map[string]string) string {
	t.Helper()

	chartPath := ChartPath(t)

	env.Logger.Logf(t, "namespace %q created; installing PostgreSQL…", env.Namespace)
	WithPostgres(t, env)

	uniqueDomain := fmt.Sprintf("%s.test.local", env.Namespace)
	commonSetValues := map[string]string{
		"zitadel.masterkey":                      "x123456789012345678901234567891y",
		"zitadel.configmapConfig.ExternalDomain": uniqueDomain,
		"zitadel.configmapConfig.ExternalPort":   "443",
		"zitadel.configmapConfig.TLS.Enabled":    "false",
		"ingress.enabled":                        "true",
		"login.ingress.enabled":                  "true",
	}
	// Hardcoded discrete Database.Postgres.* fields put the chart in legacy
	// "configmap mode". Skip them if the caller is exercising DSN mode (either
	// via an env var ZITADEL_DATABASE_POSTGRES_DSN or by enabling the bundled
	// postgresql subchart) so the chart's DSN code paths are actually tested.
	if !setValuesUseDSNMode(setValues) {
		commonSetValues["zitadel.configmapConfig.Database.Postgres.Host"] = "db-postgresql"
		commonSetValues["zitadel.configmapConfig.Database.Postgres.Port"] = "5432"
		commonSetValues["zitadel.configmapConfig.Database.Postgres.Database"] = "zitadel"
		commonSetValues["zitadel.configmapConfig.Database.Postgres.MaxOpenConns"] = "20"
		commonSetValues["zitadel.configmapConfig.Database.Postgres.MaxIdleConns"] = "10"
		commonSetValues["zitadel.configmapConfig.Database.Postgres.MaxConnLifetime"] = "30m"
		commonSetValues["zitadel.configmapConfig.Database.Postgres.MaxConnIdleTime"] = "5m"
		commonSetValues["zitadel.configmapConfig.Database.Postgres.User.Username"] = "postgres"
		commonSetValues["zitadel.configmapConfig.Database.Postgres.User.SSL.Mode"] = "disable"
		commonSetValues["zitadel.configmapConfig.Database.Postgres.Admin.Username"] = "postgres"
		commonSetValues["zitadel.configmapConfig.Database.Postgres.Admin.SSL.Mode"] = "disable"
	}

	releaseName := env.MakeRelease("zitadel-test", testName)

	mergedSetValues := make(map[string]string)
	for key, value := range commonSetValues {
		mergedSetValues[key] = value
	}
	for key, value := range setValues {
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

	return releaseName
}

// setValuesUseDSNMode returns true if the caller's setValues opt into DSN mode,
// either by setting an env entry whose name is ZITADEL_DATABASE_POSTGRES_DSN
// or by enabling the bundled postgresql subchart. Helm --set encodes list
// entries as e.g. "env[0].name=ZITADEL_DATABASE_POSTGRES_DSN" so we have to
// scan the keys for that pattern.
func setValuesUseDSNMode(setValues map[string]string) bool {
	if setValues["postgresql.enabled"] == "true" {
		return true
	}
	for key, value := range setValues {
		if strings.HasPrefix(key, "env[") && strings.HasSuffix(key, "].name") &&
			value == "ZITADEL_DATABASE_POSTGRES_DSN" {
			return true
		}
	}
	return false
}

func dumpSetupAndInitJobLogs(t *testing.T, env *testsupport.Env, releaseName string) {
	namespace := env.Kube.Namespace
	jobNames := []string{fmt.Sprintf("%s-setup", releaseName), fmt.Sprintf("%s-init", releaseName)}

	for _, jobName := range jobNames {
		labelSelector := fmt.Sprintf("job-name=%s", jobName)
		pods := listPods(t, env, labelSelector)

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

func listPods(t *testing.T, env *testsupport.Env, labelSelector string) []corev1.Pod {
	podList, err := env.Client.CoreV1().Pods(env.Kube.Namespace).List(
		env.Ctx,
		metav1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		env.Logger.Logf(t, "warn: list pods selector=%q: %v", labelSelector, err)
		return nil
	}
	return podList.Items
}
