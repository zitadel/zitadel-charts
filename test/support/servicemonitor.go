package support

import (
	"testing"

	"github.com/stretchr/testify/require"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var serviceMonitorGVR = schema.GroupVersionResource{
	Group:    "monitoring.coreos.com",
	Version:  "v1",
	Resource: "servicemonitors",
}

// GetServiceMonitor fetches a ServiceMonitor by name, failing the test on error.
func (env *Env) GetServiceMonitor(t *testing.T, name string) *monitoringv1.ServiceMonitor {
	t.Helper()
	obj, err := env.DynamicClient.Resource(serviceMonitorGVR).Namespace(env.Namespace).Get(env.Ctx, name, metav1.GetOptions{})
	require.NoError(t, err, "failed to get ServiceMonitor %s", name)
	var sm monitoringv1.ServiceMonitor
	require.NoError(t, runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &sm))
	return &sm
}

// GetServiceMonitorE fetches a ServiceMonitor by name, returning the error for non-existence checks.
func (env *Env) GetServiceMonitorE(t *testing.T, name string) (*monitoringv1.ServiceMonitor, error) {
	t.Helper()
	obj, err := env.DynamicClient.Resource(serviceMonitorGVR).Namespace(env.Namespace).Get(env.Ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	var sm monitoringv1.ServiceMonitor
	if convErr := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &sm); convErr != nil {
		return nil, convErr
	}
	return &sm, nil
}
