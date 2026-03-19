package smoke_test_test

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/zitadel/zitadel-charts/test/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

//goland:noinspection ALL
func TestServiceMonitorDisabledByDefault(t *testing.T) {
	t.Parallel()

	support.WithNamespace(t, func(env *support.Env) {
		releaseName := setup.InstallZitadel(t, env, "sm-disabled", nil)

		env.AssertNone(t, releaseName, assert.ServiceMonitorAssertion{})
		env.AssertNone(t, releaseName+"-login", assert.ServiceMonitorAssertion{})
	})
}

//goland:noinspection ALL
func TestServiceMonitorMatrix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setValues map[string]string
		zitadel   *assert.ServiceMonitorAssertion
		login     *assert.ServiceMonitorAssertion
	}{
		{
			name: "zitadel-only",
			setValues: map[string]string{
				"metrics.enabled":                "true",
				"metrics.serviceMonitor.enabled": "true",
			},
			zitadel: &assert.ServiceMonitorAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel"),
						gomega.HaveKeyWithValue("app.kubernetes.io/managed-by", "Helm"),
					)),
				},
			},
		},
		{
			name: "login-only",
			setValues: map[string]string{
				"login.enabled":                        "true",
				"login.metrics.enabled":                "true",
				"login.metrics.serviceMonitor.enabled": "true",
			},
			login: &assert.ServiceMonitorAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel-login"),
						gomega.HaveKeyWithValue("app.kubernetes.io/managed-by", "Helm"),
						gomega.HaveKeyWithValue("app.kubernetes.io/component", "login"),
					)),
				},
			},
		},
		{
			name: "both-enabled",
			setValues: map[string]string{
				"metrics.enabled":                      "true",
				"metrics.serviceMonitor.enabled":       "true",
				"login.enabled":                        "true",
				"login.metrics.enabled":                "true",
				"login.metrics.serviceMonitor.enabled": "true",
			},
			zitadel: &assert.ServiceMonitorAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel"),
						gomega.HaveKeyWithValue("app.kubernetes.io/managed-by", "Helm"),
					)),
				},
			},
			login: &assert.ServiceMonitorAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel-login"),
						gomega.HaveKeyWithValue("app.kubernetes.io/managed-by", "Helm"),
					)),
				},
			},
		},
		{
			name: "additional-labels",
			setValues: map[string]string{
				"metrics.enabled":                            "true",
				"metrics.serviceMonitor.enabled":             "true",
				"metrics.serviceMonitor.additionalLabels.team": "platform",
			},
			zitadel: &assert.ServiceMonitorAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("team", "platform"),
						gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel"),
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
