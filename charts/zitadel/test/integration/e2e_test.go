package integration

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *integrationTest) TestZITADELEnd2End() {
	// given
	options := &helm.Options{
		KubectlOptions: s.kubeOptions,
		SetStrValues: map[string]string{
			"zitadel.masterkey":                           "x123456789012345678901234567891y",
			"zitadel.secretConfig.Database.User.Password": "xy",
			"zitadel.configmapConfig.ExternalPort":        "8080",
		},
	}

	// when
	helm.Install(s.T(), options, s.chartPath, s.release)

	// then
	// await that all zitadel related pods become ready
	pods := k8s.ListPods(s.T(), s.kubeOptions, metav1.ListOptions{LabelSelector: `app.kubernetes.io/instance=zitadel-test, app.kubernetes.io/component notin (init)`})
	s.awaitAvailability(pods)
	zitadelPods := make([]corev1.Pod, 0)
	for i := range pods {
		pod := pods[i]
		if name, ok := pod.GetObjectMeta().GetLabels()["app.kubernetes.io/name"]; ok && name == "zitadel" {
			zitadelPods = append(zitadelPods, pod)
		}
	}
	s.awaitListening(zitadelPods)
	s.awaitAccessibility(zitadelPods)
}
