package integration

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func (s *integrationTest) TestZITADELEnd2End() {
	// given
	options := &helm.Options{
		KubectlOptions: s.kubeOptions,
		SetStrValues: map[string]string{
			"zitadel.masterkey":                           "x123456789012345678901234567891y",
			"zitadel.secretConfig.Database.User.Password": "xy",
			"zitadel.configmapConfig.ExternalDomain":      "test.domain",
		},
	}

	// when
	helm.Install(s.T(), options, s.chartPath, s.release)

	// then
	// await that all zitadel related pods become ready
	pods := k8s.ListPods(s.T(), s.kubeOptions, v1.ListOptions{LabelSelector: `app.kubernetes.io/instance=zitadel-test, app.kubernetes.io/component notin (init)`})

	for _, pod := range pods {
		k8s.WaitUntilPodAvailable(s.T(), s.kubeOptions, pod.Name, 180, time.Second)
	}

}
