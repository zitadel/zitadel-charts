package smoke_test_test

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/zitadel/zitadel-charts/test/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

func TestPodDisruptionBudgetMatrix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setValues map[string]string
		zitadel   *assert.PodDisruptionBudgetAssertion
		login     *assert.PodDisruptionBudgetAssertion
	}{
		{
			name: "both-enabled-minAvailable-int",
			setValues: map[string]string{
				"pdb.enabled":      "true",
				"pdb.minAvailable": "2",

				"login.enabled":          "true",
				"login.pdb.enabled":      "true",
				"login.pdb.minAvailable": "1",
			},
			zitadel: &assert.PodDisruptionBudgetAssertion{
				Spec: assert.PodDisruptionBudgetSpecAssertion{
					MinAvailable: assert.IntOrStringAssertion{
						Type:   assert.Some(intstr.Int),
						IntVal: assert.Some(int32(2)),
					},
				},
			},
			login: &assert.PodDisruptionBudgetAssertion{
				Spec: assert.PodDisruptionBudgetSpecAssertion{
					MinAvailable: assert.IntOrStringAssertion{
						Type:   assert.Some(intstr.Int),
						IntVal: assert.Some(int32(1)),
					},
				},
			},
		},
		{
			name: "both-enabled-minAvailable-percentage",
			setValues: map[string]string{
				"pdb.enabled":      "true",
				"pdb.minAvailable": "50%",

				"login.enabled":          "true",
				"login.pdb.enabled":      "true",
				"login.pdb.minAvailable": "75%",
			},
			zitadel: &assert.PodDisruptionBudgetAssertion{
				Spec: assert.PodDisruptionBudgetSpecAssertion{
					MinAvailable: assert.IntOrStringAssertion{
						Type:   assert.Some(intstr.String),
						StrVal: assert.Some("50%"),
					},
				},
			},
			login: &assert.PodDisruptionBudgetAssertion{
				Spec: assert.PodDisruptionBudgetSpecAssertion{
					MinAvailable: assert.IntOrStringAssertion{
						Type:   assert.Some(intstr.String),
						StrVal: assert.Some("75%"),
					},
				},
			},
		},
		{
			name: "both-enabled-maxUnavailable-int",
			setValues: map[string]string{
				"pdb.enabled":        "true",
				"pdb.maxUnavailable": "1",

				"login.enabled":            "true",
				"login.pdb.enabled":        "true",
				"login.pdb.maxUnavailable": "2",
			},
			zitadel: &assert.PodDisruptionBudgetAssertion{
				Spec: assert.PodDisruptionBudgetSpecAssertion{
					MaxUnavailable: assert.IntOrStringAssertion{
						Type:   assert.Some(intstr.Int),
						IntVal: assert.Some(int32(1)),
					},
				},
			},
			login: &assert.PodDisruptionBudgetAssertion{
				Spec: assert.PodDisruptionBudgetSpecAssertion{
					MaxUnavailable: assert.IntOrStringAssertion{
						Type:   assert.Some(intstr.Int),
						IntVal: assert.Some(int32(2)),
					},
				},
			},
		},
		{
			name: "both-enabled-maxUnavailable-percentage",
			setValues: map[string]string{
				"pdb.enabled":        "true",
				"pdb.maxUnavailable": "25%",

				"login.enabled":            "true",
				"login.pdb.enabled":        "true",
				"login.pdb.maxUnavailable": "33%",
			},
			zitadel: &assert.PodDisruptionBudgetAssertion{
				Spec: assert.PodDisruptionBudgetSpecAssertion{
					MaxUnavailable: assert.IntOrStringAssertion{
						Type:   assert.Some(intstr.String),
						StrVal: assert.Some("25%"),
					},
				},
			},
			login: &assert.PodDisruptionBudgetAssertion{
				Spec: assert.PodDisruptionBudgetSpecAssertion{
					MaxUnavailable: assert.IntOrStringAssertion{
						Type:   assert.Some(intstr.String),
						StrVal: assert.Some("33%"),
					},
				},
			},
		},
		{
			name: "zitadel-enabled-login-disabled",
			setValues: map[string]string{
				"pdb.enabled":      "true",
				"pdb.minAvailable": "1",

				"login.enabled":     "true",
				"login.pdb.enabled": "false",
			},
			zitadel: &assert.PodDisruptionBudgetAssertion{
				Spec: assert.PodDisruptionBudgetSpecAssertion{
					MinAvailable: assert.IntOrStringAssertion{
						Type:   assert.Some(intstr.Int),
						IntVal: assert.Some(int32(1)),
					},
				},
			},
		},
		{
			name: "both-enabled-with-annotations",
			setValues: map[string]string{
				"pdb.enabled":           "true",
				"pdb.minAvailable":      "1",
				"pdb.annotations.team":  "platform",
				"pdb.annotations.owner": "sre",

				"login.enabled":              "true",
				"login.pdb.enabled":          "true",
				"login.pdb.minAvailable":     "1",
				"login.pdb.annotations.team": "frontend",
			},
			zitadel: &assert.PodDisruptionBudgetAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"team":  "platform",
						"owner": "sre",
					}),
				},
				Spec: assert.PodDisruptionBudgetSpecAssertion{
					MinAvailable: assert.IntOrStringAssertion{
						Type:   assert.Some(intstr.Int),
						IntVal: assert.Some(int32(1)),
					},
				},
			},
			login: &assert.PodDisruptionBudgetAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"team": "frontend",
					}),
				},
				Spec: assert.PodDisruptionBudgetSpecAssertion{
					MinAvailable: assert.IntOrStringAssertion{
						Type:   assert.Some(intstr.Int),
						IntVal: assert.Some(int32(1)),
					},
				},
			},
		},
		{
			name: "both-disabled",
			setValues: map[string]string{
				"pdb.enabled": "false",

				"login.enabled":     "true",
				"login.pdb.enabled": "false",
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
