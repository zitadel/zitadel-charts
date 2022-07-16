package integration

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
)

func (s *integrationTest) TearDownTest() {
	if !s.T().Failed() {
		k8s.DeleteNamespace(s.T(), s.kubeOptions, s.namespace)
	} else {
		s.log.Logf(s.T(), "Test failed on namespace: %s!", s.namespace)
	}
	s.kubeOptions.Namespace = ""
	k8s.KubectlDeleteFromString(s.T(), s.kubeOptions, `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: crdb
`)
	k8s.KubectlDeleteFromString(s.T(), s.kubeOptions, `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crdb
`)
}
