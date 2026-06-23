package smoke_test_test

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/mridang/wilhelm/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

// validSystemUserKeyData is a base64-encoded RSA public key (PEM). ZITADEL's
// config loader base64-decodes SystemAPIUsers KeyData, so the value must be
// valid base64 for the setup job to start. The key itself is only parsed when
// that system user authenticates, which this render-focused test never does.
const validSystemUserKeyData = "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FROEFNSUlCQ2dLQ0FRRUFzRjdQWlhZUzV6TWFESHlVK1I3ZQovQlIyRUkzTkZNSWlLSWJrNFk2WXZQNHV1S3huTHlLMy9pM0t0ZUEzOE5Bb21rZTdnNXNkM1VJSDIrdGllTmVtCmlyQUlvUWwwZDZOMTUrYU5VTG9haFpFTjdnVVBjQWJXeW1OeUtYZ1BLMkVYc2lhbzFBcVFVQVdnb2UxYjZ2a08KdVdHSklIRVlhYUlmQjdoVlNQTjh5UTJha2xkYTdIU3kzK2ZpVHVBT3dxRlJqSW51OEROL1hHcC9YYTNIK2kySApCWVUvZ2syT3UvMHFxbDRXY0ZxbVJVQjdKRzZFZEZNSStzUG5iOEkzWTFSazV0Q1Y3TDk0bjhJdVg3MW5EUXdzCjJ6THlpRGFjQkxsWGdyZmx1ZFB1akNMelVtTDIxMlVIeDBtT0UvSm5JdTBzaU54ejEzRXhOcFljc3lkcE1mMFMKWlFJREFRQUIKLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg=="

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
				Data: assert.Matching[map[string]string](gomega.And(
					gomega.HaveKeyWithValue("zitadel-config-yaml",
						gomega.ContainSubstring("SystemAPIUsers")),
					gomega.HaveKeyWithValue("zitadel-config-yaml",
						gomega.ContainSubstring("login-client")),
					gomega.HaveKeyWithValue("zitadel-config-yaml",
						gomega.ContainSubstring("IAM_LOGIN_CLIENT")),
				)),
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
			// Regression for https://github.com/zitadel/zitadel-charts/issues/602:
			// a user-supplied SystemAPIUsers list (dash form) must merge with the
			// chart's login-client entry rather than overwriting it. Both the
			// operator's "superuser" and the generated "login-client" must appear.
			name: "user-systemapiusers-list-merges-login-client",
			setValues: map[string]string{
				"login.enabled": "true",
				"zitadel.configmapConfig.SystemAPIUsers[0].superuser.KeyData": validSystemUserKeyData,
			},
			zitadel: &assert.ConfigMapAssertion{
				Data: assert.Matching[map[string]string](gomega.And(
					gomega.HaveKeyWithValue("zitadel-config-yaml",
						gomega.ContainSubstring("login-client")),
					gomega.HaveKeyWithValue("zitadel-config-yaml",
						gomega.ContainSubstring("IAM_LOGIN_CLIENT")),
					gomega.HaveKeyWithValue("zitadel-config-yaml",
						gomega.ContainSubstring("superuser")),
				)),
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
				Data: assert.Matching[map[string]string](gomega.And(
					gomega.HaveKeyWithValue("zitadel-config-yaml",
						gomega.Not(gomega.ContainSubstring("SystemAPIUsers"))),
				)),
			},
		},
		{
			name: "x509-login-env-vars",
			setValues: map[string]string{
				"login.enabled": "true",
			},
			login: &assert.ConfigMapAssertion{
				Data: assert.Matching[map[string]string](gomega.And(
					gomega.HaveKeyWithValue(".env",
						gomega.ContainSubstring("ZITADEL_LOGINCLIENT_KEYFILE")),
					gomega.HaveKeyWithValue(".env",
						gomega.ContainSubstring("AUDIENCE")),
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
