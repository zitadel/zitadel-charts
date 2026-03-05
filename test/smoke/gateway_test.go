package smoke_test_test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"

	"github.com/zitadel/zitadel-charts/test/smoke/support"
)

func TestGatewayHTTPRouteZitadelLabels(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                                    support.DigestTag,
			"gateway.zitadel.httpRoute.enabled":            "true",
			"gateway.zitadel.httpRoute.parentRefs[0].name": "my-gateway",
			"zitadel.configmapConfig.ExternalDomain":       "zitadel.example.local",
			"zitadel.masterkey":                            "01234567890123456789012345678901",
		},
	}

	releaseName := "gateway-httproute-labels"

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{"templates/httproute_zitadel.yaml"})
	require.Contains(t, rendered, "kind: HTTPRoute")
	require.Contains(t, rendered, "name: "+releaseName+"-zitadel")
	require.Contains(t, rendered, "namespace:")
	require.Contains(t, rendered, "zitadel.example.local")
	require.Contains(t, rendered, "my-gateway")
}

func TestGatewayGRPCRouteZitadelLabels(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                                    support.DigestTag,
			"gateway.zitadel.grpcRoute.enabled":            "true",
			"gateway.zitadel.grpcRoute.parentRefs[0].name": "my-gateway",
			"zitadel.configmapConfig.ExternalDomain":       "zitadel.example.local",
			"zitadel.masterkey":                            "01234567890123456789012345678901",
		},
	}

	releaseName := "gateway-grpcroute-labels"

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{"templates/grpcroute_zitadel.yaml"})
	require.Contains(t, rendered, "kind: GRPCRoute")
	require.Contains(t, rendered, "name: "+releaseName+"-zitadel-grpc")
	require.Contains(t, rendered, "namespace:")
	require.Contains(t, rendered, "zitadel.example.local")
	require.Contains(t, rendered, "my-gateway")
}

func TestGatewayHTTPRouteLoginLabels(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                       support.DigestTag,
			"login.enabled":                   "true",
			"gateway.login.httpRoute.enabled": "true",
			"gateway.login.httpRoute.parentRefs[0].name": "my-gateway",
			"zitadel.configmapConfig.ExternalDomain":     "zitadel.example.local",
			"zitadel.masterkey":                          "01234567890123456789012345678901",
		},
	}

	releaseName := "gateway-login-httproute"

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{"templates/httproute_login.yaml"})
	require.Contains(t, rendered, "kind: HTTPRoute")
	require.Contains(t, rendered, releaseName+"-zitadel-login")
	require.Contains(t, rendered, "namespace:")
	require.Contains(t, rendered, "/ui/v2/login")
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

func TestGatewayHTTPRouteCustomHostnames(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                                    support.DigestTag,
			"gateway.zitadel.httpRoute.enabled":            "true",
			"gateway.zitadel.httpRoute.parentRefs[0].name": "my-gw",
			"gateway.zitadel.httpRoute.hostnames[0]":       "custom.example.com",
			"gateway.zitadel.httpRoute.hostnames[1]":       "other.example.com",
			"zitadel.configmapConfig.ExternalDomain":       "default.example.com",
			"zitadel.masterkey":                            "01234567890123456789012345678901",
		},
	}

	releaseName := "gateway-custom-hosts"

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName,
		[]string{"templates/httproute_zitadel.yaml"})
	require.Contains(t, rendered, "custom.example.com")
	require.Contains(t, rendered, "other.example.com")
	require.NotContains(t, rendered, "default.example.com")
}
