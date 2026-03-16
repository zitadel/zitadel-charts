package smoke_test_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zitadel/zitadel-charts/test/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

//goland:noinspection ALL
func TestDeploymentMatrix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setValues map[string]string
		zitadel   *assert.DeploymentAssertion
		login     *assert.DeploymentAssertion
	}{
		{
			name: "defaults",
			setValues: map[string]string{
				"login.enabled":         "true",
				"login.ingress.enabled": "true",
			},
			zitadel: &assert.DeploymentAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Some(map[string]string{
						"app.kubernetes.io/name":       "zitadel",
						"app.kubernetes.io/version":    "v4.12.1",
						"app.kubernetes.io/managed-by": "Helm",
						"app.kubernetes.io/component":  "start",
					}),
				},
				Spec: assert.DeploymentSpecAssertion{
					Selector: assert.LabelSelectorAssertion{
						MatchLabels: assert.Some(map[string]string{
							"app.kubernetes.io/name":      "zitadel",
							"app.kubernetes.io/component": "start",
							"app.kubernetes.io/instance":  support.ReleaseName,
						}),
					},
					Template: assert.PodTemplateSpecAssertion{
						ObjectMeta: assert.ObjectMetaAssertion{
							Labels: assert.Some(map[string]string{
								"app.kubernetes.io/name":       "zitadel",
								"app.kubernetes.io/version":    "v4.12.1",
								"app.kubernetes.io/managed-by": "Helm",
								"app.kubernetes.io/component":  "start",
							}),
						},
						Spec: assert.PodSpecAssertion{
							SecurityContext: assert.PodSecurityContextAssertion{
								RunAsNonRoot: assert.SomePtr(true),
								RunAsUser:    assert.SomePtr(int64(1000)),
								FSGroup:      assert.SomePtr(int64(1000)),
							},
							InitContainers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("wait-for-postgres"),
									Resources: assert.ResourceRequirementsAssertion{
										Requests: assert.Some(corev1.ResourceList{}),
										Limits:   assert.Some(corev1.ResourceList{}),
									},
									SecurityContext: assert.SecurityContextAssertion{
										RunAsNonRoot:           assert.SomePtr(true),
										RunAsUser:              assert.SomePtr(int64(1000)),
										ReadOnlyRootFilesystem: assert.SomePtr(true),
										Privileged:             assert.SomePtr(false),
									},
								},
							}),
							Containers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("zitadel"),
									SecurityContext: assert.SecurityContextAssertion{
										RunAsNonRoot:           assert.SomePtr(true),
										RunAsUser:              assert.SomePtr(int64(1000)),
										ReadOnlyRootFilesystem: assert.SomePtr(true),
										Privileged:             assert.SomePtr(false),
									},
								},
							}),
						},
					},
				},
			},
			login: &assert.DeploymentAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Some(map[string]string{
						"app.kubernetes.io/name":       "zitadel-login",
						"app.kubernetes.io/version":    "v4.12.1",
						"app.kubernetes.io/managed-by": "Helm",
						"app.kubernetes.io/component":  "login",
					}),
				},
				Spec: assert.DeploymentSpecAssertion{
					Selector: assert.LabelSelectorAssertion{
						MatchLabels: assert.Some(map[string]string{
							"app.kubernetes.io/name":      "zitadel-login",
							"app.kubernetes.io/component": "login",
						}),
					},
					Template: assert.PodTemplateSpecAssertion{
						ObjectMeta: assert.ObjectMetaAssertion{
							Labels: assert.Some(map[string]string{
								"app.kubernetes.io/name":       "zitadel-login",
								"app.kubernetes.io/version":    "v4.12.1",
								"app.kubernetes.io/managed-by": "Helm",
								"app.kubernetes.io/component":  "login",
							}),
						},
						Spec: assert.PodSpecAssertion{
							SecurityContext: assert.PodSecurityContextAssertion{
								RunAsNonRoot: assert.SomePtr(true),
								RunAsUser:    assert.SomePtr(int64(1000)),
								FSGroup:      assert.SomePtr(int64(1000)),
							},
							InitContainers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("wait-for-zitadel"),
									Resources: assert.ResourceRequirementsAssertion{
										Requests: assert.Some(corev1.ResourceList{}),
										Limits:   assert.Some(corev1.ResourceList{}),
									},
									SecurityContext: assert.SecurityContextAssertion{
										RunAsNonRoot:           assert.SomePtr(true),
										RunAsUser:              assert.SomePtr(int64(1000)),
										ReadOnlyRootFilesystem: assert.SomePtr(true),
										Privileged:             assert.SomePtr(false),
									},
								},
							}),
							Containers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("zitadel-login"),
									SecurityContext: assert.SecurityContextAssertion{
										RunAsNonRoot:           assert.SomePtr(true),
										RunAsUser:              assert.SomePtr(int64(1000)),
										ReadOnlyRootFilesystem: assert.SomePtr(true),
										Privileged:             assert.SomePtr(false),
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
		{
			name: "component-overrides",
			setValues: map[string]string{
				"login.enabled":         "true",
				"login.ingress.enabled": "true",

				"zitadel.podSecurityContext.runAsNonRoot":          "true",
				"zitadel.podSecurityContext.runAsUser":             "2000",
				"zitadel.podSecurityContext.fsGroup":               "2000",
				"zitadel.podSecurityContext.seccompProfile.type":   "RuntimeDefault",
				"zitadel.securityContext.runAsNonRoot":             "true",
				"zitadel.securityContext.runAsUser":                "2000",
				"zitadel.securityContext.readOnlyRootFilesystem":   "true",
				"zitadel.securityContext.privileged":               "false",
				"zitadel.securityContext.allowPrivilegeEscalation": "false",
				"zitadel.securityContext.capabilities.drop[0]":     "ALL",
				"login.podSecurityContext.runAsNonRoot":            "true",
				"login.podSecurityContext.runAsUser":               "3000",
				"login.podSecurityContext.fsGroup":                 "3000",
				"login.podSecurityContext.seccompProfile.type":     "RuntimeDefault",
				"login.securityContext.runAsNonRoot":               "true",
				"login.securityContext.runAsUser":                  "3000",
				"login.securityContext.readOnlyRootFilesystem":     "true",
				"login.securityContext.privileged":                 "false",
				"login.securityContext.allowPrivilegeEscalation":   "false",
				"login.securityContext.capabilities.drop[0]":       "NET_RAW",
			},
			zitadel: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						Spec: assert.PodSpecAssertion{
							SecurityContext: assert.PodSecurityContextAssertion{
								RunAsNonRoot: assert.SomePtr(true),
								RunAsUser:    assert.SomePtr(int64(2000)),
								FSGroup:      assert.SomePtr(int64(2000)),
								SeccompProfile: assert.SeccompProfileAssertion{
									Type: assert.Some(corev1.SeccompProfileTypeRuntimeDefault),
								},
							},
							InitContainers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("wait-for-postgres"),
									SecurityContext: assert.SecurityContextAssertion{
										RunAsNonRoot:             assert.SomePtr(true),
										RunAsUser:                assert.SomePtr(int64(2000)),
										ReadOnlyRootFilesystem:   assert.SomePtr(true),
										Privileged:               assert.SomePtr(false),
										AllowPrivilegeEscalation: assert.SomePtr(false),
										Capabilities: assert.CapabilitiesAssertion{
											Drop: assert.Some([]corev1.Capability{"ALL"}),
										},
									},
								},
							}),
							Containers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("zitadel"),
									SecurityContext: assert.SecurityContextAssertion{
										RunAsNonRoot:             assert.SomePtr(true),
										RunAsUser:                assert.SomePtr(int64(2000)),
										ReadOnlyRootFilesystem:   assert.SomePtr(true),
										Privileged:               assert.SomePtr(false),
										AllowPrivilegeEscalation: assert.SomePtr(false),
										Capabilities: assert.CapabilitiesAssertion{
											Drop: assert.Some([]corev1.Capability{"ALL"}),
										},
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
							SecurityContext: assert.PodSecurityContextAssertion{
								RunAsNonRoot: assert.SomePtr(true),
								RunAsUser:    assert.SomePtr(int64(3000)),
								FSGroup:      assert.SomePtr(int64(3000)),
								SeccompProfile: assert.SeccompProfileAssertion{
									Type: assert.Some(corev1.SeccompProfileTypeRuntimeDefault),
								},
							},
							InitContainers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("wait-for-zitadel"),
									SecurityContext: assert.SecurityContextAssertion{
										RunAsNonRoot:             assert.SomePtr(true),
										RunAsUser:                assert.SomePtr(int64(3000)),
										ReadOnlyRootFilesystem:   assert.SomePtr(true),
										Privileged:               assert.SomePtr(false),
										AllowPrivilegeEscalation: assert.SomePtr(false),
										Capabilities: assert.CapabilitiesAssertion{
											Drop: assert.Some([]corev1.Capability{"NET_RAW"}),
										},
									},
								},
							}),
							Containers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("zitadel-login"),
									SecurityContext: assert.SecurityContextAssertion{
										RunAsNonRoot:             assert.SomePtr(true),
										RunAsUser:                assert.SomePtr(int64(3000)),
										ReadOnlyRootFilesystem:   assert.SomePtr(true),
										Privileged:               assert.SomePtr(false),
										AllowPrivilegeEscalation: assert.SomePtr(false),
										Capabilities: assert.CapabilitiesAssertion{
											Drop: assert.Some([]corev1.Capability{"NET_RAW"}),
										},
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
			support.WithNamespace(t, func(env *support.Env) {
				releaseName := setup.InstallZitadel(t, env, tc.name, tc.setValues)

				if tc.zitadel != nil {
					env.AssertPartial(t, releaseName, *tc.zitadel)
				}
				if tc.login != nil {
					env.AssertPartial(t, releaseName+"-login", *tc.login)
				}
			})
		})
	}
}
