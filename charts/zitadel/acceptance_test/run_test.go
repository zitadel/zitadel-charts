package acceptance_test

import (
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *ConfigurationTest) TestZITADELInstallation() {
	helm.Install(s.T(), &helm.Options{
		KubectlOptions: s.KubeOptions,
		ValuesFiles:    s.zitadelValues,
		SetValues: map[string]string{
			"replicaCount": "1",
			"pdb.enabled":  "true",
		},
	}, s.zitadelChartPath, s.zitadelRelease)
	k8s.WaitUntilJobSucceed(s.T(), s.KubeOptions, "zitadel-test-init", 900, time.Second)
	k8s.WaitUntilJobSucceed(s.T(), s.KubeOptions, "zitadel-test-setup", 900, time.Second)
	pods := listPods(s.T(), 5, s.KubeOptions)
	s.awaitReadiness(pods)
	s.checkAccessibility()
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
