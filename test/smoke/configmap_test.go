package smoke_test_test

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/zitadel/zitadel-charts/test/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

func TestConfigMapMatrix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setValues map[string]string
		zitadel   *assert.ConfigMapAssertion
		login     *assert.ConfigMapAssertion
	}{
		{
			name: "both-enabled-default",
			setValues: map[string]string{
				"login.enabled": "true",
			},
			zitadel: &assert.ConfigMapAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"helm.sh/hook":               "pre-install,pre-upgrade",
						"helm.sh/hook-delete-policy": "before-hook-creation",
						"helm.sh/hook-weight":        "0",
					}),
				},
			},
			login: &assert.ConfigMapAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"helm.sh/hook":               "pre-install,pre-upgrade",
						"helm.sh/hook-delete-policy": "before-hook-creation",
						"helm.sh/hook-weight":        "0",
					}),
				},
			},
		},
		{
			name: "both-enabled-with-annotations",
			setValues: map[string]string{
				"configMap.annotations.owner":      "platform-team",
				"login.enabled":                    "true",
				"login.configMap.annotations.team": "frontend",
			},
			zitadel: &assert.ConfigMapAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"owner":                      "platform-team",
						"helm.sh/hook":               "pre-install,pre-upgrade",
						"helm.sh/hook-delete-policy": "before-hook-creation",
						"helm.sh/hook-weight":        "0",
					}),
				},
			},
			login: &assert.ConfigMapAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"team":                       "frontend",
						"helm.sh/hook":               "pre-install,pre-upgrade",
						"helm.sh/hook-delete-policy": "before-hook-creation",
						"helm.sh/hook-weight":        "0",
					}),
				},
			},
		},
		{
			name: "zitadel-only-login-disabled",
			setValues: map[string]string{
				"configMap.annotations.config-version": "v2",
				"login.enabled":                        "false",
			},
			zitadel: &assert.ConfigMapAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"config-version":             "v2",
						"helm.sh/hook":               "pre-install,pre-upgrade",
						"helm.sh/hook-delete-policy": "before-hook-creation",
						"helm.sh/hook-weight":        "0",
					}),
				},
			},
		},
		{
			// In DSN mode the chart strips the entire Database section out of
			// the rendered configmap. ZITADEL picks up its database connection
			// from the ZITADEL_DATABASE_POSTGRES_DSN env var instead. The
			// install helper detects DSN mode and skips its hardcoded
			// Database.Postgres.* set values, so the chart sees only the DSN.
			name: "dsn-mode-strips-database",
			setValues: map[string]string{
				"login.enabled": "true",
				"env[0].name":   "ZITADEL_DATABASE_POSTGRES_DSN",
				"env[0].value":  "host=db-postgresql port=5432 user=postgres dbname=zitadel sslmode=disable",
			},
			zitadel: &assert.ConfigMapAssertion{
				Data: assert.Matching[map[string]string](gomega.And(
					gomega.HaveKeyWithValue("zitadel-config-yaml",
						gomega.Not(gomega.ContainSubstring("Database:"))),
					gomega.HaveKeyWithValue("zitadel-config-yaml",
						gomega.Not(gomega.ContainSubstring("Postgres:"))),
				)),
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
					env.AssertPartial(t, releaseName+"-config-yaml", *tc.zitadel)
				}
				if tc.login != nil {
					env.AssertPartial(t, releaseName+"-login-config-dotenv", *tc.login)
				}
			})
		})
	}
}
