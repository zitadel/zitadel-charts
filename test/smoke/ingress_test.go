package smoke_test_test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
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
