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
	testCtx, testCancel := context.WithCancelCause(context.Background())
	defer cancelTest(testCtx, testCancel, s.T())
	ctx, cancel := context.WithTimeout(testCtx, 30*time.Minute)
	defer cancel()
	s.T().Run("install", func(t *testing.T) {
		helm.Install(t, &helm.Options{
			KubectlOptions: s.KubeOptions,
			ValuesFiles:    s.zitadelValues,
			SetValues: map[string]string{
				"replicaCount":       "1",
				"login.replicaCount": "1",
				"pdb.enabled":        "true",
			},
		}, s.zitadelChartPath, s.zitadelRelease)
	})
	s.T().Run("init", func(t *testing.T) {
		k8s.WaitUntilJobSucceed(t, s.KubeOptions, "zitadel-test-init", 900, time.Second)
	})
	s.T().Run("setup", func(t *testing.T) {
		k8s.WaitUntilJobSucceed(t, s.KubeOptions, "zitadel-test-setup", 900, time.Second)
	})
	s.T().Run("readiness", func(t *testing.T) {
		pods := listPods(t, 5, s.KubeOptions)
		s.awaitReadiness(t, pods)
	})
	s.T().Run("accessibility", func(t *testing.T) {
		accessCtx, accessCancel := context.WithCancelCause(ctx)
		defer cancelTest(accessCtx, accessCancel, t)
		s.checkAccessibility(accessCtx, t)
	})
	s.T().Run("login", func(t *testing.T) {
		loginCtx, loginCancel := context.WithCancelCause(ctx)
		defer cancelTest(loginCtx, loginCancel, t)
		s.login(loginCtx, t)
	})
}

// listPods retries until all three start pods are returned from the kubeapi
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
