package smoke_test_test

import (
	"testing"

	"github.com/onsi/gomega"
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
					Labels: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel"),
						gomega.HaveKeyWithValue("app.kubernetes.io/managed-by", "Helm"),
						gomega.HaveKeyWithValue("app.kubernetes.io/version", gomega.MatchRegexp(`^(v?\d+\.\d+\.\d+|[0-9a-f]{7,40})`)),
					)),
				},
			},
			binding: &assert.RoleBindingAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel"),
						gomega.HaveKeyWithValue("app.kubernetes.io/managed-by", "Helm"),
						gomega.HaveKeyWithValue("app.kubernetes.io/version", gomega.MatchRegexp(`^(v?\d+\.\d+\.\d+|[0-9a-f]{7,40})`)),
					)),
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
