package acceptance

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
)

func (s *ConfigurationTest) TearDownTest() {
	if !s.T().Failed() {
		k8s.DeleteNamespace(s.T(), s.KubeOptions, s.KubeOptions.Namespace)
	} else {
		s.log.Logf(s.T(), "Test failed on namespace %s. Omitting cleanup.", s.KubeOptions.Namespace)
	}
}
