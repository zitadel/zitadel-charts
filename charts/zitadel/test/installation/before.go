package installation

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/jinzhu/copier"
)

func (s *configurationTest) SetupTest() {

	clusterKubectl := new(k8s.KubectlOptions)
	if err := copier.Copy(clusterKubectl, s.kubeOptions); err != nil {
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

	if _, err := k8s.GetNamespaceE(s.T(), s.kubeOptions, s.kubeOptions.Namespace); err != nil {
		k8s.CreateNamespace(s.T(), s.kubeOptions, s.kubeOptions.Namespace)
	} else {
		s.log.Logf(s.T(), "Namespace: %s already exist!", s.kubeOptions.Namespace)
	}

	if s.beforeFunc == nil {
		return
	}

	clientset, err := k8s.GetKubernetesClientFromOptionsE(s.T(), s.kubeOptions)
	if err != nil {
		s.T().Fatal(err)
	}

	if err = s.beforeFunc(s.context, s.kubeOptions.Namespace, clientset); err != nil {
		s.T().Fatal(err)
	}
}
