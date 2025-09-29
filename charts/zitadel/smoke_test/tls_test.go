package smoke_test_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zitadel/zitadel-charts/charts/zitadel/smoke_test/support"
)

func TestTLSWithUserProvidedCertificate(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)

	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	support.WithNamespace(t, cluster, func(env *support.Env) {
		env.Logger.Logf(t, "namespace %q created; installing PostgreSQL…", env.Namespace)
		support.WithPostgres(t, env)

		support.WithTLSSecret(t, env, "wildcard-tls", "*.dev.mrida.ng", "dev.mrida.ng")

		releaseName := env.MakeRelease("zitadel-tls", "external-cert")

		helmOptions := &helm.Options{
			KubectlOptions: env.Kube,
			SetValues: map[string]string{
				"zitadel.masterkey":                                         "x123456789012345678901234567891y",
				"zitadel.configmapConfig.ExternalDomain":                    "zitadel.dev.mrida.ng",
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
				"zitadel.serverSslCrtSecret":                                "wildcard-tls",
				"login.serverSslCrtSecret":                                  "wildcard-tls",
				"ingress.enabled":                                           "false",
				"login.ingress.enabled":                                     "false",
				"service.appProtocol":                                       "kubernetes.io/https",
				"service.port":                                              "443",
			},
			SetJsonValues: map[string]string{
				"login.env": `[{"name":"NODE_TLS_REJECT_UNAUTHORIZED","value":"0"}]`,
			},
			ExtraArgs: map[string][]string{
				"upgrade": {"--install", "--wait", "--timeout", "30m"},
			},
		}

		if err := helm.UpgradeE(t, helmOptions, chartPath, releaseName); err != nil {
			dumpSetupAndInitJobLogs(t, env, releaseName)
			require.NoError(t, err)
		}

		waitForDeploymentReady(t, env, releaseName)
		waitForDeploymentReady(t, env, releaseName+"-login")

		assertHTTPSAccessible(t, env, releaseName, 443)
		assertHTTPSAccessible(t, env, releaseName+"-login", 443)
	})
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

func waitForDeploymentReady(t *testing.T, env *support.Env, deploymentName string) {
	t.Helper()

	require.Eventually(t, func() bool {
		deployment, err := k8s.GetDeploymentE(t, env.Kube, deploymentName)
		if err != nil {
			return false
		}
		return deployment.Status.ReadyReplicas > 0 &&
			deployment.Status.ReadyReplicas == deployment.Status.Replicas
	}, 5*time.Minute, 5*time.Second)
}

func assertHTTPSAccessible(t *testing.T, env *support.Env, serviceName string, port int) {
	t.Helper()

	tunnel := k8s.NewTunnel(env.Kube, k8s.ResourceTypeService, serviceName, 0, port)
	defer tunnel.Close()
	tunnel.ForwardPort(t)

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	require.Eventually(t, func() bool {
		resp, err := client.Get(fmt.Sprintf("https://%s/healthz", tunnel.Endpoint()))
		if err != nil {
			env.Logger.Logf(t, "health check failed for %s: %v", serviceName, err)
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 2*time.Minute, 5*time.Second)
}
