package smoke_test_test

import (
	"testing"

	"github.com/mridang/wilhelm/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

// TestLifecycle verifies that the lifecycle value is rendered onto the ZITADEL
// container. ZITADEL closes its listener as soon as it receives SIGTERM, so a
// preStop sleep is the only way to keep a terminating pod serving until its
// endpoint removal has propagated. The image has no shell, which is why the
// test uses the native sleep action rather than an exec command.
func TestLifecycle(t *testing.T) {
	t.Parallel()

	support.WithNamespace(t, func(env *support.Env) {
		releaseName := setup.InstallZitadel(t, env, "lifecycle", map[string]string{
			"lifecycle.preStop.sleep.seconds": "15",
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
	})
}
