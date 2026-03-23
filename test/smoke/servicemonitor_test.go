package smoke_test_test

import (
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

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
				Spec: assert.ServiceMonitorSpecAssertion{
					Endpoints: assert.Some([]assert.EndpointAssertion{
						{
							Port:            assert.Some("http2-server"),
							Path:            assert.Some("/debug/metrics"),
							HonorLabels:     assert.Some(false),
							HonorTimestamps: assert.SomePtr(true),
						},
					}),
					Selector: assert.LabelSelectorAssertion{
						MatchLabels: assert.Matching[map[string]string](gomega.And(
							gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel"),
							gomega.HaveKey("app.kubernetes.io/instance"),
						)),
					},
					NamespaceSelector: assert.NamespaceSelectorAssertion{
						MatchNames: assert.Matching[[]string](gomega.HaveLen(1)),
					},
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
				Spec: assert.ServiceMonitorSpecAssertion{
					Endpoints: assert.Some([]assert.EndpointAssertion{
						{
							Port:            assert.Some("metrics"),
							Path:            assert.Some("/metrics"),
							HonorLabels:     assert.Some(false),
							HonorTimestamps: assert.SomePtr(true),
						},
					}),
					Selector: assert.LabelSelectorAssertion{
						MatchLabels: assert.Matching[map[string]string](gomega.And(
							gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel-login"),
							gomega.HaveKeyWithValue("app.kubernetes.io/component", "login"),
							gomega.HaveKey("app.kubernetes.io/instance"),
						)),
					},
					NamespaceSelector: assert.NamespaceSelectorAssertion{
						MatchNames: assert.Matching[[]string](gomega.HaveLen(1)),
					},
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
				Spec: assert.ServiceMonitorSpecAssertion{
					Endpoints: assert.Some([]assert.EndpointAssertion{
						{
							Port: assert.Some("http2-server"),
							Path: assert.Some("/debug/metrics"),
						},
					}),
				},
			},
			login: &assert.ServiceMonitorAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel-login"),
						gomega.HaveKeyWithValue("app.kubernetes.io/managed-by", "Helm"),
					)),
				},
				Spec: assert.ServiceMonitorSpecAssertion{
					Endpoints: assert.Some([]assert.EndpointAssertion{
						{
							Port: assert.Some("metrics"),
							Path: assert.Some("/metrics"),
						},
					}),
				},
			},
		},
		{
			name: "additional-labels",
			setValues: map[string]string{
				"metrics.enabled":                              "true",
				"metrics.serviceMonitor.enabled":               "true",
				"metrics.serviceMonitor.additionalLabels.team": "platform",

				"login.enabled":                                      "true",
				"login.metrics.enabled":                               "true",
				"login.metrics.serviceMonitor.enabled":                "true",
				"login.metrics.serviceMonitor.additionalLabels.team":  "frontend",
			},
			zitadel: &assert.ServiceMonitorAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("team", "platform"),
						gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel"),
					)),
				},
			},
			login: &assert.ServiceMonitorAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("team", "frontend"),
						gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel-login"),
					)),
				},
			},
		},
		{
			name: "scrape-interval-and-timeout",
			setValues: map[string]string{
				"login.enabled":                                  "true",
				"login.metrics.enabled":                          "true",
				"login.metrics.serviceMonitor.enabled":           "true",
				"login.metrics.serviceMonitor.scrapeInterval":    "15s",
				"login.metrics.serviceMonitor.scrapeTimeout":     "10s",
				"login.metrics.serviceMonitor.honorLabels":       "true",
				"login.metrics.serviceMonitor.honorTimestamps":   "false",
			},
			login: &assert.ServiceMonitorAssertion{
				Spec: assert.ServiceMonitorSpecAssertion{
					Endpoints: assert.Some([]assert.EndpointAssertion{
						{
							Port:            assert.Some("metrics"),
							Path:            assert.Some("/metrics"),
							Interval:        assert.Some(monitoringv1.Duration("15s")),
							ScrapeTimeout:   assert.Some(monitoringv1.Duration("10s")),
							HonorLabels:     assert.Some(true),
							HonorTimestamps: assert.SomePtr(false),
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
