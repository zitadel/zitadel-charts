package smoke_test_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zitadel/zitadel-charts/test/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

func TestWait4xInitContainerResources(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)
	chartPath := setup.ChartPath(t)

	testCases := []struct {
		name      string
		setValues map[string]string
		zitadel   *assert.DeploymentAssertion
		login     *assert.DeploymentAssertion
	}{
		{
			name: "default-no-resources",
			setValues: map[string]string{
				"login.enabled": "true",
			},
			zitadel: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						Spec: assert.PodSpecAssertion{
							InitContainers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("wait-for-postgres"),
									Resources: assert.ResourceRequirementsAssertion{
										Requests: assert.Some(corev1.ResourceList{}),
										Limits:   assert.Some(corev1.ResourceList{}),
									},
								},
							}),
						},
					},
				},
			},
			login: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						Spec: assert.PodSpecAssertion{
							InitContainers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("wait-for-zitadel"),
									Resources: assert.ResourceRequirementsAssertion{
										Requests: assert.Some(corev1.ResourceList{}),
										Limits:   assert.Some(corev1.ResourceList{}),
									},
								},
							}),
						},
					},
				},
			},
		},
		{
			name: "with-wait4x-resources",
			setValues: map[string]string{
				"login.enabled":                          "true",
				"tools.wait4x.resources.requests.cpu":    "50m",
				"tools.wait4x.resources.requests.memory": "32Mi",
				"tools.wait4x.resources.limits.cpu":      "100m",
				"tools.wait4x.resources.limits.memory":   "64Mi",
			},
			zitadel: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						Spec: assert.PodSpecAssertion{
							InitContainers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("wait-for-postgres"),
									Resources: assert.ResourceRequirementsAssertion{
										Requests: assert.Some(corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("50m"),
											corev1.ResourceMemory: resource.MustParse("32Mi"),
										}),
										Limits: assert.Some(corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("100m"),
											corev1.ResourceMemory: resource.MustParse("64Mi"),
										}),
									},
								},
							}),
						},
					},
				},
			},
			login: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						Spec: assert.PodSpecAssertion{
							InitContainers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("wait-for-zitadel"),
									Resources: assert.ResourceRequirementsAssertion{
										Requests: assert.Some(corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("50m"),
											corev1.ResourceMemory: resource.MustParse("32Mi"),
										}),
										Limits: assert.Some(corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("100m"),
											corev1.ResourceMemory: resource.MustParse("64Mi"),
										}),
									},
								},
							}),
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			support.WithNamespace(t, cluster, func(env *support.Env) {
				releaseName := setup.InstallZitadel(t, env, chartPath, tc.name, tc.setValues)

				if tc.zitadel != nil {
					assert.AssertPartial(t, env.GetDeployment(t, releaseName), *tc.zitadel, releaseName)
				}
				if tc.login != nil {
					assert.AssertPartial(t, env.GetDeployment(t, releaseName+"-login"), *tc.login, releaseName+"-login")
				}
			})
		})
	}
}
