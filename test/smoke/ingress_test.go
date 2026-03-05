package smoke_test_test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"

	"github.com/zitadel/zitadel-charts/test/smoke/support"
)

func TestIngressLabels(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                              support.DigestTag,
			"ingress.enabled":                        "true",
			"ingress.hosts[0].host":                  "zitadel.example.local",
			"zitadel.configmapConfig.ExternalDomain": "zitadel.example.local",
			"login.enabled":                          "true",
			"login.ingress.enabled":                  "true",
			"login.ingress.hosts[0].host":            "login.example.local",
			"zitadel.masterkey":                      "01234567890123456789012345678901",
		},
	}

	releaseName := "ingress-labels"

	renderedZitadel := helm.RenderTemplate(t, options, chartPath, releaseName, []string{"templates/ingress_zitadel.yaml"})
	var zitadelIngress networkingv1.Ingress
	helm.UnmarshalK8SYaml(t, renderedZitadel, &zitadelIngress)
	support.AssertLabels(
		t,
		zitadelIngress.Labels,
		support.ExpectedLabels(releaseName, "zitadel", support.ExpectedVersion, "", nil),
	)

	renderedLogin := helm.RenderTemplate(t, options, chartPath, releaseName, []string{"templates/ingress_login.yaml"})
	var loginIngress networkingv1.Ingress
	helm.UnmarshalK8SYaml(t, renderedLogin, &loginIngress)
	support.AssertLabels(
		t,
		loginIngress.Labels,
		support.ExpectedLabels(releaseName, "zitadel-login", support.ExpectedVersion, "login", nil),
	)
}

func TestIngressNginxControllerAnnotation(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)
	releaseName := "ingress-nginx"

	testCases := []struct {
		name                string
		controller          string
		expectedAnnotations map[string]string
	}{
		{
			name:       "nginx controller injects backend-protocol annotation",
			controller: "nginx",
			expectedAnnotations: map[string]string{
				"nginx.ingress.kubernetes.io/backend-protocol": "GRPC",
			},
		},
		{
			name:                "generic controller omits backend-protocol annotation",
			controller:          "generic",
			expectedAnnotations: map[string]string{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			options := &helm.Options{
				SetValues: map[string]string{
					"image.tag":                              support.DigestTag,
					"ingress.enabled":                        "true",
					"ingress.controller":                     tc.controller,
					"ingress.hosts[0].host":                  "zitadel.example.local",
					"zitadel.configmapConfig.ExternalDomain": "zitadel.example.local",
					"zitadel.masterkey":                      "01234567890123456789012345678901",
				},
			}
			rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{"templates/ingress_zitadel.yaml"})
			var ingress networkingv1.Ingress
			helm.UnmarshalK8SYaml(t, rendered, &ingress)

			for key, value := range tc.expectedAnnotations {
				assert.Equal(t, value, ingress.Annotations[key])
			}
			for key := range ingress.Annotations {
				assert.Contains(t, tc.expectedAnnotations, key, "unexpected annotation %q", key)
			}
		})
	}
}
