package installation

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/jinzhu/copier"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (s *ConfigurationTest) SetupTest() {
	clusterKubectl := new(k8s.KubectlOptions)
	t := s.T()
	if err := copier.Copy(clusterKubectl, s.KubeOptions); err != nil {
		t.Fatal(err)
	}
	clusterKubectl.Namespace = ""
	if err := k8s.KubectlDeleteFromStringE(t, clusterKubectl, `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: crdb
`); err != nil && !errors.IsNotFound(err) {
		t.Fatal(err)
	}
	if err := k8s.KubectlDeleteFromStringE(t, clusterKubectl, `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crdb
`); err != nil && !errors.IsNotFound(err) {
		t.Fatal(err)
	}
	_, err := k8s.GetNamespaceE(t, s.KubeOptions, s.KubeOptions.Namespace)
	isNotFound := errors.IsNotFound(err)
	if err != nil && !isNotFound {
		t.Fatal(err)
	}
	if isNotFound {
		k8s.CreateNamespace(t, s.KubeOptions, s.KubeOptions.Namespace)
		return
	}
	s.log.Logf(s.T(), "Namespace: %s already exist!", s.KubeOptions.Namespace)
}
