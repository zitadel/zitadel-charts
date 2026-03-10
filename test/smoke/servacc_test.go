package smoke_test_test

import (
	"testing"

	"github.com/zitadel/zitadel-charts/test/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

func TestServiceAccountMatrix(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)
	chartPath := setup.ChartPath(t)

	testCases := []struct {
		name      string
		setValues map[string]string
		zitadel   *assert.ServiceAccountAssertion
		login     *assert.ServiceAccountAssertion
	}{
		{
			name: "both-enabled-default",
			setValues: map[string]string{
				"serviceAccount.create":       "true",
				"login.enabled":               "true",
				"login.serviceAccount.create": "true",
			},
			zitadel: &assert.ServiceAccountAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"helm.sh/hook":               "pre-install,pre-upgrade",
						"helm.sh/hook-delete-policy": "before-hook-creation",
						"helm.sh/hook-weight":        "0",
					}),
				},
			},
			login: &assert.ServiceAccountAssertion{
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
				"serviceAccount.create": "true",
				"serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn":       "arn:aws:iam::123456789012:role/zitadel-role",
				"serviceAccount.annotations.owner":                                "platform-team",
				"login.enabled":                                                   "true",
				"login.serviceAccount.create":                                     "true",
				"login.serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn": "arn:aws:iam::123456789012:role/login-role",
				"login.serviceAccount.annotations.team":                           "frontend",
			},
			zitadel: &assert.ServiceAccountAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/zitadel-role",
						"owner":                      "platform-team",
						"helm.sh/hook":               "pre-install,pre-upgrade",
						"helm.sh/hook-delete-policy": "before-hook-creation",
						"helm.sh/hook-weight":        "0",
					}),
				},
			},
			login: &assert.ServiceAccountAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/login-role",
						"team":                       "frontend",
						"helm.sh/hook":               "pre-install,pre-upgrade",
						"helm.sh/hook-delete-policy": "before-hook-creation",
						"helm.sh/hook-weight":        "0",
					}),
				},
			},
		},
		{
			name: "zitadel-enabled-login-disabled",
			setValues: map[string]string{
				"serviceAccount.create":                        "true",
				"serviceAccount.annotations.workload-identity": "enabled",
				"login.enabled":                                "true",
				"login.serviceAccount.create":                  "false",
			},
			zitadel: &assert.ServiceAccountAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"workload-identity":          "enabled",
						"helm.sh/hook":               "pre-install,pre-upgrade",
						"helm.sh/hook-delete-policy": "before-hook-creation",
						"helm.sh/hook-weight":        "0",
					}),
				},
			},
		},
		{
			name: "both-disabled",
			setValues: map[string]string{
				"serviceAccount.create":       "false",
				"login.enabled":               "true",
				"login.serviceAccount.create": "false",
			},
		},
		{
			name: "zitadel-only-with-gcp-workload-identity",
			setValues: map[string]string{
				"serviceAccount.create": "true",
				"serviceAccount.annotations.iam\\.gke\\.io/gcp-service-account": "zitadel@project.iam.gserviceaccount.com",
				"login.enabled": "false",
			},
			zitadel: &assert.ServiceAccountAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"iam.gke.io/gcp-service-account": "zitadel@project.iam.gserviceaccount.com",
						"helm.sh/hook":                   "pre-install,pre-upgrade",
						"helm.sh/hook-delete-policy":     "before-hook-creation",
						"helm.sh/hook-weight":            "0",
					}),
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			support.WithNamespace(t, cluster, func(env *support.Env) {
				releaseName := setup.InstallZitadel(t, env, chartPath, tc.name, tc.setValues)

				if tc.zitadel != nil {
					assert.AssertPartial(t, env.GetServiceAccount(t, releaseName), *tc.zitadel, releaseName)
				}
				if tc.login != nil {
					assert.AssertPartial(t, env.GetServiceAccount(t, releaseName+"-login"), *tc.login, releaseName+"-login")
				}
			})
		})
	}
}
