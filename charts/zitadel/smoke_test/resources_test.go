package smoke_test_test

import (
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zitadel/zitadel-charts/charts/zitadel/smoke_test/support"
)

func TestWait4xInitContainerResources(t *testing.T) {
	t.Parallel()

	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	testCases := []struct {
		name            string
		setValues       map[string]string
		expectResources bool
	}{
		{
			name: "default-no-resources",
			setValues: map[string]string{
				"image.tag":         support.DigestTag,
				"login.enabled":     "true",
				"zitadel.masterkey": "01234567890123456789012345678901",
			},
			expectResources: false,
		},
		{
			name: "with-wait4x-resources",
			setValues: map[string]string{
				"image.tag":                              support.DigestTag,
				"login.enabled":                          "true",
				"zitadel.masterkey":                      "01234567890123456789012345678901",
				"tools.wait4x.resources.requests.cpu":    "50m",
				"tools.wait4x.resources.requests.memory": "32Mi",
				"tools.wait4x.resources.limits.cpu":      "100m",
				"tools.wait4x.resources.limits.memory":   "64Mi",
			},
			expectResources: true,
		},
	}

	releaseName := "wait4x-resources"

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := &helm.Options{
				SetValues: tc.setValues,
			}

			t.Run("login-deployment", func(t *testing.T) {
				rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{"templates/deployment_login.yaml"})

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(t, rendered, &deployment)

				initContainer := findInitContainer(deployment.Spec.Template.Spec.InitContainers, "wait-for-zitadel")
				require.NotNil(t, initContainer, "wait-for-zitadel init container not found")

				if tc.expectResources {
					require.NotEmpty(t, initContainer.Resources.Requests, "expected resources.requests to be set")
					require.NotEmpty(t, initContainer.Resources.Limits, "expected resources.limits to be set")
					require.Equal(t, resource.MustParse("50m"), initContainer.Resources.Requests["cpu"])
					require.Equal(t, resource.MustParse("32Mi"), initContainer.Resources.Requests["memory"])
					require.Equal(t, resource.MustParse("100m"), initContainer.Resources.Limits["cpu"])
					require.Equal(t, resource.MustParse("64Mi"), initContainer.Resources.Limits["memory"])
				} else {
					require.Empty(t, initContainer.Resources.Requests, "expected no resources.requests")
					require.Empty(t, initContainer.Resources.Limits, "expected no resources.limits")
				}
			})

			t.Run("zitadel-deployment", func(t *testing.T) {
				rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{"templates/deployment_zitadel.yaml"})

				var deployment appsv1.Deployment
				helm.UnmarshalK8SYaml(t, rendered, &deployment)

				initContainer := findInitContainer(deployment.Spec.Template.Spec.InitContainers, "wait-for-postgres")
				if initContainer == nil {
					t.Skip("wait-for-postgres init container not present (postgres endpoint not configured)")
				}

				if tc.expectResources {
					require.NotEmpty(t, initContainer.Resources.Requests, "expected resources.requests to be set")
					require.NotEmpty(t, initContainer.Resources.Limits, "expected resources.limits to be set")
					require.Equal(t, resource.MustParse("50m"), initContainer.Resources.Requests["cpu"])
					require.Equal(t, resource.MustParse("32Mi"), initContainer.Resources.Requests["memory"])
					require.Equal(t, resource.MustParse("100m"), initContainer.Resources.Limits["cpu"])
					require.Equal(t, resource.MustParse("64Mi"), initContainer.Resources.Limits["memory"])
				} else {
					require.Empty(t, initContainer.Resources.Requests, "expected no resources.requests")
					require.Empty(t, initContainer.Resources.Limits, "expected no resources.limits")
				}
			})
		})
	}
}

func findInitContainer(containers []corev1.Container, name string) *corev1.Container {
	for i := range containers {
		if containers[i].Name == name {
			return &containers[i]
		}
	}
	return nil
}
