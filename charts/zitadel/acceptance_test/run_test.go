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

func (s *ConfigurationTest) TestZitadelInstallation() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	t := s.T()
	if !t.Run("install", func(t *testing.T) {
		helm.Install(t, &helm.Options{
			KubectlOptions: s.KubeOptions,
			ValuesFiles:    s.zitadelValues,
			SetValues: map[string]string{
				"replicaCount":       "1",
				"login.replicaCount": "1",
				"pdb.enabled":        "true",
			},
		}, s.zitadelChartPath, s.zitadelRelease)
	}) {
		t.FailNow()
	}
	if !t.Run("init", func(t *testing.T) {
		k8s.WaitUntilJobSucceed(t, s.KubeOptions, "zitadel-test-init", 900, time.Second)
	}) {
		t.FailNow()
	}
	if !t.Run("setup", func(t *testing.T) {
		k8s.WaitUntilJobSucceed(t, s.KubeOptions, "zitadel-test-setup", 900, time.Second)
	}) {
		t.FailNow()
	}
	if !t.Run("readiness", func(t *testing.T) {
		pods := listPods(t, 5, s.KubeOptions)
		s.awaitReadiness(t, pods)
	}) {
		t.FailNow()
	}
	if !t.Run("login", func(t *testing.T) {
		s.login(ctx, t)
	}) {
		t.FailNow()
	}
	if !t.Run("grpc", func(t *testing.T) {
		assertGRPCWorks(ctx, t, s, "iam-admin")
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
