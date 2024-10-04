package acceptance

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/jinzhu/copier"
	"k8s.io/apimachinery/pkg/api/errors"
	"strings"
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
`); err != nil && !isNotFoundFromKubectl(err) {
		t.Fatal(err)
	}
	if err := k8s.KubectlDeleteFromStringE(t, clusterKubectl, `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crdb
`); err != nil && !isNotFoundFromKubectl(err) {
		t.Fatal(err)
	}
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

func isNotFoundFromKubectl(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "not found")
}
