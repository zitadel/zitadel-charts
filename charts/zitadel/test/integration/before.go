package integration

import "github.com/gruntwork-io/terratest/modules/k8s"

func (s *integrationTest) SetupTest() {
	if _, err := k8s.GetNamespaceE(s.T(), s.kubeOptions, s.namespace); err != nil {
		k8s.CreateNamespace(s.T(), s.kubeOptions, s.namespace)
	} else {
		s.log.Logf(s.T(), "Namespace: %s already exist!", s.namespace)
	}
}
