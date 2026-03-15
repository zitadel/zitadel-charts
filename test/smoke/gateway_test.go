package smoke_test_test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zitadel/zitadel-charts/test/smoke/support"
)

var gatewayBaseValues = map[string]string{
	"zitadel.configmapConfig.ExternalDomain": "zitadel.example.local",
	"zitadel.masterkey":                      "01234567890123456789012345678901",
}

func gatewayOptions(extra map[string]string) *helm.Options {
	merged := make(map[string]string, len(gatewayBaseValues)+len(extra))
	for k, v := range gatewayBaseValues {
		merged[k] = v
	}
	for k, v := range extra {
		merged[k] = v
	}
	return &helm.Options{SetValues: merged}
}

func TestGatewayHTTPRouteLabels(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)
	releaseName := "gateway-httproute-labels"

	options := gatewayOptions(map[string]string{
		"gateway.httpRoute.enabled":            "true",
		"gateway.httpRoute.parentRefs[0].name": "my-gateway",
		"gateway.httpRoute.hostnames[0]":       "zitadel.example.local",
	})

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{"templates/httproute_zitadel.yaml"})

	var route map[string]interface{}
	helm.UnmarshalK8SYaml(t, rendered, &route)

	metadata := route["metadata"].(map[string]interface{})
	labels := metadata["labels"].(map[string]interface{})

	assert.Equal(t, "zitadel", labels["app.kubernetes.io/name"])
	assert.Equal(t, releaseName, labels["app.kubernetes.io/instance"])
	assert.Equal(t, "Helm", labels["app.kubernetes.io/managed-by"])
	assert.Contains(t, labels, "app.kubernetes.io/version")
	assert.Equal(t, "HTTPRoute", route["kind"])
	assert.Equal(t, releaseName+"-zitadel", metadata["name"])
	assert.Contains(t, metadata, "namespace")
}

func TestGatewayGRPCRouteLabels(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)
	releaseName := "gateway-grpcroute-labels"

	options := gatewayOptions(map[string]string{
		"gateway.grpcRoute.enabled":            "true",
		"gateway.grpcRoute.parentRefs[0].name": "my-gateway",
		"gateway.grpcRoute.hostnames[0]":       "zitadel.example.local",
	})

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{"templates/grpcroute_zitadel.yaml"})

	var route map[string]interface{}
	helm.UnmarshalK8SYaml(t, rendered, &route)

	metadata := route["metadata"].(map[string]interface{})
	labels := metadata["labels"].(map[string]interface{})

	assert.Equal(t, "zitadel", labels["app.kubernetes.io/name"])
	assert.Equal(t, releaseName, labels["app.kubernetes.io/instance"])
	assert.Equal(t, "Helm", labels["app.kubernetes.io/managed-by"])
	assert.Contains(t, labels, "app.kubernetes.io/version")
	assert.Equal(t, "GRPCRoute", route["kind"])
	assert.Equal(t, releaseName+"-zitadel-grpc", metadata["name"])
	assert.Contains(t, metadata, "namespace")
}

func TestGatewayHTTPRouteLoginLabels(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)
	releaseName := "gateway-login-httproute"

	options := gatewayOptions(map[string]string{
		"login.enabled":                               "true",
		"login.gateway.httpRoute.enabled":              "true",
		"login.gateway.httpRoute.parentRefs[0].name":   "my-gateway",
		"login.gateway.httpRoute.hostnames[0]":         "zitadel.example.local",
	})

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{"templates/httproute_login.yaml"})

	var route map[string]interface{}
	helm.UnmarshalK8SYaml(t, rendered, &route)

	metadata := route["metadata"].(map[string]interface{})
	labels := metadata["labels"].(map[string]interface{})

	assert.Equal(t, "zitadel-login", labels["app.kubernetes.io/name"])
	assert.Equal(t, releaseName, labels["app.kubernetes.io/instance"])
	assert.Equal(t, "Helm", labels["app.kubernetes.io/managed-by"])
	assert.Contains(t, labels, "app.kubernetes.io/version")
	assert.Equal(t, "login", labels["app.kubernetes.io/component"])
	assert.Equal(t, "HTTPRoute", route["kind"])
	assert.Contains(t, metadata, "namespace")
}

func TestGatewayRoutesDisabledByDefault(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)
	releaseName := "gateway-disabled"
	options := gatewayOptions(nil)

	_, err := helm.RenderTemplateE(t, options, chartPath, releaseName,
		[]string{"templates/httproute_zitadel.yaml"})
	require.Error(t, err, "httproute_zitadel should not render when disabled")

	_, err = helm.RenderTemplateE(t, options, chartPath, releaseName,
		[]string{"templates/grpcroute_zitadel.yaml"})
	require.Error(t, err, "grpcroute_zitadel should not render when disabled")

	_, err = helm.RenderTemplateE(t, options, chartPath, releaseName,
		[]string{"templates/httproute_login.yaml"})
	require.Error(t, err, "httproute_login should not render when disabled")
}

func TestGatewayHTTPRouteCustomHosts(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)
	releaseName := "gateway-custom-hosts"

	options := gatewayOptions(map[string]string{
		"gateway.httpRoute.enabled":              "true",
		"gateway.httpRoute.parentRefs[0].name":   "my-gw",
		"gateway.httpRoute.hostnames[0]":         "custom.example.com",
		"gateway.httpRoute.hostnames[1]":         "other.example.com",
		"zitadel.configmapConfig.ExternalDomain": "default.example.com",
	})

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName,
		[]string{"templates/httproute_zitadel.yaml"})
	require.Contains(t, rendered, "custom.example.com")
	require.Contains(t, rendered, "other.example.com")
	require.NotContains(t, rendered, "default.example.com")
}

func TestGatewayHTTPRouteAnnotations(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)
	releaseName := "gateway-annotations"

	testCases := []struct {
		name                string
		setJsonValues       map[string]string
		expectedAnnotations map[string]string
	}{
		{
			name: "custom annotations are applied",
			setJsonValues: map[string]string{
				"gateway.httpRoute.annotations": `{"example-foo":"bar","example-baz":"qux"}`,
			},
			expectedAnnotations: map[string]string{
				"example-foo": "bar",
				"example-baz": "qux",
			},
		},
		{
			name:                "no annotations when none specified",
			setJsonValues:       map[string]string{},
			expectedAnnotations: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			options := gatewayOptions(map[string]string{
				"gateway.httpRoute.enabled":            "true",
				"gateway.httpRoute.parentRefs[0].name": "my-gw",
			})
			options.SetJsonValues = tc.setJsonValues

			rendered := helm.RenderTemplate(t, options, chartPath, releaseName,
				[]string{"templates/httproute_zitadel.yaml"})

			if tc.expectedAnnotations == nil {
				assert.NotContains(t, rendered, "annotations:")
			} else {
				for key, value := range tc.expectedAnnotations {
					assert.Contains(t, rendered, key)
					assert.Contains(t, rendered, value)
				}
			}
		})
	}
}

func TestGatewayHTTPRouteDefaultPaths(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	t.Run("zitadel default path is /", func(t *testing.T) {
		options := gatewayOptions(map[string]string{
			"gateway.httpRoute.enabled":            "true",
			"gateway.httpRoute.parentRefs[0].name": "my-gw",
		})

		rendered := helm.RenderTemplate(t, options, chartPath, "gateway-default-paths",
			[]string{"templates/httproute_zitadel.yaml"})
		require.Contains(t, rendered, `value: "/"`)
		require.Contains(t, rendered, "type: PathPrefix")
	})

	t.Run("login default path is /ui/v2/login", func(t *testing.T) {
		options := gatewayOptions(map[string]string{
			"login.enabled":                               "true",
			"login.gateway.httpRoute.enabled":              "true",
			"login.gateway.httpRoute.parentRefs[0].name":   "my-gw",
		})

		rendered := helm.RenderTemplate(t, options, chartPath, "gateway-default-paths",
			[]string{"templates/httproute_login.yaml"})
		require.Contains(t, rendered, `value: "/ui/v2/login"`)
		require.Contains(t, rendered, "type: PathPrefix")
	})
}

func TestGatewayHTTPRouteEmptyPathsFails(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	t.Run("zitadel empty paths fails", func(t *testing.T) {
		options := gatewayOptions(map[string]string{
			"gateway.httpRoute.enabled":            "true",
			"gateway.httpRoute.parentRefs[0].name": "my-gw",
		})
		options.SetJsonValues = map[string]string{
			"gateway.httpRoute.paths": "[]",
		}

		_, err := helm.RenderTemplateE(t, options, chartPath, "gateway-empty-paths",
			[]string{"templates/httproute_zitadel.yaml"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "gateway.httpRoute.paths must not be empty")
	})

	t.Run("login empty paths fails", func(t *testing.T) {
		options := gatewayOptions(map[string]string{
			"login.enabled":                               "true",
			"login.gateway.httpRoute.enabled":              "true",
			"login.gateway.httpRoute.parentRefs[0].name":   "my-gw",
		})
		options.SetJsonValues = map[string]string{
			"login.gateway.httpRoute.paths": "[]",
		}

		_, err := helm.RenderTemplateE(t, options, chartPath, "gateway-empty-paths",
			[]string{"templates/httproute_login.yaml"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "login.gateway.httpRoute.paths must not be empty")
	})
}

func TestGatewayHTTPRouteParentRefs(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)
	releaseName := "gateway-parentrefs"

	options := gatewayOptions(map[string]string{
		"gateway.httpRoute.enabled":                   "true",
		"gateway.httpRoute.parentRefs[0].name":        "my-gateway",
		"gateway.httpRoute.parentRefs[0].namespace":   "gateway-ns",
		"gateway.httpRoute.parentRefs[0].sectionName": "https",
		"gateway.httpRoute.parentRefs[0].port":        "443",
	})

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName,
		[]string{"templates/httproute_zitadel.yaml"})
	require.Contains(t, rendered, "name: my-gateway")
	require.Contains(t, rendered, "namespace: gateway-ns")
	require.Contains(t, rendered, "sectionName: https")
	require.Contains(t, rendered, "port: 443")
}

func TestGatewayHTTPRouteFilters(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)
	releaseName := "gateway-filters"

	options := gatewayOptions(map[string]string{
		"gateway.httpRoute.enabled":            "true",
		"gateway.httpRoute.parentRefs[0].name": "my-gw",
	})
	options.SetJsonValues = map[string]string{
		"gateway.httpRoute.filters": `[{"type":"RequestHeaderModifier","requestHeaderModifier":{"set":[{"name":"X-Custom","value":"test"}]}}]`,
	}

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName,
		[]string{"templates/httproute_zitadel.yaml"})
	require.Contains(t, rendered, "filters:")
	require.Contains(t, rendered, "RequestHeaderModifier")
	require.Contains(t, rendered, "X-Custom")
}

func TestGatewayHTTPRouteTimeouts(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)
	releaseName := "gateway-timeouts"

	options := gatewayOptions(map[string]string{
		"gateway.httpRoute.enabled":                 "true",
		"gateway.httpRoute.parentRefs[0].name":      "my-gw",
		"gateway.httpRoute.timeouts.request":        "30s",
		"gateway.httpRoute.timeouts.backendRequest": "20s",
	})

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName,
		[]string{"templates/httproute_zitadel.yaml"})
	require.Contains(t, rendered, "timeouts:")
	require.Contains(t, rendered, "request: 30s")
	require.Contains(t, rendered, "backendRequest: 20s")
}

func TestGatewayHTTPRouteCustomLabels(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)
	releaseName := "gateway-custom-labels"

	options := gatewayOptions(map[string]string{
		"gateway.httpRoute.enabled":             "true",
		"gateway.httpRoute.parentRefs[0].name":  "my-gw",
		"gateway.httpRoute.labels.custom-label": "custom-value",
	})

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName,
		[]string{"templates/httproute_zitadel.yaml"})

	var route map[string]interface{}
	helm.UnmarshalK8SYaml(t, rendered, &route)

	metadata := route["metadata"].(map[string]interface{})
	labels := metadata["labels"].(map[string]interface{})
	assert.Equal(t, "custom-value", labels["custom-label"])
}

func TestGatewayGRPCRouteFilters(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)
	releaseName := "gateway-grpc-filters"

	options := gatewayOptions(map[string]string{
		"gateway.grpcRoute.enabled":            "true",
		"gateway.grpcRoute.parentRefs[0].name": "my-gw",
	})
	options.SetJsonValues = map[string]string{
		"gateway.grpcRoute.filters": `[{"type":"RequestHeaderModifier","requestHeaderModifier":{"set":[{"name":"X-Custom","value":"grpc-test"}]}}]`,
	}

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName,
		[]string{"templates/grpcroute_zitadel.yaml"})
	require.Contains(t, rendered, "filters:")
	require.Contains(t, rendered, "RequestHeaderModifier")
	require.Contains(t, rendered, "grpc-test")
}
