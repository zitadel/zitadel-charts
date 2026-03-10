package smoke_test_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/zitadel/zitadel-charts/test/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

func TestSecurityContexts(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)
	chartPath := setup.ChartPath(t)

	testCases := []struct {
		name      string
		setValues map[string]string
		zitadel   assert.DeploymentAssertion
		login     assert.DeploymentAssertion
	}{
		{
			name: "defaults-use-global",
			setValues: map[string]string{
				"login.enabled":         "true",
				"login.ingress.enabled": "true",
			},
			zitadel: assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						Spec: assert.PodSpecAssertion{
							SecurityContext: assert.PodSecurityContextAssertion{
								RunAsNonRoot: assert.SomePtr(true),
								RunAsUser:    assert.SomePtr(int64(1000)),
								FSGroup:      assert.SomePtr(int64(1000)),
							},
							InitContainers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("wait-for-postgres"),
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
			login: assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						Spec: assert.PodSpecAssertion{
							SecurityContext: assert.PodSecurityContextAssertion{
								RunAsNonRoot: assert.SomePtr(true),
								RunAsUser:    assert.SomePtr(int64(1000)),
								FSGroup:      assert.SomePtr(int64(1000)),
							},
							InitContainers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("wait-for-zitadel"),
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
			zitadel: assert.DeploymentAssertion{
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
			login: assert.DeploymentAssertion{
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
			support.WithNamespace(t, cluster, func(env *support.Env) {
				releaseName := setup.InstallZitadel(t, env, chartPath, tc.name, tc.setValues)

				assert.AssertPartial(t, env.GetDeployment(t, releaseName), tc.zitadel, "zitadel deployment")
				assert.AssertPartial(t, env.GetDeployment(t, releaseName+"-login"), tc.login, "login deployment")
			})
		})
	}
}
