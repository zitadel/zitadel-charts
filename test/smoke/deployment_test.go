package smoke_test_test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/zitadel/zitadel-charts/test/smoke/support"
)

func TestDeploymentLabels(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":         support.DigestTag,
			"login.enabled":     "true",
			"ingress.enabled":   "true",
			"zitadel.masterkey": "01234567890123456789012345678901",
		},
	}

	releaseName := "deployment-labels"

	testCases := []struct {
		name      string
		template  string
		appName   string
		component string
	}{
		{
			name:      "zitadel",
			template:  "templates/deployment_zitadel.yaml",
			appName:   "zitadel",
			component: "start",
		},
		{
			name:      "login",
			template:  "templates/deployment_login.yaml",
			appName:   "zitadel-login",
			component: "login",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{tc.template})

			var deployment appsv1.Deployment
			helm.UnmarshalK8SYaml(t, rendered, &deployment)

			expectedLabels := support.ExpectedLabels(releaseName, tc.appName, support.ExpectedVersion, tc.component, nil)
			selectorLabels := map[string]string{
				"app.kubernetes.io/name":     tc.appName,
				"app.kubernetes.io/instance": releaseName,
			}
			if tc.component != "" {
				selectorLabels["app.kubernetes.io/component"] = tc.component
			}

			support.AssertLabels(t, deployment.Labels, expectedLabels)
			support.AssertLabels(t, deployment.Spec.Selector.MatchLabels, selectorLabels)
			support.AssertLabels(t, deployment.Spec.Template.Labels, expectedLabels)
		})
	}
}
