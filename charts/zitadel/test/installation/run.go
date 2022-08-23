package installation

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *configurationTest) TestZITADELInstallation() {
	// given
	options := s.options

	// when
	helm.Install(s.T(), options, s.chartPath, s.release)

	// then
	// await that all zitadel related pods become ready
	pods := k8s.ListPods(s.T(), s.options.KubectlOptions, metav1.ListOptions{LabelSelector: `app.kubernetes.io/instance=zitadel-test, app.kubernetes.io/component notin (init)`})
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
