package smoke_test_test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/zitadel/zitadel-charts/test/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

//goland:noinspection ALL
func TestGatewayHTTPRouteMatrix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setValues map[string]string
		zitadel   *assert.HTTPRouteAssertion
		login     *assert.HTTPRouteAssertion
	}{
		{
			name: "labels",
			setValues: map[string]string{
				"gateway.httpRoute.enabled":                   "true",
				"gateway.httpRoute.parentRefs[0].name":        "my-gateway",
				"gateway.httpRoute.hostnames[0]":              "zitadel.example.local",
				"login.enabled":                               "true",
				"login.gateway.httpRoute.enabled":              "true",
				"login.gateway.httpRoute.parentRefs[0].name":  "my-gateway",
				"login.gateway.httpRoute.hostnames[0]":        "zitadel.example.local",
			},
			zitadel: &assert.HTTPRouteAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel"),
						gomega.HaveKeyWithValue("app.kubernetes.io/managed-by", "Helm"),
						gomega.HaveKeyWithValue("app.kubernetes.io/version", gomega.MatchRegexp(`^v?\d+\.\d+\.\d+`)),
					)),
				},
			},
			login: &assert.HTTPRouteAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel-login"),
						gomega.HaveKeyWithValue("app.kubernetes.io/managed-by", "Helm"),
						gomega.HaveKeyWithValue("app.kubernetes.io/version", gomega.MatchRegexp(`^v?\d+\.\d+\.\d+`)),
						gomega.HaveKeyWithValue("app.kubernetes.io/component", "login"),
					)),
				},
			},
		},
		{
			name: "custom-labels",
			setValues: map[string]string{
				"gateway.httpRoute.enabled":             "true",
				"gateway.httpRoute.parentRefs[0].name":  "my-gw",
				"gateway.httpRoute.labels.custom-label": "custom-value",
			},
			zitadel: &assert.HTTPRouteAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("custom-label", "custom-value"),
						gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel"),
					)),
				},
			},
		},
		{
			name: "custom-hosts",
			setValues: map[string]string{
				"gateway.httpRoute.enabled":            "true",
				"gateway.httpRoute.parentRefs[0].name": "my-gw",
				"gateway.httpRoute.hostnames[0]":       "custom.example.com",
				"gateway.httpRoute.hostnames[1]":       "other.example.com",
			},
			zitadel: &assert.HTTPRouteAssertion{
				Spec: assert.HTTPRouteSpecAssertion{
					Hostnames: assert.Some([]gatewayv1.Hostname{
						"custom.example.com",
						"other.example.com",
					}),
				},
			},
		},
		{
			name: "custom-annotations",
			setValues: map[string]string{
				"gateway.httpRoute.enabled":                 "true",
				"gateway.httpRoute.parentRefs[0].name":      "my-gw",
				"gateway.httpRoute.annotations.example-foo": "bar",
				"gateway.httpRoute.annotations.example-baz": "qux",
			},
			zitadel: &assert.HTTPRouteAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("example-foo", "bar"),
						gomega.HaveKeyWithValue("example-baz", "qux"),
					)),
				},
			},
		},
		{
			name: "no-custom-annotations",
			setValues: map[string]string{
				"gateway.httpRoute.enabled":            "true",
				"gateway.httpRoute.parentRefs[0].name": "my-gw",
			},
			zitadel: &assert.HTTPRouteAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Annotations: assert.Matching[map[string]string](gomega.And(
						gomega.Not(gomega.HaveKey("example-foo")),
						gomega.Not(gomega.HaveKey("example-baz")),
					)),
				},
			},
		},
		{
			name: "default-path",
			setValues: map[string]string{
				"gateway.httpRoute.enabled":                   "true",
				"gateway.httpRoute.parentRefs[0].name":        "my-gw",
				"login.enabled":                               "true",
				"login.gateway.httpRoute.enabled":              "true",
				"login.gateway.httpRoute.parentRefs[0].name":  "my-gw",
			},
			zitadel: &assert.HTTPRouteAssertion{
				Spec: assert.HTTPRouteSpecAssertion{
					Rules: assert.Some([]assert.HTTPRouteRuleAssertion{
						{
							Matches: assert.Some([]assert.HTTPRouteMatchAssertion{
								{
									Path: assert.HTTPPathMatchAssertion{
										Type:  assert.SomePtr(gatewayv1.PathMatchPathPrefix),
										Value: assert.SomePtr("/"),
									},
								},
							}),
						},
					}),
				},
			},
			login: &assert.HTTPRouteAssertion{
				Spec: assert.HTTPRouteSpecAssertion{
					Rules: assert.Some([]assert.HTTPRouteRuleAssertion{
						{
							Matches: assert.Some([]assert.HTTPRouteMatchAssertion{
								{
									Path: assert.HTTPPathMatchAssertion{
										Type:  assert.SomePtr(gatewayv1.PathMatchPathPrefix),
										Value: assert.SomePtr("/ui/v2/login"),
									},
								},
							}),
						},
					}),
				},
			},
		},
		{
			name: "parent-refs",
			setValues: map[string]string{
				"gateway.httpRoute.enabled":                   "true",
				"gateway.httpRoute.parentRefs[0].name":        "my-gateway",
				"gateway.httpRoute.parentRefs[0].namespace":   "gateway-ns",
				"gateway.httpRoute.parentRefs[0].sectionName": "https",
				"gateway.httpRoute.parentRefs[0].port":        "443",
			},
			zitadel: &assert.HTTPRouteAssertion{
				Spec: assert.HTTPRouteSpecAssertion{
					CommonRouteSpec: assert.CommonRouteSpecAssertion{
						ParentRefs: assert.Some([]assert.ApisParentReferenceAssertion{
							{
								Name:        assert.Some(gatewayv1.ObjectName("my-gateway")),
								Namespace:   assert.SomePtr(gatewayv1.Namespace("gateway-ns")),
								SectionName: assert.SomePtr(gatewayv1.SectionName("https")),
								Port:        assert.SomePtr(gatewayv1.PortNumber(443)),
							},
						}),
					},
				},
			},
		},
		{
			name: "filters",
			setValues: map[string]string{
				"gateway.httpRoute.enabled":                                       "true",
				"gateway.httpRoute.parentRefs[0].name":                            "my-gw",
				"gateway.httpRoute.filters[0].type":                               "RequestHeaderModifier",
				"gateway.httpRoute.filters[0].requestHeaderModifier.set[0].name":  "X-Custom",
				"gateway.httpRoute.filters[0].requestHeaderModifier.set[0].value": "test",
			},
			zitadel: &assert.HTTPRouteAssertion{
				Spec: assert.HTTPRouteSpecAssertion{
					Rules: assert.Some([]assert.HTTPRouteRuleAssertion{
						{
							Filters: assert.Some([]assert.HTTPRouteFilterAssertion{
								{
									Type: assert.Some(gatewayv1.HTTPRouteFilterRequestHeaderModifier),
									RequestHeaderModifier: assert.HTTPHeaderFilterAssertion{
										Set: assert.Some([]assert.ApisHTTPHeaderAssertion{
											{
												Name:  assert.Some(gatewayv1.HTTPHeaderName("X-Custom")),
												Value: assert.Some("test"),
											},
										}),
									},
								},
							}),
						},
					}),
				},
			},
		},
		{
			name: "timeouts",
			setValues: map[string]string{
				"gateway.httpRoute.enabled":                 "true",
				"gateway.httpRoute.parentRefs[0].name":      "my-gw",
				"gateway.httpRoute.timeouts.request":        "30s",
				"gateway.httpRoute.timeouts.backendRequest": "20s",
			},
			zitadel: &assert.HTTPRouteAssertion{
				Spec: assert.HTTPRouteSpecAssertion{
					Rules: assert.Some([]assert.HTTPRouteRuleAssertion{
						{
							Timeouts: assert.HTTPRouteTimeoutsAssertion{
								Request:        assert.SomePtr(gatewayv1.Duration("30s")),
								BackendRequest: assert.SomePtr(gatewayv1.Duration("20s")),
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

//goland:noinspection ALL
func TestGatewayGRPCRouteMatrix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setValues map[string]string
		zitadel   *assert.GRPCRouteAssertion
	}{
		{
			name: "labels",
			setValues: map[string]string{
				"gateway.grpcRoute.enabled":            "true",
				"gateway.grpcRoute.parentRefs[0].name": "my-gateway",
				"gateway.grpcRoute.hostnames[0]":       "zitadel.example.local",
			},
			zitadel: &assert.GRPCRouteAssertion{
				ObjectMeta: assert.ObjectMetaAssertion{
					Labels: assert.Matching[map[string]string](gomega.And(
						gomega.HaveKeyWithValue("app.kubernetes.io/name", "zitadel"),
						gomega.HaveKeyWithValue("app.kubernetes.io/managed-by", "Helm"),
						gomega.HaveKeyWithValue("app.kubernetes.io/version", gomega.MatchRegexp(`^v?\d+\.\d+\.\d+`)),
					)),
				},
			},
		},
		{
			name: "filters",
			setValues: map[string]string{
				"gateway.grpcRoute.enabled":                                       "true",
				"gateway.grpcRoute.parentRefs[0].name":                            "my-gw",
				"gateway.grpcRoute.filters[0].type":                               "RequestHeaderModifier",
				"gateway.grpcRoute.filters[0].requestHeaderModifier.set[0].name":  "X-Custom",
				"gateway.grpcRoute.filters[0].requestHeaderModifier.set[0].value": "grpc-test",
			},
			zitadel: &assert.GRPCRouteAssertion{
				Spec: assert.GRPCRouteSpecAssertion{
					Rules: assert.Some([]assert.GRPCRouteRuleAssertion{
						{
							Filters: assert.Some([]assert.GRPCRouteFilterAssertion{
								{
									Type: assert.Some(gatewayv1.GRPCRouteFilterRequestHeaderModifier),
									RequestHeaderModifier: assert.HTTPHeaderFilterAssertion{
										Set: assert.Some([]assert.ApisHTTPHeaderAssertion{
											{
												Name:  assert.Some(gatewayv1.HTTPHeaderName("X-Custom")),
												Value: assert.Some("grpc-test"),
											},
										}),
									},
								},
							}),
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
					env.AssertPartial(t, releaseName+"-grpc", *tc.zitadel)
				}
			})
		})
	}
}

//goland:noinspection ALL
func TestGatewayRoutesDisabledByDefault(t *testing.T) {
	t.Parallel()

	support.WithNamespace(t, func(env *support.Env) {
		releaseName := setup.InstallZitadel(t, env, "gw-disabled", nil)

		env.AssertNone(t, releaseName, assert.HTTPRouteAssertion{})
		env.AssertNone(t, releaseName+"-grpc", assert.GRPCRouteAssertion{})
		env.AssertNone(t, releaseName+"-login", assert.HTTPRouteAssertion{})
	})
}

func TestGatewayHTTPRouteEmptyPathsFails(t *testing.T) {
	t.Parallel()

	chartPath := setup.ChartPath(t)

	t.Run("zitadel-empty-paths", func(t *testing.T) {
		options := &helm.Options{
			SetValues: map[string]string{
				"zitadel.configmapConfig.ExternalDomain": "zitadel.example.local",
				"zitadel.masterkey":                      "01234567890123456789012345678901",
				"gateway.httpRoute.enabled":               "true",
				"gateway.httpRoute.parentRefs[0].name":    "my-gw",
			},
			SetJsonValues: map[string]string{
				"gateway.httpRoute.paths": "[]",
			},
		}

		_, err := helm.RenderTemplateE(t, options, chartPath, "gateway-empty-paths",
			[]string{"templates/httproute_zitadel.yaml"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "gateway.httpRoute.paths must not be empty")
	})

	t.Run("login-empty-paths", func(t *testing.T) {
		options := &helm.Options{
			SetValues: map[string]string{
				"zitadel.configmapConfig.ExternalDomain":    "zitadel.example.local",
				"zitadel.masterkey":                          "01234567890123456789012345678901",
				"login.enabled":                              "true",
				"login.gateway.httpRoute.enabled":             "true",
				"login.gateway.httpRoute.parentRefs[0].name": "my-gw",
			},
			SetJsonValues: map[string]string{
				"login.gateway.httpRoute.paths": "[]",
			},
		}

		_, err := helm.RenderTemplateE(t, options, chartPath, "gateway-empty-paths",
			[]string{"templates/httproute_login.yaml"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "login.gateway.httpRoute.paths must not be empty")
	})
}
