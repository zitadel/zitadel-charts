package installation

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
)

func (s *configurationTest) TearDownTest() {
	if !s.T().Failed() {
		k8s.DeleteNamespace(s.T(), s.kubeOptions, s.kubeOptions.Namespace)
	} else {
		s.log.Logf(s.T(), "Test failed on namespace %s. Omitting cleanup.", s.kubeOptions.Namespace)
	}
}
