package smoke_test_test

import (
	"testing"

	"github.com/mridang/wilhelm/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

// TestLifecycle verifies that the lifecycle value is rendered onto both the
// ZITADEL and Login containers. Each closes its listener as soon as it receives
// SIGTERM, so a preStop sleep is the only way to keep a terminating pod serving
// until its endpoint removal has propagated during rolling updates and node
// drains. The native sleep action is used rather than an exec command because it
// needs no shell in the container image (the ZITADEL image is built from
// scratch). Distinct sleep durations per container prove each value is wired to
// its own workload rather than shared.
func TestLifecycle(t *testing.T) {
	t.Parallel()

	support.WithNamespace(t, func(env *support.Env) {
		releaseName := setup.InstallZitadel(t, env, "lifecycle", map[string]string{
			"login.enabled":                         "true",
			"login.ingress.enabled":                 "true",
			"lifecycle.preStop.sleep.seconds":       "15",
			"login.lifecycle.preStop.sleep.seconds": "20",
		})

		env.AssertPartial(t, releaseName, assert.DeploymentAssertion{
			Spec: assert.DeploymentSpecAssertion{
				Template: assert.PodTemplateSpecAssertion{
					Spec: assert.PodSpecAssertion{
						Containers: assert.Some([]assert.ContainerAssertion{
							{
								Name: assert.Some("zitadel"),
								Lifecycle: assert.LifecycleAssertion{
									PreStop: assert.LifecycleHandlerAssertion{
										Sleep: assert.SleepActionAssertion{Seconds: assert.Some(int64(15))},
									},
								},
							},
						}),
					},
				},
			},
		})

		env.AssertPartial(t, releaseName+"-login", assert.DeploymentAssertion{
			Spec: assert.DeploymentSpecAssertion{
				Template: assert.PodTemplateSpecAssertion{
					Spec: assert.PodSpecAssertion{
						Containers: assert.Some([]assert.ContainerAssertion{
							{
								Name: assert.Some("zitadel-login"),
								Lifecycle: assert.LifecycleAssertion{
									PreStop: assert.LifecycleHandlerAssertion{
										Sleep: assert.SleepActionAssertion{Seconds: assert.Some(int64(20))},
									},
								},
							},
						}),
					},
				},
			},
		})
	})
}
