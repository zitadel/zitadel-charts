package support

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetDeployment fetches a Deployment by name, failing the test on error.
func (env *Env) GetDeployment(t *testing.T, name string) *appsv1.Deployment {
	t.Helper()
	obj, err := env.Client.AppsV1().Deployments(env.Kube.Namespace).Get(context.Background(), name, metav1.GetOptions{})
	require.NoError(t, err, "failed to get Deployment %s", name)
	return obj
}

// GetService fetches a Service by name, failing the test on error.
func (env *Env) GetService(t *testing.T, name string) *corev1.Service {
	t.Helper()
	obj, err := env.Client.CoreV1().Services(env.Kube.Namespace).Get(context.Background(), name, metav1.GetOptions{})
	require.NoError(t, err, "failed to get Service %s", name)
	return obj
}

// GetServiceAccount fetches a ServiceAccount by name, failing the test on error.
func (env *Env) GetServiceAccount(t *testing.T, name string) *corev1.ServiceAccount {
	t.Helper()
	obj, err := env.Client.CoreV1().ServiceAccounts(env.Kube.Namespace).Get(context.Background(), name, metav1.GetOptions{})
	require.NoError(t, err, "failed to get ServiceAccount %s", name)
	return obj
}

// GetConfigMap fetches a ConfigMap by name, failing the test on error.
func (env *Env) GetConfigMap(t *testing.T, name string) *corev1.ConfigMap {
	t.Helper()
	obj, err := env.Client.CoreV1().ConfigMaps(env.Kube.Namespace).Get(context.Background(), name, metav1.GetOptions{})
	require.NoError(t, err, "failed to get ConfigMap %s", name)
	return obj
}

// GetSecret fetches a Secret by name, failing the test on error.
func (env *Env) GetSecret(t *testing.T, name string) *corev1.Secret {
	t.Helper()
	obj, err := env.Client.CoreV1().Secrets(env.Kube.Namespace).Get(context.Background(), name, metav1.GetOptions{})
	require.NoError(t, err, "failed to get Secret %s", name)
	return obj
}

// GetSecretE fetches a Secret by name, returning the error for non-existence checks.
func (env *Env) GetSecretE(t *testing.T, name string) (*corev1.Secret, error) {
	t.Helper()
	return env.Client.CoreV1().Secrets(env.Kube.Namespace).Get(context.Background(), name, metav1.GetOptions{})
}

// GetPDB fetches a PodDisruptionBudget by name, failing the test on error.
func (env *Env) GetPDB(t *testing.T, name string) *policyv1.PodDisruptionBudget {
	t.Helper()
	obj, err := env.Client.PolicyV1().PodDisruptionBudgets(env.Kube.Namespace).Get(context.Background(), name, metav1.GetOptions{})
	require.NoError(t, err, "failed to get PDB %s", name)
	return obj
}

// GetHPA fetches a HorizontalPodAutoscaler by name, failing the test on error.
func (env *Env) GetHPA(t *testing.T, name string) *autoscalingv2.HorizontalPodAutoscaler {
	t.Helper()
	obj, err := env.Client.AutoscalingV2().HorizontalPodAutoscalers(env.Kube.Namespace).Get(context.Background(), name, metav1.GetOptions{})
	require.NoError(t, err, "failed to get HPA %s", name)
	return obj
}

// GetIngress fetches an Ingress by name, failing the test on error.
func (env *Env) GetIngress(t *testing.T, name string) *networkingv1.Ingress {
	t.Helper()
	obj, err := env.Client.NetworkingV1().Ingresses(env.Kube.Namespace).Get(context.Background(), name, metav1.GetOptions{})
	require.NoError(t, err, "failed to get Ingress %s", name)
	return obj
}

// GetRole fetches a Role by name, failing the test on error.
func (env *Env) GetRole(t *testing.T, name string) *rbacv1.Role {
	t.Helper()
	obj, err := env.Client.RbacV1().Roles(env.Kube.Namespace).Get(context.Background(), name, metav1.GetOptions{})
	require.NoError(t, err, "failed to get Role %s", name)
	return obj
}

// GetRoleBinding fetches a RoleBinding by name, failing the test on error.
func (env *Env) GetRoleBinding(t *testing.T, name string) *rbacv1.RoleBinding {
	t.Helper()
	obj, err := env.Client.RbacV1().RoleBindings(env.Kube.Namespace).Get(context.Background(), name, metav1.GetOptions{})
	require.NoError(t, err, "failed to get RoleBinding %s", name)
	return obj
}
