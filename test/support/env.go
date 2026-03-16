// Package support provides generic Kubernetes and Helm test infrastructure
// for chart testing: ephemeral namespaces, resource getters, and template
// rendering helpers.
package support

import (
	"context"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// Env represents a per-test environment created by WithNamespace. It provides
// namespace-scoped kubectl options, a Kubernetes client, a timeout-scoped
// context, and a logger for consistent test output across test execution.
type Env struct {
	Ctx           context.Context
	Namespace     string
	Kube          *k8s.KubectlOptions
	Client        *kubernetes.Clientset
	DynamicClient dynamic.Interface
	Logger        *logger.Logger
}
