package smoke_test_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/zitadel/zitadel-charts/test/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

func TestServiceMatrix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setValues map[string]string
		zitadel   *assert.ServiceAssertion
		login     *assert.ServiceAssertion
	}{
		{
			name: "both-enabled-default-clusterip",
			setValues: map[string]string{
				"login.enabled": "true",
			},
			zitadel: &assert.ServiceAssertion{
				Spec: assert.ServiceSpecAssertion{
					Type: assert.Some(corev1.ServiceTypeClusterIP),
					Ports: assert.Some([]assert.ServicePortAssertion{
						{
							Port:     assert.Some(int32(8080)),
							Protocol: assert.Some(corev1.ProtocolTCP),
							TargetPort: assert.IntOrStringAssertion{
								Type:   assert.Some(intstr.Int),
								IntVal: assert.Some(int32(8080)),
							},
						},
					}),
				},
			},
			login: &assert.ServiceAssertion{
				Spec: assert.ServiceSpecAssertion{
					Type: assert.Some(corev1.ServiceTypeClusterIP),
					Ports: assert.Some([]assert.ServicePortAssertion{
						{
							Port:     assert.Some(int32(3000)),
							Protocol: assert.Some(corev1.ProtocolTCP),
							TargetPort: assert.IntOrStringAssertion{
								Type:   assert.Some(intstr.Int),
								IntVal: assert.Some(int32(3000)),
							},
						},
					}),
				},
			},
		},
		{
			name: "both-enabled-custom-ports",
			setValues: map[string]string{
				"service.port":       "9090",
				"login.enabled":      "true",
				"login.service.port": "9091",
			},
			zitadel: &assert.ServiceAssertion{
				Spec: assert.ServiceSpecAssertion{
					Type: assert.Some(corev1.ServiceTypeClusterIP),
					Ports: assert.Some([]assert.ServicePortAssertion{
						{
							Port:     assert.Some(int32(9090)),
							Protocol: assert.Some(corev1.ProtocolTCP),
							TargetPort: assert.IntOrStringAssertion{
								Type:   assert.Some(intstr.Int),
								IntVal: assert.Some(int32(8080)),
							},
						},
					}),
				},
			},
			login: &assert.ServiceAssertion{
				Spec: assert.ServiceSpecAssertion{
					Type: assert.Some(corev1.ServiceTypeClusterIP),
					Ports: assert.Some([]assert.ServicePortAssertion{
						{
							Port:     assert.Some(int32(9091)),
							Protocol: assert.Some(corev1.ProtocolTCP),
							TargetPort: assert.IntOrStringAssertion{
								Type:   assert.Some(intstr.Int),
								IntVal: assert.Some(int32(3000)),
							},
						},
					}),
				},
			},
		},
		{
			name: "both-enabled-with-annotations",
			setValues: map[string]string{
				"service.annotations.cloud\\.google\\.com/load-balancer-type": "Internal",
				"service.annotations.owner":                                   "platform-team",
				"login.enabled":                                               "true",
				"login.service.annotations.service\\.beta\\.kubernetes\\.io/aws-load-balancer-internal": "yes",
			},
			zitadel: &assert.ServiceAssertion{
				Spec: assert.ServiceSpecAssertion{
					Type: assert.Some(corev1.ServiceTypeClusterIP),
					Ports: assert.Some([]assert.ServicePortAssertion{
						{
							Port:     assert.Some(int32(8080)),
							Protocol: assert.Some(corev1.ProtocolTCP),
							TargetPort: assert.IntOrStringAssertion{
								Type:   assert.Some(intstr.Int),
								IntVal: assert.Some(int32(8080)),
							},
						},
					}),
				},
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"cloud.google.com/load-balancer-type": "Internal",
						"owner":                               "platform-team",
					}),
				},
			},
			login: &assert.ServiceAssertion{
				Spec: assert.ServiceSpecAssertion{
					Type: assert.Some(corev1.ServiceTypeClusterIP),
					Ports: assert.Some([]assert.ServicePortAssertion{
						{
							Port:     assert.Some(int32(3000)),
							Protocol: assert.Some(corev1.ProtocolTCP),
							TargetPort: assert.IntOrStringAssertion{
								Type:   assert.Some(intstr.Int),
								IntVal: assert.Some(int32(3000)),
							},
						},
					}),
				},
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Some(map[string]string{
						"service.beta.kubernetes.io/aws-load-balancer-internal": "yes",
					}),
				},
			},
		},
		{
			name: "zitadel-only-login-disabled",
			setValues: map[string]string{
				"service.type":  "ClusterIP",
				"service.port":  "8888",
				"login.enabled": "false",
			},
			zitadel: &assert.ServiceAssertion{
				Spec: assert.ServiceSpecAssertion{
					Type: assert.Some(corev1.ServiceTypeClusterIP),
					Ports: assert.Some([]assert.ServicePortAssertion{
						{
							Port:     assert.Some(int32(8888)),
							Protocol: assert.Some(corev1.ProtocolTCP),
							TargetPort: assert.IntOrStringAssertion{
								Type:   assert.Some(intstr.Int),
								IntVal: assert.Some(int32(8080)),
							},
						},
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
