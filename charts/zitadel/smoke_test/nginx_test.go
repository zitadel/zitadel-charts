package smoke_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zitadel/zitadel-charts/charts/zitadel/smoke_test/support"
)

func TestNginxConfiguration(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)

	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	testCases := []struct {
		name                 string
		zitadelNginxImage    string
		loginNginxImage      string
		expectedZitadelNginx string
		expectedLoginNginx   string
	}{
		{
			name:                 "default-nginx-images",
			zitadelNginxImage:    "",
			loginNginxImage:      "",
			expectedZitadelNginx: "nginx:1.27-alpine",
			expectedLoginNginx:   "nginx:1.27-alpine",
		},
		{
			name:                 "custom-nginx-images",
			zitadelNginxImage:    "nginx:1.26-alpine",
			loginNginxImage:      "nginx:1.25-alpine",
			expectedZitadelNginx: "nginx:1.26-alpine",
			expectedLoginNginx:   "nginx:1.25-alpine",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			support.WithNamespace(t, cluster, func(env *support.Env) {
				env.Logger.Logf(t, "namespace %q created; installing PostgreSQL…", env.Namespace)
				support.WithPostgres(t, env)

				releaseName := env.MakeRelease("zitadel-nginx", testCase.name)

				setValues := map[string]string{
					"zitadel.masterkey":                                         "x123456789012345678901234567891y",
					"zitadel.configmapConfig.ExternalDomain":                    "nginx-test.127.0.0.1.sslip.io",
					"zitadel.configmapConfig.ExternalPort":                      "443",
					"zitadel.configmapConfig.ExternalSecure":                    "true",
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
					"zitadel.selfSignedCert.enabled":                            "true",
					"login.selfSignedCert.enabled":                              "true",
					"ingress.enabled":                                           "false",
					"login.ingress.enabled":                                     "false",
					"service.appProtocol":                                       "kubernetes.io/https",
					"service.port":                                              "443",
				}

				if testCase.zitadelNginxImage != "" {
					setValues["zitadel.nginx.image"] = testCase.zitadelNginxImage
				}
				if testCase.loginNginxImage != "" {
					setValues["login.nginx.image"] = testCase.loginNginxImage
				}

				helmOptions := &helm.Options{
					KubectlOptions: env.Kube,
					SetValues:      setValues,
					ExtraArgs: map[string][]string{
						"upgrade": {"--install", "--wait", "--timeout", "30m"},
					},
				}

				require.NoError(t, helm.UpgradeE(t, helmOptions, chartPath, releaseName))

				assertNginxImageAndValidateLogs(t, env, releaseName, testCase.expectedZitadelNginx)
				assertNginxImageAndValidateLogs(t, env, releaseName+"-login", testCase.expectedLoginNginx)
			})
		})
	}
}

func assertNginxImageAndValidateLogs(t *testing.T, env *support.Env, deploymentName string, expectedImage string) {
	t.Helper()

	deployment, err := k8s.GetDeploymentE(t, env.Kube, deploymentName)
	require.NoError(t, err)

	var nginxContainer *corev1.Container
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == "zitadel-reverse-proxy" {
			nginxContainer = &container
			break
		}
	}

	require.NotNil(t, nginxContainer, "nginx reverse-proxy container not found in deployment %s", deploymentName)
	require.Equal(t, expectedImage, nginxContainer.Image, "nginx image mismatch for deployment %s", deploymentName)

	env.Logger.Logf(t, "✓ Verified nginx image for %s: %s", deploymentName, nginxContainer.Image)

	validateNginxLogs(t, env, deploymentName)
}

func validateNginxLogs(t *testing.T, env *support.Env, deploymentName string) {
	t.Helper()

	labelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", deploymentName)

	podList, err := env.Client.CoreV1().Pods(env.Kube.Namespace).List(
		context.Background(),
		metav1.ListOptions{LabelSelector: labelSelector},
	)

	if err != nil {
		env.Logger.Logf(t, "failed to list pods for %s: %v", deploymentName, err)
		return
	}

	for _, pod := range podList.Items {
		if strings.Contains(pod.Name, deploymentName) {
			validateContainerLogs(t, env, pod, "zitadel-reverse-proxy")
		}
	}
}

func validateContainerLogs(t *testing.T, env *support.Env, pod corev1.Pod, containerName string) {
	t.Helper()

	env.Logger.Logf(t, "==== Validating logs for pod=%s container=%s ====", pod.Name, containerName)

	logOutput, err := k8s.RunKubectlAndGetOutputE(
		t,
		env.Kube,
		"logs",
		pod.Name,
		"-c", containerName,
		"--tail=100",
	)

	if err != nil {
		env.Logger.Logf(t, "failed to get logs: %v", err)
		return
	}

	env.Logger.Logf(t, "%s", logOutput)

	lines := strings.Split(logOutput, "\n")
	var unexpectedWarnings []string

	for _, line := range lines {
		if strings.Contains(line, "[warn]") {
			if strings.Contains(line, `the "user" directive makes sense only if the master process runs with super-user privileges`) {
				continue
			}
			unexpectedWarnings = append(unexpectedWarnings, line)
		}
	}

	if len(unexpectedWarnings) > 0 {
		env.Logger.Logf(t, "❌ Found unexpected nginx warnings in pod=%s:", pod.Name)
		for _, warning := range unexpectedWarnings {
			env.Logger.Logf(t, "  %s", warning)
		}
		require.Fail(t, fmt.Sprintf("unexpected nginx warnings found in pod %s", pod.Name))
	} else {
		env.Logger.Logf(t, "✓ No unexpected nginx warnings found in pod=%s", pod.Name)
	}

	env.Logger.Logf(t, "==== End validation for pod=%s container=%s ====", pod.Name, containerName)
}
