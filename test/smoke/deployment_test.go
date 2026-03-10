package smoke_test_test

import (
	"testing"

	"github.com/zitadel/zitadel-charts/test/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

func TestDeploymentLabels(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)
	chartPath := setup.ChartPath(t)

	testCases := []struct {
		name      string
		setValues map[string]string
		zitadel   *assert.DeploymentAssertion
		login     *assert.DeploymentAssertion
	}{
		{
			name: "labels",
			setValues: map[string]string{
				"login.enabled": "true",
			},
			zitadel: &assert.DeploymentAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Some(map[string]string{
						"app.kubernetes.io/name":       "zitadel",
						"app.kubernetes.io/version":    "v4.10.1",
						"app.kubernetes.io/managed-by": "Helm",
						"app.kubernetes.io/component":  "start",
					}),
				},
				Spec: assert.DeploymentSpecAssertion{
					Selector: assert.LabelSelectorAssertion{
						MatchLabels: assert.Some(map[string]string{
							"app.kubernetes.io/name":      "zitadel",
							"app.kubernetes.io/component": "start",
						}),
					},
					Template: assert.PodTemplateSpecAssertion{
						ObjectMeta: assert.ObjectMetaAssertion{
							Labels: assert.Some(map[string]string{
								"app.kubernetes.io/name":       "zitadel",
								"app.kubernetes.io/version":    "v4.10.1",
								"app.kubernetes.io/managed-by": "Helm",
								"app.kubernetes.io/component":  "start",
							}),
						},
					},
				},
			},
			login: &assert.DeploymentAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Some(map[string]string{
						"app.kubernetes.io/name":       "zitadel-login",
						"app.kubernetes.io/version":    "v4.10.1",
						"app.kubernetes.io/managed-by": "Helm",
						"app.kubernetes.io/component":  "login",
					}),
				},
				Spec: assert.DeploymentSpecAssertion{
					Selector: assert.LabelSelectorAssertion{
						MatchLabels: assert.Some(map[string]string{
							"app.kubernetes.io/name":      "zitadel-login",
							"app.kubernetes.io/component": "login",
						}),
					},
					Template: assert.PodTemplateSpecAssertion{
						ObjectMeta: assert.ObjectMetaAssertion{
							Labels: assert.Some(map[string]string{
								"app.kubernetes.io/name":       "zitadel-login",
								"app.kubernetes.io/version":    "v4.10.1",
								"app.kubernetes.io/managed-by": "Helm",
								"app.kubernetes.io/component":  "login",
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
			support.WithNamespace(t, cluster, func(env *support.Env) {
				releaseName := setup.InstallZitadel(t, env, chartPath, tc.name, tc.setValues)

				if tc.zitadel != nil {
					assert.AssertPartial(t, env.GetDeployment(t, releaseName), *tc.zitadel, releaseName)
				}
				if tc.login != nil {
					assert.AssertPartial(t, env.GetDeployment(t, releaseName+"-login"), *tc.login, releaseName+"-login")
				}
			})
		})
	}
}
