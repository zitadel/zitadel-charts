package smoke_test_test

import (
	"testing"

	"github.com/mridang/wilhelm/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

// TestProbeTimeoutSeconds covers the probe `timeoutSeconds` setting across the
// ZITADEL and Login deployments.
//
// Regression guard: the chart used to render only initialDelaySeconds,
// periodSeconds and failureThreshold, leaving `timeoutSeconds` to the
// Kubernetes default of 1 second. That default is too tight on slow or
// resource-constrained nodes — a healthy ZITADEL can need longer than one
// second to answer /ready, and a failed readiness probe pulls the pod's only
// endpoint out of the Service, which under an EDS-only, health-check-free
// ingress controller surfaces as an immediate `503 no healthy upstream`.
//
// The default deliberately stays at 1 so existing installs are unaffected; this
// test pins both the default and the override path.
//
//goland:noinspection DuplicatedCode
func TestProbeTimeoutSeconds(t *testing.T) {
	t.Parallel()

	// probeTimeouts builds the assertion for a container whose three probes all
	// share the same timeout.
	probeTimeouts := func(timeout int32) (liveness, readiness, startup assert.CoreProbeAssertion) {
		p := assert.CoreProbeAssertion{TimeoutSeconds: assert.Some(timeout)}
		return p, p, p
	}

	testCases := []struct {
		name      string
		setValues map[string]string
		zitadel   *assert.DeploymentAssertion
		login     *assert.DeploymentAssertion
	}{
		{
			// Default is 1, matching the Kubernetes default the chart relied on
			// implicitly before this setting existed. Rendering it explicitly
			// must not change behaviour.
			name: "defaults-to-kubernetes-default",
			setValues: map[string]string{
				"login.enabled":         "true",
				"login.ingress.enabled": "true",
			},
			zitadel: func() *assert.DeploymentAssertion {
				liveness, readiness, startup := probeTimeouts(1)
				return &assert.DeploymentAssertion{
					Spec: assert.DeploymentSpecAssertion{
						Template: assert.PodTemplateSpecAssertion{
							Spec: assert.PodSpecAssertion{
								Containers: assert.Some([]assert.ContainerAssertion{
									{
										Name:           assert.Some("zitadel"),
										LivenessProbe:  liveness,
										ReadinessProbe: readiness,
										StartupProbe:   startup,
									},
								}),
							},
						},
					},
				}
			}(),
			login: func() *assert.DeploymentAssertion {
				// login.startupProbe.enabled ships false, so the chart renders no
				// startup probe at all; only liveness and readiness are asserted.
				liveness, readiness, _ := probeTimeouts(1)
				return &assert.DeploymentAssertion{
					Spec: assert.DeploymentSpecAssertion{
						Template: assert.PodTemplateSpecAssertion{
							Spec: assert.PodSpecAssertion{
								Containers: assert.Some([]assert.ContainerAssertion{
									{
										Name:           assert.Some("zitadel-login"),
										LivenessProbe:  liveness,
										ReadinessProbe: readiness,
									},
								}),
							},
						},
					},
				}
			}(),
		},
		{
			// Distinct values per probe prove the setting is wired per-probe and
			// per-workload rather than globally replaced.
			name: "custom-timeout-per-probe",
			setValues: map[string]string{
				"login.enabled":                       "true",
				"login.ingress.enabled":               "true",
				"login.startupProbe.enabled":          "true",
				"readinessProbe.timeoutSeconds":       "7",
				"livenessProbe.timeoutSeconds":        "8",
				"startupProbe.timeoutSeconds":         "9",
				"login.readinessProbe.timeoutSeconds": "11",
				"login.livenessProbe.timeoutSeconds":  "12",
				"login.startupProbe.timeoutSeconds":   "13",
			},
			zitadel: func() *assert.DeploymentAssertion {
				return &assert.DeploymentAssertion{
					Spec: assert.DeploymentSpecAssertion{
						Template: assert.PodTemplateSpecAssertion{
							Spec: assert.PodSpecAssertion{
								Containers: assert.Some([]assert.ContainerAssertion{
									{
										Name:           assert.Some("zitadel"),
										LivenessProbe:  assert.CoreProbeAssertion{TimeoutSeconds: assert.Some(int32(8))},
										ReadinessProbe: assert.CoreProbeAssertion{TimeoutSeconds: assert.Some(int32(7))},
										StartupProbe:   assert.CoreProbeAssertion{TimeoutSeconds: assert.Some(int32(9))},
									},
								}),
							},
						},
					},
				}
			}(),
			login: func() *assert.DeploymentAssertion {
				return &assert.DeploymentAssertion{
					Spec: assert.DeploymentSpecAssertion{
						Template: assert.PodTemplateSpecAssertion{
							Spec: assert.PodSpecAssertion{
								Containers: assert.Some([]assert.ContainerAssertion{
									{
										Name:           assert.Some("zitadel-login"),
										LivenessProbe:  assert.CoreProbeAssertion{TimeoutSeconds: assert.Some(int32(12))},
										ReadinessProbe: assert.CoreProbeAssertion{TimeoutSeconds: assert.Some(int32(11))},
										StartupProbe:   assert.CoreProbeAssertion{TimeoutSeconds: assert.Some(int32(13))},
									},
								}),
							},
						},
					},
				}
			}(),
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
