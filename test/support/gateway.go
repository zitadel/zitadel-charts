package support

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zitadel/zitadel-charts/test/assert"
	errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var httpRouteGVR = schema.GroupVersionResource{
	Group:    "gateway.networking.k8s.io",
	Version:  "v1",
	Resource: "httproutes",
}

var grpcRouteGVR = schema.GroupVersionResource{
	Group:    "gateway.networking.k8s.io",
	Version:  "v1",
	Resource: "grpcroutes",
}

// GetHTTPRoute fetches an HTTPRoute by name, failing the test on error.
func (env *Env) GetHTTPRoute(t *testing.T, name string) *gatewayv1.HTTPRoute {
	t.Helper()
	obj, err := env.DynamicClient.Resource(httpRouteGVR).Namespace(env.Namespace).Get(env.Ctx, name, metav1.GetOptions{})
	require.NoError(t, err, "failed to get HTTPRoute %s", name)
	var route gatewayv1.HTTPRoute
	require.NoError(t, runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &route))
	return &route
}

// GetHTTPRouteE fetches an HTTPRoute by name, returning the error for non-existence checks.
func (env *Env) GetHTTPRouteE(t *testing.T, name string) (*gatewayv1.HTTPRoute, error) {
	t.Helper()
	obj, err := env.DynamicClient.Resource(httpRouteGVR).Namespace(env.Namespace).Get(env.Ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	var route gatewayv1.HTTPRoute
	if convErr := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &route); convErr != nil {
		return nil, convErr
	}
	return &route, nil
}

// GetGRPCRoute fetches a GRPCRoute by name, failing the test on error.
func (env *Env) GetGRPCRoute(t *testing.T, name string) *gatewayv1.GRPCRoute {
	t.Helper()
	obj, err := env.DynamicClient.Resource(grpcRouteGVR).Namespace(env.Namespace).Get(env.Ctx, name, metav1.GetOptions{})
	require.NoError(t, err, "failed to get GRPCRoute %s", name)
	var route gatewayv1.GRPCRoute
	require.NoError(t, runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &route))
	return &route
}

// GetGRPCRouteE fetches a GRPCRoute by name, returning the error for non-existence checks.
func (env *Env) GetGRPCRouteE(t *testing.T, name string) (*gatewayv1.GRPCRoute, error) {
	t.Helper()
	obj, err := env.DynamicClient.Resource(grpcRouteGVR).Namespace(env.Namespace).Get(env.Ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	var route gatewayv1.GRPCRoute
	if convErr := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &route); convErr != nil {
		return nil, convErr
	}
	return &route, nil
}

// assertPartialFallback handles assertion types not covered by the generated
// type switch in zz_generated.go. Gateway API types live outside
// kubernetes.Interface so supportgen cannot auto-generate cases for them.
func (env *Env) assertPartialFallback(t *testing.T, name string, assertion assert.Assertable) {
	t.Helper()
	switch a := assertion.(type) {
	case assert.HTTPRouteAssertion:
		assert.AssertPartial(t, env.GetHTTPRoute(t, name), a, name)
	case assert.GRPCRouteAssertion:
		assert.AssertPartial(t, env.GetGRPCRoute(t, name), a, name)
	default:
		t.Fatalf("env.AssertPartial: unsupported assertion type %T", assertion)
	}
}

// assertNoneFallback handles assertion types not covered by the generated
// type switch in zz_generated.go.
func (env *Env) assertNoneFallback(t *testing.T, name string, assertion assert.Assertable) {
	t.Helper()
	switch assertion.(type) {
	case assert.HTTPRouteAssertion, *assert.HTTPRouteAssertion:
		_, err := env.GetHTTPRouteE(t, name)
		require.True(t, errors.IsNotFound(err), "HTTPRoute %q should not exist (err: %v)", name, err)
	case assert.GRPCRouteAssertion, *assert.GRPCRouteAssertion:
		_, err := env.GetGRPCRouteE(t, name)
		require.True(t, errors.IsNotFound(err), "GRPCRoute %q should not exist (err: %v)", name, err)
	default:
		t.Fatalf("env.AssertNone: unsupported assertion type %T", assertion)
	}
}
