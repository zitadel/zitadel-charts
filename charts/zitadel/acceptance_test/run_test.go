package acceptance_test

import (
	"context"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (suite *IntegrationSuite) TestZitadelInstallation() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	t := suite.T()
	if !t.Run("install", func(t *testing.T) {
		helm.Install(t, &helm.Options{
			KubectlOptions: suite.KubeOptions,
			ValuesFiles:    suite.ZitadelValues,
			SetValues: map[string]string{
				"replicaCount":       "1",
				"login.replicaCount": "1",
				"pdb.enabled":        "true",
			},
		}, suite.ZitadelChartPath, suite.ZitadelRelease)
	}) {
		t.FailNow()
	}
	if !t.Run("readiness", func(t *testing.T) {
		pods := listPods(t, 5, suite.KubeOptions)
		suite.awaitReadiness(t, pods)
	}) {
		t.FailNow()
	}
	if !t.Run("login", func(t *testing.T) {
		suite.login(ctx, t)
	}) {
		t.FailNow()
	}
	if !t.Run("grpc", func(t *testing.T) {
		assertGRPCWorks(ctx, t, suite, "iam-admin")
	}) {
		t.FailNow()
	}
}

// listPods retries until the expected pods are returned from the Kubernetes
// API. It attempts the specified number of times with a one second delay
// between attempts.
func listPods(t *testing.T, try int, kubeOptions *k8s.KubectlOptions) []corev1.Pod {
	if try == 0 {
		t.Fatal("no trials left")
	}
	pods := k8s.ListPods(t, kubeOptions, metav1.ListOptions{LabelSelector: `app.kubernetes.io/instance=zitadel-test, app.kubernetes.io/component=start`})
	if len(pods) == 1 {
		return pods
	}
	time.Sleep(time.Second)
	return listPods(t, try-1, kubeOptions)
}
