package smoke_test_test

import (
	"testing"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"

	"github.com/zitadel/zitadel-charts/test/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

func TestAutoscalingMatrix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		setValues         map[string]string
		zitadel           *assert.HorizontalPodAutoscalerAssertion
		login             *assert.HorizontalPodAutoscalerAssertion
		zitadelDeployment *assert.DeploymentAssertion
		loginDeployment   *assert.DeploymentAssertion
	}{
		{
			name: "both-enabled-cpu-only",
			setValues: map[string]string{
				"zitadel.enabled":               "true",
				"zitadel.autoscaling.enabled":   "true",
				"zitadel.autoscaling.targetCPU": "60",

				"login.enabled":               "true",
				"login.autoscaling.enabled":   "true",
				"login.autoscaling.targetCPU": "55",
			},
			zitadel: &assert.HorizontalPodAutoscalerAssertion{
				Spec: assert.HorizontalPodAutoscalerSpecAssertion{
					ScaleTargetRef: assert.CrossVersionObjectReferenceAssertion{
						Kind:       assert.Some("Deployment"),
						APIVersion: assert.Some("apps/v1"),
					},
					MinReplicas: assert.SomePtr(int32(3)),
					MaxReplicas: assert.Some(int32(10)),
					Metrics: assert.Some([]assert.MetricSpecAssertion{
						{
							Type: assert.Some(autoscalingv2.ResourceMetricSourceType),
							Resource: assert.ResourceMetricSourceAssertion{
								Name: assert.Some(corev1.ResourceName("cpu")),
								Target: assert.MetricTargetAssertion{
									Type:               assert.Some(autoscalingv2.UtilizationMetricType),
									AverageUtilization: assert.SomePtr(int32(60)),
								},
							},
						},
					}),
				},
			},
			login: &assert.HorizontalPodAutoscalerAssertion{
				Spec: assert.HorizontalPodAutoscalerSpecAssertion{
					ScaleTargetRef: assert.CrossVersionObjectReferenceAssertion{
						Kind:       assert.Some("Deployment"),
						APIVersion: assert.Some("apps/v1"),
					},
					MinReplicas: assert.SomePtr(int32(3)),
					MaxReplicas: assert.Some(int32(10)),
					Metrics: assert.Some([]assert.MetricSpecAssertion{
						{
							Type: assert.Some(autoscalingv2.ResourceMetricSourceType),
							Resource: assert.ResourceMetricSourceAssertion{
								Name: assert.Some(corev1.ResourceName("cpu")),
								Target: assert.MetricTargetAssertion{
									Type:               assert.Some(autoscalingv2.UtilizationMetricType),
									AverageUtilization: assert.SomePtr(int32(55)),
								},
							},
						},
					}),
				},
			},
		},
		{
			name: "zitadel-enabled-mem-only-login-disabled",
			setValues: map[string]string{
				"zitadel.enabled":                  "true",
				"zitadel.autoscaling.enabled":      "true",
				"zitadel.autoscaling.targetMemory": "70",

				"login.enabled":             "true",
				"login.autoscaling.enabled": "false",
				"login.replicaCount":        "2",
			},
			zitadel: &assert.HorizontalPodAutoscalerAssertion{
				Spec: assert.HorizontalPodAutoscalerSpecAssertion{
					ScaleTargetRef: assert.CrossVersionObjectReferenceAssertion{
						Kind:       assert.Some("Deployment"),
						APIVersion: assert.Some("apps/v1"),
					},
					MinReplicas: assert.SomePtr(int32(3)),
					MaxReplicas: assert.Some(int32(10)),
					Metrics: assert.Some([]assert.MetricSpecAssertion{
						{
							Type: assert.Some(autoscalingv2.ResourceMetricSourceType),
							Resource: assert.ResourceMetricSourceAssertion{
								Name: assert.Some(corev1.ResourceName("memory")),
								Target: assert.MetricTargetAssertion{
									Type:               assert.Some(autoscalingv2.UtilizationMetricType),
									AverageUtilization: assert.SomePtr(int32(70)),
								},
							},
						},
					}),
				},
			},
			loginDeployment: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Replicas: assert.SomePtr(int32(2)),
				},
			},
		},
		{
			name: "both-enabled-with-annotations-and-behavior",
			setValues: map[string]string{
				"zitadel.enabled":                                                   "true",
				"zitadel.autoscaling.enabled":                                       "true",
				"zitadel.autoscaling.targetCPU":                                     "60",
				"zitadel.autoscaling.annotations.team":                              "platform",
				"zitadel.autoscaling.behavior.scaleDown.stabilizationWindowSeconds": "300",

				"login.enabled":                       "true",
				"login.autoscaling.enabled":           "true",
				"login.autoscaling.targetCPU":         "50",
				"login.autoscaling.annotations.owner": "iam",
				"login.autoscaling.behavior.scaleDown.stabilizationWindowSeconds": "300",
			},
			zitadel: &assert.HorizontalPodAutoscalerAssertion{
				Spec: assert.HorizontalPodAutoscalerSpecAssertion{
					ScaleTargetRef: assert.CrossVersionObjectReferenceAssertion{
						Kind:       assert.Some("Deployment"),
						APIVersion: assert.Some("apps/v1"),
					},
					MinReplicas: assert.SomePtr(int32(3)),
					MaxReplicas: assert.Some(int32(10)),
					Metrics: assert.Some([]assert.MetricSpecAssertion{
						{
							Type: assert.Some(autoscalingv2.ResourceMetricSourceType),
							Resource: assert.ResourceMetricSourceAssertion{
								Name: assert.Some(corev1.ResourceName("cpu")),
								Target: assert.MetricTargetAssertion{
									Type:               assert.Some(autoscalingv2.UtilizationMetricType),
									AverageUtilization: assert.SomePtr(int32(60)),
								},
							},
						},
					}),
					Behavior: assert.HorizontalPodAutoscalerBehaviorAssertion{
						ScaleUp: assert.HPAScalingRulesAssertion{
							StabilizationWindowSeconds: assert.SomePtr(int32(0)),
							SelectPolicy:               assert.SomePtr(autoscalingv2.MaxChangePolicySelect),
							Policies: assert.Some([]assert.HPAScalingPolicyAssertion{
								{Type: assert.Some(autoscalingv2.PodsScalingPolicy), Value: assert.Some(int32(4)), PeriodSeconds: assert.Some(int32(15))},
								{Type: assert.Some(autoscalingv2.PercentScalingPolicy), Value: assert.Some(int32(100)), PeriodSeconds: assert.Some(int32(15))},
							}),
						},
						ScaleDown: assert.HPAScalingRulesAssertion{
							StabilizationWindowSeconds: assert.SomePtr(int32(300)),
							SelectPolicy:               assert.SomePtr(autoscalingv2.MaxChangePolicySelect),
							Policies: assert.Some([]assert.HPAScalingPolicyAssertion{
								{Type: assert.Some(autoscalingv2.PercentScalingPolicy), Value: assert.Some(int32(100)), PeriodSeconds: assert.Some(int32(15))},
							}),
						},
					},
				},
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"team": "platform",
					}),
				},
			},
			login: &assert.HorizontalPodAutoscalerAssertion{
				Spec: assert.HorizontalPodAutoscalerSpecAssertion{
					ScaleTargetRef: assert.CrossVersionObjectReferenceAssertion{
						Kind:       assert.Some("Deployment"),
						APIVersion: assert.Some("apps/v1"),
					},
					MinReplicas: assert.SomePtr(int32(3)),
					MaxReplicas: assert.Some(int32(10)),
					Metrics: assert.Some([]assert.MetricSpecAssertion{
						{
							Type: assert.Some(autoscalingv2.ResourceMetricSourceType),
							Resource: assert.ResourceMetricSourceAssertion{
								Name: assert.Some(corev1.ResourceName("cpu")),
								Target: assert.MetricTargetAssertion{
									Type:               assert.Some(autoscalingv2.UtilizationMetricType),
									AverageUtilization: assert.SomePtr(int32(50)),
								},
							},
						},
					}),
					Behavior: assert.HorizontalPodAutoscalerBehaviorAssertion{
						ScaleUp: assert.HPAScalingRulesAssertion{
							StabilizationWindowSeconds: assert.SomePtr(int32(0)),
							SelectPolicy:               assert.SomePtr(autoscalingv2.MaxChangePolicySelect),
							Policies: assert.Some([]assert.HPAScalingPolicyAssertion{
								{Type: assert.Some(autoscalingv2.PodsScalingPolicy), Value: assert.Some(int32(4)), PeriodSeconds: assert.Some(int32(15))},
								{Type: assert.Some(autoscalingv2.PercentScalingPolicy), Value: assert.Some(int32(100)), PeriodSeconds: assert.Some(int32(15))},
							}),
						},
						ScaleDown: assert.HPAScalingRulesAssertion{
							StabilizationWindowSeconds: assert.SomePtr(int32(300)),
							SelectPolicy:               assert.SomePtr(autoscalingv2.MaxChangePolicySelect),
							Policies: assert.Some([]assert.HPAScalingPolicyAssertion{
								{Type: assert.Some(autoscalingv2.PercentScalingPolicy), Value: assert.Some(int32(100)), PeriodSeconds: assert.Some(int32(15))},
							}),
						},
					},
				},
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"owner": "iam",
					}),
				},
			},
		},
		{
			name: "both-disabled-replicas-set",
			setValues: map[string]string{
				"zitadel.enabled":             "true",
				"zitadel.autoscaling.enabled": "false",
				"replicaCount":                "2",

				"login.enabled":             "true",
				"login.autoscaling.enabled": "false",
				"login.replicaCount":        "2",
			},
			zitadelDeployment: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Replicas: assert.SomePtr(int32(2)),
				},
			},
			loginDeployment: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Replicas: assert.SomePtr(int32(2)),
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
				} else if tc.zitadelDeployment != nil {
					env.AssertPartial(t, releaseName, *tc.zitadelDeployment)
				}

				if tc.login != nil {
					env.AssertPartial(t, releaseName+"-login", *tc.login)
				} else if tc.loginDeployment != nil {
					env.AssertPartial(t, releaseName+"-login", *tc.loginDeployment)
				}
			})
		})
	}
}
