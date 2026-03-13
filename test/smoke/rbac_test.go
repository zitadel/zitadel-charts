package smoke_test_test

import (
	"testing"

	"github.com/zitadel/zitadel-charts/test/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

func TestRBACLabels(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		role    *assert.RoleAssertion
		binding *assert.RoleBindingAssertion
	}{
		{
			name: "labels",
			role: &assert.RoleAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Some(map[string]string{
						"app.kubernetes.io/name":       "zitadel",
						"app.kubernetes.io/version":    "v4.10.1",
						"app.kubernetes.io/managed-by": "Helm",
					}),
				},
			},
			binding: &assert.RoleBindingAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Some(map[string]string{
						"app.kubernetes.io/name":       "zitadel",
						"app.kubernetes.io/version":    "v4.10.1",
						"app.kubernetes.io/managed-by": "Helm",
					}),
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			support.WithNamespace(t, func(env *support.Env) {
				releaseName := setup.InstallZitadel(t, env, tc.name, nil)

				if tc.role != nil {
					env.AssertPartial(t, releaseName, *tc.role)
				}
				if tc.binding != nil {
					env.AssertPartial(t, releaseName, *tc.binding)
				}
			})
		})
	}
}
