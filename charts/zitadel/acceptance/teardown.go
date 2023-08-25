package acceptance

import (
	"os"

	"github.com/gruntwork-io/terratest/modules/k8s"
)

func (s *ConfigurationTest) TearDownTest() {
	if _, exists := os.LookupEnv("GITHUB_SHA"); exists {
		s.log.Logf(s.T(), "Not running cleanup tasks, as tests were run on a throwaway runner")
		return
	}
	if !s.T().Failed() {
		k8s.DeleteNamespace(s.T(), s.KubeOptions, s.KubeOptions.Namespace)
	} else {
		s.log.Logf(s.T(), "Test failed on namespace %s. Omitting cleanup.", s.KubeOptions.Namespace)
	}
}
