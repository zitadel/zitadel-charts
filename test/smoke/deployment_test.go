package smoke_test_test

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mridang/wilhelm/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

//goland:noinspection ALL
func TestDeploymentMatrix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		setValues  map[string]string
		preInstall func(t *testing.T, env *support.Env)
		zitadel    *assert.DeploymentAssertion
		login      *assert.DeploymentAssertion
	}{
		{
			name: "defaults",
			setValues: map[string]string{
				"login.enabled":         "true",
				"login.ingress.enabled": "true",
			},
			zitadel: &assert.DeploymentAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel"),
						gomega.HaveKeyWithValue("app.kubernetes.io/managed-by", "Helm"),
						gomega.HaveKeyWithValue("app.kubernetes.io/component", "start"),
						gomega.HaveKeyWithValue("app.kubernetes.io/version", gomega.MatchRegexp(`^(v?\d+\.\d+\.\d+|[0-9a-f]{7,40})`)),
					)),
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
							Labels: assert.Matching[map[string]string](gomega.And(
								gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel"),
								gomega.HaveKeyWithValue("app.kubernetes.io/managed-by", "Helm"),
								gomega.HaveKeyWithValue("app.kubernetes.io/component", "start"),
								gomega.HaveKeyWithValue("app.kubernetes.io/version", gomega.MatchRegexp(`^(v?\d+\.\d+\.\d+|[0-9a-f]{7,40})`)),
							)),
						},
						Spec: assert.PodSpecAssertion{
							SecurityContext: assert.PodSecurityContextAssertion{
								RunAsNonRoot: assert.SomePtr(true),
								RunAsUser:    assert.SomePtr(int64(1000)),
								FSGroup:      assert.SomePtr(int64(1000)),
							},
							Containers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("zitadel"),
									SecurityContext: assert.SecurityContextAssertion{
										RunAsNonRoot:           assert.SomePtr(true),
										RunAsUser:              assert.SomePtr(int64(1000)),
										ReadOnlyRootFilesystem: assert.SomePtr(true),
										Privileged:             assert.SomePtr(false),
									},
								},
							}),
						},
					},
				},
			},
			login: &assert.DeploymentAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel-login"),
						gomega.HaveKeyWithValue("app.kubernetes.io/managed-by", "Helm"),
						gomega.HaveKeyWithValue("app.kubernetes.io/component", "login"),
						gomega.HaveKeyWithValue("app.kubernetes.io/version", gomega.MatchRegexp(`^(v?\d+\.\d+\.\d+|[0-9a-f]{7,40})`)),
					)),
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
							Labels: assert.Matching[map[string]string](gomega.And(
								gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel-login"),
								gomega.HaveKeyWithValue("app.kubernetes.io/managed-by", "Helm"),
								gomega.HaveKeyWithValue("app.kubernetes.io/component", "login"),
								gomega.HaveKeyWithValue("app.kubernetes.io/version", gomega.MatchRegexp(`^(v?\d+\.\d+\.\d+|[0-9a-f]{7,40})`)),
							)),
						},
						Spec: assert.PodSpecAssertion{
							SecurityContext: assert.PodSecurityContextAssertion{
								RunAsNonRoot: assert.SomePtr(true),
								RunAsUser:    assert.SomePtr(int64(1000)),
								FSGroup:      assert.SomePtr(int64(1000)),
							},
							Containers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("zitadel-login"),
									SecurityContext: assert.SecurityContextAssertion{
										RunAsNonRoot:           assert.SomePtr(true),
										RunAsUser:              assert.SomePtr(int64(1000)),
										ReadOnlyRootFilesystem: assert.SomePtr(true),
										Privileged:             assert.SomePtr(false),
									},
								},
							}),
						},
					},
				},
			},
		},
		{
			name: "login-metrics-annotations",
			setValues: map[string]string{
				"login.enabled":         "true",
				"login.metrics.enabled": "true",
			},
			login: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						ObjectMeta: assert.ObjectMetaAssertion{
							Annotations: assert.Matching[map[string]string](gomega.And(
								gomega.HaveKeyWithValue("prometheus.io/scrape", "true"),
								gomega.HaveKeyWithValue("prometheus.io/path", "/metrics"),
								gomega.HaveKeyWithValue("prometheus.io/port", "9464"),
							)),
						},
					},
				},
			},
		},
		{
			name: "login-metrics-disabled-no-annotations",
			setValues: map[string]string{
				"login.enabled": "true",
			},
			login: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						ObjectMeta: assert.ObjectMetaAssertion{
							Annotations: assert.Matching[map[string]string](gomega.And(
								gomega.Not(gomega.HaveKey("prometheus.io/scrape")),
								gomega.Not(gomega.HaveKey("prometheus.io/path")),
								gomega.Not(gomega.HaveKey("prometheus.io/port")),
							)),
						},
					},
				},
			},
		},
		{
			name: "service-key-checksum-annotations",
			setValues: map[string]string{
				"login.enabled": "true",
			},
			zitadel: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						ObjectMeta: assert.ObjectMetaAssertion{
							Annotations: assert.Matching[map[string]string](gomega.And(
								gomega.HaveKey("checksum/secret-login-service-key"),
								gomega.HaveKey("checksum/secret-admin-service-key"),
							)),
						},
					},
				},
			},
			login: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						ObjectMeta: assert.ObjectMetaAssertion{
							Annotations: assert.Matching[map[string]string](gomega.And(
								gomega.HaveKey("checksum/secret-login-service-key"),
							)),
						},
					},
				},
			},
		},
		{
			name: "login-external-secret-no-checksum-annotation",
			setValues: map[string]string{
				"login.enabled":                   "true",
				"login.loginServiceKeySecretName": "my-custom-cert",
			},
			preInstall: func(t *testing.T, env *support.Env) {
				t.Helper()
				certPEM, keyPEM := generateSelfSignedTLS(t)
				_, err := env.Client.CoreV1().Secrets(env.Namespace).Create(
					env.Ctx,
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{Name: "my-custom-cert"},
						Type:       corev1.SecretTypeTLS,
						Data:       map[string][]byte{"tls.crt": certPEM, "tls.key": keyPEM},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
			},
			zitadel: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						ObjectMeta: assert.ObjectMetaAssertion{
							Annotations: assert.Matching[map[string]string](gomega.And(
								gomega.Not(gomega.HaveKey("checksum/secret-login-service-key")),
								gomega.HaveKey("checksum/secret-admin-service-key"),
							)),
						},
					},
				},
			},
			login: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						ObjectMeta: assert.ObjectMetaAssertion{
							Annotations: assert.Matching[map[string]string](gomega.And(
								gomega.Not(gomega.HaveKey("checksum/secret-login-service-key")),
							)),
						},
					},
				},
			},
		},
		{
			name: "admin-key-external-secret-no-checksum-annotation",
			setValues: map[string]string{
				"login.enabled": "true",
				"zitadel.adminServiceKey.existingSecretName": "my-admin-cert",
			},
			preInstall: func(t *testing.T, env *support.Env) {
				t.Helper()
				certPEM, keyPEM := generateSelfSignedTLS(t)
				_, err := env.Client.CoreV1().Secrets(env.Namespace).Create(
					env.Ctx,
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{Name: "my-admin-cert"},
						Type:       corev1.SecretTypeTLS,
						Data:       map[string][]byte{"tls.crt": certPEM, "tls.key": keyPEM},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
			},
			zitadel: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						ObjectMeta: assert.ObjectMetaAssertion{
							Annotations: assert.Matching[map[string]string](gomega.And(
								gomega.Not(gomega.HaveKey("checksum/secret-admin-service-key")),
								gomega.HaveKey("checksum/secret-login-service-key"),
							)),
						},
					},
				},
			},
			login: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						ObjectMeta: assert.ObjectMetaAssertion{
							Annotations: assert.Matching[map[string]string](gomega.And(
								gomega.HaveKey("checksum/secret-login-service-key"),
							)),
						},
					},
				},
			},
		},
		{
			name: "component-overrides",
			setValues: map[string]string{
				"login.enabled":         "true",
				"login.ingress.enabled": "true",

				"zitadel.podSecurityContext.runAsNonRoot":          "true",
				"zitadel.podSecurityContext.runAsUser":             "2000",
				"zitadel.podSecurityContext.fsGroup":               "2000",
				"zitadel.podSecurityContext.seccompProfile.type":   "RuntimeDefault",
				"zitadel.securityContext.runAsNonRoot":             "true",
				"zitadel.securityContext.runAsUser":                "2000",
				"zitadel.securityContext.readOnlyRootFilesystem":   "true",
				"zitadel.securityContext.privileged":               "false",
				"zitadel.securityContext.allowPrivilegeEscalation": "false",
				"zitadel.securityContext.capabilities.drop[0]":     "ALL",
				"login.podSecurityContext.runAsNonRoot":            "true",
				"login.podSecurityContext.runAsUser":               "3000",
				"login.podSecurityContext.fsGroup":                 "3000",
				"login.podSecurityContext.seccompProfile.type":     "RuntimeDefault",
				"login.securityContext.runAsNonRoot":               "true",
				"login.securityContext.runAsUser":                  "3000",
				"login.securityContext.readOnlyRootFilesystem":     "true",
				"login.securityContext.privileged":                 "false",
				"login.securityContext.allowPrivilegeEscalation":   "false",
				"login.securityContext.capabilities.drop[0]":       "NET_RAW",
			},
			zitadel: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						Spec: assert.PodSpecAssertion{
							SecurityContext: assert.PodSecurityContextAssertion{
								RunAsNonRoot: assert.SomePtr(true),
								RunAsUser:    assert.SomePtr(int64(2000)),
								FSGroup:      assert.SomePtr(int64(2000)),
								SeccompProfile: assert.SeccompProfileAssertion{
									Type: assert.Some(corev1.SeccompProfileTypeRuntimeDefault),
								},
							},
							Containers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("zitadel"),
									SecurityContext: assert.SecurityContextAssertion{
										RunAsNonRoot:             assert.SomePtr(true),
										RunAsUser:                assert.SomePtr(int64(2000)),
										ReadOnlyRootFilesystem:   assert.SomePtr(true),
										Privileged:               assert.SomePtr(false),
										AllowPrivilegeEscalation: assert.SomePtr(false),
										Capabilities: assert.CapabilitiesAssertion{
											Drop: assert.Some([]corev1.Capability{"ALL"}),
										},
									},
								},
							}),
						},
					},
				},
			},
			login: &assert.DeploymentAssertion{
				Spec: assert.DeploymentSpecAssertion{
					Template: assert.PodTemplateSpecAssertion{
						Spec: assert.PodSpecAssertion{
							SecurityContext: assert.PodSecurityContextAssertion{
								RunAsNonRoot: assert.SomePtr(true),
								RunAsUser:    assert.SomePtr(int64(3000)),
								FSGroup:      assert.SomePtr(int64(3000)),
								SeccompProfile: assert.SeccompProfileAssertion{
									Type: assert.Some(corev1.SeccompProfileTypeRuntimeDefault),
								},
							},
							Containers: assert.Some([]assert.ContainerAssertion{
								{
									Name: assert.Some("zitadel-login"),
									SecurityContext: assert.SecurityContextAssertion{
										RunAsNonRoot:             assert.SomePtr(true),
										RunAsUser:                assert.SomePtr(int64(3000)),
										ReadOnlyRootFilesystem:   assert.SomePtr(true),
										Privileged:               assert.SomePtr(false),
										AllowPrivilegeEscalation: assert.SomePtr(false),
										Capabilities: assert.CapabilitiesAssertion{
											Drop: assert.Some([]corev1.Capability{"NET_RAW"}),
										},
									},
								},
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
			support.WithNamespace(t, func(env *support.Env) {
				if tc.preInstall != nil {
					tc.preInstall(t, env)
				}

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
