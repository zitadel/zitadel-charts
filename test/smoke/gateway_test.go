package smoke_test_test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zitadel/zitadel-charts/test/smoke/support"
)

func TestGatewayHTTPRouteLabels(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                              support.DigestTag,
			"gateway.httpRoute.enabled":              "true",
			"gateway.httpRoute.parentRefs[0].name":   "my-gateway",
			"gateway.httpRoute.hostnames[0]":         "zitadel.example.local",
			"zitadel.configmapConfig.ExternalDomain": "zitadel.example.local",
			"zitadel.masterkey":                      "01234567890123456789012345678901",
		},
	}

	releaseName := "gateway-httproute-labels"

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{"templates/httproute_zitadel.yaml"})

	var route map[string]interface{}
	helm.UnmarshalK8SYaml(t, rendered, &route)

	metadata := route["metadata"].(map[string]interface{})
	labels := make(map[string]string)
	for k, v := range metadata["labels"].(map[string]interface{}) {
		labels[k] = v.(string)
	}
	support.AssertLabels(
		t,
		labels,
		support.ExpectedLabels(releaseName, "zitadel", support.ExpectedVersion, "", nil),
	)

	assert.Equal(t, "HTTPRoute", route["kind"])
	assert.Equal(t, releaseName+"-zitadel", metadata["name"])
	assert.Contains(t, metadata, "namespace")
}

func TestGatewayGRPCRouteLabels(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                              support.DigestTag,
			"gateway.grpcRoute.enabled":              "true",
			"gateway.grpcRoute.parentRefs[0].name":   "my-gateway",
			"gateway.grpcRoute.hosts[0]":             "zitadel.example.local",
			"zitadel.configmapConfig.ExternalDomain": "zitadel.example.local",
			"zitadel.masterkey":                      "01234567890123456789012345678901",
		},
	}

	releaseName := "gateway-grpcroute-labels"

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{"templates/grpcroute_zitadel.yaml"})

	var route map[string]interface{}
	helm.UnmarshalK8SYaml(t, rendered, &route)

	metadata := route["metadata"].(map[string]interface{})
	labels := make(map[string]string)
	for k, v := range metadata["labels"].(map[string]interface{}) {
		labels[k] = v.(string)
	}
	support.AssertLabels(
		t,
		labels,
		support.ExpectedLabels(releaseName, "zitadel", support.ExpectedVersion, "", nil),
	)

	assert.Equal(t, "GRPCRoute", route["kind"])
	assert.Equal(t, releaseName+"-zitadel-grpc", metadata["name"])
	assert.Contains(t, metadata, "namespace")
}

func TestGatewayHTTPRouteLoginLabels(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                       support.DigestTag,
			"login.enabled":                   "true",
			"login.gateway.httpRoute.enabled": "true",
			"login.gateway.httpRoute.parentRefs[0].name": "my-gateway",
			"login.gateway.httpRoute.hostnames[0]":       "zitadel.example.local",
			"zitadel.configmapConfig.ExternalDomain":     "zitadel.example.local",
			"zitadel.masterkey":                          "01234567890123456789012345678901",
		},
	}

	releaseName := "gateway-login-httproute"

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{"templates/httproute_login.yaml"})

	var route map[string]interface{}
	helm.UnmarshalK8SYaml(t, rendered, &route)

	metadata := route["metadata"].(map[string]interface{})
	labels := make(map[string]string)
	for k, v := range metadata["labels"].(map[string]interface{}) {
		labels[k] = v.(string)
	}
	support.AssertLabels(
		t,
		labels,
		support.ExpectedLabels(releaseName, "zitadel-login", support.ExpectedVersion, "login", nil),
	)

	assert.Equal(t, "HTTPRoute", route["kind"])
	assert.Contains(t, metadata, "namespace")
}

func TestGatewayRoutesDisabledByDefault(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":         support.DigestTag,
			"zitadel.masterkey": "01234567890123456789012345678901",
		},
	}

	releaseName := "gateway-disabled"

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

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                              support.DigestTag,
			"gateway.httpRoute.enabled":              "true",
			"gateway.httpRoute.parentRefs[0].name":   "my-gw",
			"gateway.httpRoute.hostnames[0]":         "custom.example.com",
			"gateway.httpRoute.hostnames[1]":         "other.example.com",
			"zitadel.configmapConfig.ExternalDomain": "default.example.com",
			"zitadel.masterkey":                      "01234567890123456789012345678901",
		},
	}

	releaseName := "gateway-custom-hosts"

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
			options := &helm.Options{
				SetValues: map[string]string{
					"image.tag":                              support.DigestTag,
					"gateway.httpRoute.enabled":              "true",
					"gateway.httpRoute.parentRefs[0].name":   "my-gw",
					"zitadel.configmapConfig.ExternalDomain": "zitadel.example.local",
					"zitadel.masterkey":                      "01234567890123456789012345678901",
				},
				SetJsonValues: tc.setJsonValues,
			}
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
		options := &helm.Options{
			SetValues: map[string]string{
				"image.tag":                              support.DigestTag,
				"gateway.httpRoute.enabled":              "true",
				"gateway.httpRoute.parentRefs[0].name":   "my-gw",
				"zitadel.configmapConfig.ExternalDomain": "zitadel.example.local",
				"zitadel.masterkey":                      "01234567890123456789012345678901",
			},
		}

		rendered := helm.RenderTemplate(t, options, chartPath, "gateway-default-paths",
			[]string{"templates/httproute_zitadel.yaml"})
		require.Contains(t, rendered, `value: "/"`)
		require.Contains(t, rendered, "type: PathPrefix")
	})

	t.Run("login default path is /ui/v2/login", func(t *testing.T) {
		options := &helm.Options{
			SetValues: map[string]string{
				"image.tag":                       support.DigestTag,
				"login.enabled":                   "true",
				"login.gateway.httpRoute.enabled": "true",
				"login.gateway.httpRoute.parentRefs[0].name": "my-gw",
				"zitadel.configmapConfig.ExternalDomain":     "zitadel.example.local",
				"zitadel.masterkey":                          "01234567890123456789012345678901",
			},
		}

		rendered := helm.RenderTemplate(t, options, chartPath, "gateway-default-paths",
			[]string{"templates/httproute_login.yaml"})
		require.Contains(t, rendered, `value: "/ui/v2/login"`)
		require.Contains(t, rendered, "type: PathPrefix")
	})
}

func TestGatewayHTTPRouteParentRefs(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                                   support.DigestTag,
			"gateway.httpRoute.enabled":                   "true",
			"gateway.httpRoute.parentRefs[0].name":        "my-gateway",
			"gateway.httpRoute.parentRefs[0].namespace":   "gateway-ns",
			"gateway.httpRoute.parentRefs[0].sectionName": "https",
			"gateway.httpRoute.parentRefs[0].port":        "443",
			"zitadel.configmapConfig.ExternalDomain":      "zitadel.example.local",
			"zitadel.masterkey":                           "01234567890123456789012345678901",
		},
	}

	releaseName := "gateway-parentrefs"

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

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                              support.DigestTag,
			"gateway.httpRoute.enabled":              "true",
			"gateway.httpRoute.parentRefs[0].name":   "my-gw",
			"zitadel.configmapConfig.ExternalDomain": "zitadel.example.local",
			"zitadel.masterkey":                      "01234567890123456789012345678901",
		},
		SetJsonValues: map[string]string{
			"gateway.httpRoute.filters": `[{"type":"RequestHeaderModifier","requestHeaderModifier":{"set":[{"name":"X-Custom","value":"test"}]}}]`,
		},
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

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                                 support.DigestTag,
			"gateway.httpRoute.enabled":                 "true",
			"gateway.httpRoute.parentRefs[0].name":      "my-gw",
			"gateway.httpRoute.timeouts.request":        "30s",
			"gateway.httpRoute.timeouts.backendRequest": "20s",
			"zitadel.configmapConfig.ExternalDomain":    "zitadel.example.local",
			"zitadel.masterkey":                         "01234567890123456789012345678901",
		},
	}

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

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                              support.DigestTag,
			"gateway.httpRoute.enabled":              "true",
			"gateway.httpRoute.parentRefs[0].name":   "my-gw",
			"gateway.httpRoute.labels.custom-label":  "custom-value",
			"zitadel.configmapConfig.ExternalDomain": "zitadel.example.local",
			"zitadel.masterkey":                      "01234567890123456789012345678901",
		},
	}

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

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                              support.DigestTag,
			"gateway.grpcRoute.enabled":              "true",
			"gateway.grpcRoute.parentRefs[0].name":   "my-gw",
			"zitadel.configmapConfig.ExternalDomain": "zitadel.example.local",
			"zitadel.masterkey":                      "01234567890123456789012345678901",
		},
		SetJsonValues: map[string]string{
			"gateway.grpcRoute.filters": `[{"type":"RequestHeaderModifier","requestHeaderModifier":{"set":[{"name":"X-Custom","value":"grpc-test"}]}}]`,
		},
	}

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName,
		[]string{"templates/grpcroute_zitadel.yaml"})
	require.Contains(t, rendered, "filters:")
	require.Contains(t, rendered, "RequestHeaderModifier")
	require.Contains(t, rendered, "grpc-test")
}
