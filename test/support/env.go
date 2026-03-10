// Package support provides generic Kubernetes and Helm test infrastructure
// for chart testing: cluster connections, ephemeral namespaces, resource
// getters, and template rendering helpers.
package support

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
)

// Cluster holds a single shared connection to the Kubernetes cluster. It
// contains the configuration path, context name, and initialized client for
// reuse across multiple test scenarios.
type Cluster struct {
	ConfigPath  string
	ContextName string
	Client      *kubernetes.Clientset
}

// Env represents a per-test environment created by WithNamespace. It provides
// namespace-scoped kubectl options, a Kubernetes client, and a logger for
// consistent test output across test execution.
type Env struct {
	Namespace string
	Kube      *k8s.KubectlOptions
	Client    *kubernetes.Clientset
	Logger    *logger.Logger
}

// ConnectCluster establishes a single shared client connection to the
// Kubernetes cluster using the current KUBECONFIG and context. This function
// should be called once at the start of your top-level test and the returned
// cluster connection reused across subtests for efficiency.
func ConnectCluster(testing *testing.T) *Cluster {
	testing.Helper()

	baseOptions := k8s.NewKubectlOptions("", "", "default")

	client, err := k8s.GetKubernetesClientFromOptionsE(testing, baseOptions)
	require.NoError(testing, err)

	return &Cluster{
		ConfigPath:  baseOptions.ConfigPath,
		ContextName: baseOptions.ContextName,
		Client:      client,
	}
}
