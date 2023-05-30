package installation

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/jinzhu/copier"
)

func (s *ConfigurationTest) SetupTest() {
	clusterKubectl := new(k8s.KubectlOptions)
	if err := copier.Copy(clusterKubectl, s.KubeOptions); err != nil {
		s.T().Fatal(err)
	}
	clusterKubectl.Namespace = ""
	// ignore error
	_ = k8s.KubectlDeleteFromStringE(s.T(), clusterKubectl, `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: crdb
`)
	// ignore error
	_ = k8s.KubectlDeleteFromStringE(s.T(), clusterKubectl, `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crdb
`)
	if _, err := k8s.GetNamespaceE(s.T(), s.KubeOptions, s.KubeOptions.Namespace); err != nil {
		k8s.CreateNamespace(s.T(), s.KubeOptions, s.KubeOptions.Namespace)
	} else {
		s.log.Logf(s.T(), "Namespace: %s already exist!", s.KubeOptions.Namespace)
	}
	if s.beforeFunc == nil {
		return
	}
}
