package smoke_test_test

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/mridang/wilhelm/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

//goland:noinspection ALL
func TestJobMatrix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setValues map[string]string
		setupJob  *assert.JobAssertion
	}{
		{
			// The setup Job must run the ZITADEL image and nothing else. Chart
			// versions before v11 ran kubectl sidecars whose image tag was
			// derived from the cluster version, so any cluster on a patch
			// release the upstream registry never published failed the
			// install with ImagePullBackOff (#614). Slice assertions require
			// an exact length match, so a single expected container guards
			// against a sidecar ever coming back.
			name:      "setup-job-runs-only-the-zitadel-image",
			setValues: map[string]string{},
			setupJob: &assert.JobAssertion{
				Spec: assert.JobSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						Spec: assert.PodSpecAssertion{
							Containers: assert.Some([]assert.ContainerAssertion{
								{
									Name:  assert.Some("zitadel-setup"),
									Image: assert.Matching[string](gomega.HavePrefix("ghcr.io/zitadel/zitadel:")),
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

				if tc.setupJob != nil {
					env.AssertPartial(t, releaseName+"-setup", *tc.setupJob)
				}
			})
		})
	}
}
