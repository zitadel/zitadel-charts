package smoke_test_test

import (
	"testing"

	"github.com/zitadel/zitadel-charts/test/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

//goland:noinspection DuplicatedCode
func TestIngressMatrix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setValues map[string]string
		zitadel   *assert.IngressAssertion
		login     *assert.IngressAssertion
	}{
		{
			name: "labels",
			setValues: map[string]string{
				"login.enabled":                            "true",
				"login.ingress.enabled":                    "true",
				"ingress.hosts[0].host":                    "zitadel.example.local",
				"ingress.hosts[0].paths[0].path":           "/",
				"ingress.hosts[0].paths[0].pathType":       "Prefix",
				"login.ingress.hosts[0].host":              "login.example.local",
				"login.ingress.hosts[0].paths[0].path":     "/",
				"login.ingress.hosts[0].paths[0].pathType": "Prefix",
			},
			zitadel: &assert.IngressAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Some(map[string]string{
						"app.kubernetes.io/name":       "zitadel",
						"app.kubernetes.io/version":    "v4.12.1",
						"app.kubernetes.io/managed-by": "Helm",
					}),
				},
			},
			login: &assert.IngressAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Some(map[string]string{
						"app.kubernetes.io/name":       "zitadel-login",
						"app.kubernetes.io/version":    "v4.12.1",
						"app.kubernetes.io/managed-by": "Helm",
						"app.kubernetes.io/component":  "login",
					}),
				},
			},
		},
		{
			name: "nginx-controller-injects-backend-protocol",
			setValues: map[string]string{
				"ingress.controller":                 "nginx",
				"ingress.hosts[0].host":              "zitadel.example.local",
				"ingress.hosts[0].paths[0].path":     "/",
				"ingress.hosts[0].paths[0].pathType": "Prefix",
			},
			zitadel: &assert.IngressAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"nginx.ingress.kubernetes.io/backend-protocol": "GRPC",
					}),
				},
			},
		},
		{
			name: "generic-controller-omits-backend-protocol",
			setValues: map[string]string{
				"ingress.controller":                 "generic",
				"ingress.hosts[0].host":              "zitadel.example.local",
				"ingress.hosts[0].paths[0].path":     "/",
				"ingress.hosts[0].paths[0].pathType": "Prefix",
			},
			zitadel: &assert.IngressAssertion{},
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
