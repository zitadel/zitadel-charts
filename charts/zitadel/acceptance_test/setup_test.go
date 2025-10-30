package acceptance_test

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (s *ConfigurationTest) SetupTest() {
	t := s.T()

	_, err := k8s.GetNamespaceE(t, s.KubeOptions, s.KubeOptions.Namespace)
	notFound := errors.IsNotFound(err)
	if err != nil && !notFound {
		t.Fatal(err)
	}
	if notFound {
		k8s.CreateNamespace(t, s.KubeOptions, s.KubeOptions.Namespace)
		return
	}
	s.log.Logf(s.T(), "Namespace: %s already exist!", s.KubeOptions.Namespace)
}
