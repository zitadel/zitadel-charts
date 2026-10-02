package smoke_test_test

import (
	"testing"

	"github.com/mridang/wilhelm/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

// TestProbeSettings covers the tunable probe fields across the ZITADEL and Login
// deployments: timeoutSeconds and successThreshold on every workload, plus the
// optional per-probe terminationGracePeriodSeconds.
//
// Regression guard: the chart used to render only initialDelaySeconds,
// periodSeconds and failureThreshold, leaving timeoutSeconds (and later
// successThreshold / terminationGracePeriodSeconds) to Kubernetes defaults with
// no supported way to change them. The 1-second timeout default in particular is
// too tight on slow or resource-constrained nodes — a healthy ZITADEL can need
// longer than one second to answer /ready, and a failed readiness probe pulls
// the pod's only endpoint out of the Service, which under an EDS-only,
// health-check-free ingress controller surfaces as an immediate `503 no healthy
// upstream`.
//
// successThreshold is exposed only on the readiness probes, since Kubernetes
// forces it to 1 for liveness and startup. terminationGracePeriodSeconds is
// exposed only on the liveness and startup probes, since Kubernetes rejects it
// on readiness probes (which never kill the container); it defaults to null and
// is rendered only when set, so it is asserted on the override path only.
// Defaults deliberately match the Kubernetes defaults so existing installs are
// unaffected; this test pins both the default and the override path.
//
// int64Ptr builds the pointer the terminationGracePeriodSeconds assertion wants.
//
//goland:noinspection DuplicatedCode
func TestProbeSettings(t *testing.T) {
	t.Parallel()

	int64Ptr := func(v int64) *int64 { return &v }

	testCases := []struct {
		name      string
		setValues map[string]string
		zitadel   *assert.DeploymentAssertion
		login     *assert.DeploymentAssertion
	}{
		{
			// Defaults match the Kubernetes defaults the chart relied on
			// implicitly before these settings existed. Rendering them explicitly
			// must not change behaviour, so timeoutSeconds and the readiness
			// successThreshold are both 1.
			name: "defaults-to-kubernetes-default",
			setValues: map[string]string{
				"login.enabled":         "true",
				"login.ingress.enabled": "true",
			},
			zitadel: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						Spec: assert.PodSpecAssertion{
							Containers: assert.Some([]assert.ContainerAssertion{
								{
									Name:          assert.Some("zitadel"),
									LivenessProbe: assert.CoreProbeAssertion{TimeoutSeconds: assert.Some(int32(1))},
									ReadinessProbe: assert.CoreProbeAssertion{
										TimeoutSeconds:   assert.Some(int32(1)),
										SuccessThreshold: assert.Some(int32(1)),
									},
									StartupProbe: assert.CoreProbeAssertion{TimeoutSeconds: assert.Some(int32(1))},
								},
							}),
						},
					},
				},
			},
			login: &assert.DeploymentAssertion{
				// login.startupProbe.enabled ships false, so the chart renders no
				// startup probe at all; only liveness and readiness are asserted.
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						Spec: assert.PodSpecAssertion{
							Containers: assert.Some([]assert.ContainerAssertion{
								{
									Name:          assert.Some("zitadel-login"),
									LivenessProbe: assert.CoreProbeAssertion{TimeoutSeconds: assert.Some(int32(1))},
									ReadinessProbe: assert.CoreProbeAssertion{
										TimeoutSeconds:   assert.Some(int32(1)),
										SuccessThreshold: assert.Some(int32(1)),
									},
								},
							}),
						},
					},
				},
			},
		},
		{
			// Distinct values per probe prove each setting is wired per-probe and
			// per-workload rather than globally replaced.
			name: "custom-settings-per-probe",
			setValues: map[string]string{
				"login.enabled":                                     "true",
				"login.ingress.enabled":                             "true",
				"login.startupProbe.enabled":                        "true",
				"readinessProbe.timeoutSeconds":                     "7",
				"livenessProbe.timeoutSeconds":                      "8",
				"startupProbe.timeoutSeconds":                       "9",
				"login.readinessProbe.timeoutSeconds":               "11",
				"login.livenessProbe.timeoutSeconds":                "12",
				"login.startupProbe.timeoutSeconds":                 "13",
				"readinessProbe.successThreshold":                   "3",
				"login.readinessProbe.successThreshold":             "4",
				"livenessProbe.terminationGracePeriodSeconds":       "21",
				"startupProbe.terminationGracePeriodSeconds":        "22",
				"login.livenessProbe.terminationGracePeriodSeconds": "24",
				"login.startupProbe.terminationGracePeriodSeconds":  "25",
			},
			zitadel: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						Spec: assert.PodSpecAssertion{
							Containers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("zitadel"),
									LivenessProbe: assert.CoreProbeAssertion{
										TimeoutSeconds:                assert.Some(int32(8)),
										TerminationGracePeriodSeconds: assert.Some(int64Ptr(21)),
									},
									ReadinessProbe: assert.CoreProbeAssertion{
										TimeoutSeconds:   assert.Some(int32(7)),
										SuccessThreshold: assert.Some(int32(3)),
									},
									StartupProbe: assert.CoreProbeAssertion{
										TimeoutSeconds:                assert.Some(int32(9)),
										TerminationGracePeriodSeconds: assert.Some(int64Ptr(22)),
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
							Containers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("zitadel-login"),
									LivenessProbe: assert.CoreProbeAssertion{
										TimeoutSeconds:                assert.Some(int32(12)),
										TerminationGracePeriodSeconds: assert.Some(int64Ptr(24)),
									},
									ReadinessProbe: assert.CoreProbeAssertion{
										TimeoutSeconds:   assert.Some(int32(11)),
										SuccessThreshold: assert.Some(int32(4)),
									},
									StartupProbe: assert.CoreProbeAssertion{
										TimeoutSeconds:                assert.Some(int32(13)),
										TerminationGracePeriodSeconds: assert.Some(int64Ptr(25)),
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
