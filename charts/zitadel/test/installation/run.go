package installation

import (
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *ConfigurationTest) TestZITADELInstallation() {
	helm.AddRepo(s.T(), &helm.Options{}, s.crdbRepoName, s.crdbRepoURL)
	helm.Install(s.T(), &helm.Options{
		KubectlOptions: s.KubeOptions,
		SetValues:      s.crdbValues,
		Version:        s.crdbVersion,
	}, s.crdbChart, s.crdbRelease)
	helm.Install(s.T(), &helm.Options{
		KubectlOptions: s.KubeOptions,
		SetValues:      s.zitadelValues,
	}, s.zitadelChartPath, s.zitadelRelease)
	k8s.WaitUntilJobSucceed(s.T(), s.KubeOptions, "zitadel-test-init", 900, time.Second)
	k8s.WaitUntilJobSucceed(s.T(), s.KubeOptions, "zitadel-test-setup", 900, time.Second)
	pods := listPods(s.T(), 5, s.KubeOptions)
	s.awaitReadiness(pods)
	zitadelPods := make([]corev1.Pod, 0)
	for i := range pods {
		pod := pods[i]
		if name, ok := pod.GetObjectMeta().GetLabels()["app.kubernetes.io/name"]; ok && name == "zitadel" {
			zitadelPods = append(zitadelPods, pod)
		}
	}
	s.log.Logf(s.T(), "ZITADEL pods are ready")
	s.checkAccessibility(zitadelPods)
}

// listPods retries until all three start pods are returned from the kubeapi
func listPods(t *testing.T, try int, kubeOptions *k8s.KubectlOptions) []corev1.Pod {
	if try == 0 {
		t.Fatal("no trials left")
	}
	pods := k8s.ListPods(t, kubeOptions, metav1.ListOptions{LabelSelector: `app.kubernetes.io/instance=zitadel-test, app.kubernetes.io/component=start`})
	if len(pods) == 3 {
		return pods
	}
	time.Sleep(time.Second)
	return listPods(t, try-1, kubeOptions)
}
