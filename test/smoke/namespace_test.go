package smoke_test_test

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"

	"github.com/zitadel/zitadel-charts/test/smoke/support"
)

func TestNamespaceFieldExplicitlySet(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                   support.DigestTag,
			"login.enabled":               "true",
			"login.pdb.enabled":           "true",
			"login.ingress.enabled":       "true",
			"ingress.enabled":             "true",
			"ingress.controller":          "aws",
			"zitadel.masterkey":           "01234567890123456789012345678901",
			"pdb.enabled":                 "true",
			"zitadel.autoscaling.enabled": "true",
			"login.autoscaling.enabled":   "true",
		},
	}

	releaseName := "namespace-test"

	templates := []string{
		"templates/configmap_login.yaml",
		"templates/configmap_zitadel.yaml",
		"templates/deployment_login.yaml",
		"templates/deployment_zitadel.yaml",
		"templates/hpa_login.yaml",
		"templates/hpa_zitadel.yaml",
		"templates/ingress_login.yaml",
		"templates/ingress_zitadel.yaml",
		"templates/job_cleanup.yaml",
		"templates/job_init.yaml",
		"templates/job_setup.yaml",
		"templates/pdb_login.yaml",
		"templates/pdb_zitadel.yaml",
		"templates/service_login.yaml",
		"templates/service_zitadel-grpc.yaml",
		"templates/service_zitadel.yaml",
		"templates/serviceaccount_login.yaml",
		"templates/serviceaccount_zitadel.yaml",
	}

	for _, tmpl := range templates {
		tmpl := tmpl
		t.Run(tmpl, func(t *testing.T) {
			rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{tmpl})

			if rendered == "" {
				t.Skip("template did not render (conditional)")
				return
			}

			require.True(t, strings.Contains(rendered, "namespace:"),
				"template %s should contain namespace field", tmpl)
			require.True(t, strings.Contains(rendered, "namespace: default"),
				"template %s should have namespace set to release namespace", tmpl)
		})
	}
}

func TestNamespaceFieldInRBACResources(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                              support.DigestTag,
			"login.enabled":                          "true",
			"zitadel.masterkey":                      "01234567890123456789012345678901",
			"zitadel.configmapConfig.ExternalDomain": "example.com",
		},
	}

	releaseName := "rbac-namespace-test"

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{
		"templates/rbac_zitadel.yaml",
	})

	require.Contains(t, rendered, "kind: Role", "RBAC template should contain Role")
	require.Contains(t, rendered, "kind: RoleBinding", "RBAC template should contain RoleBinding")
	require.True(t, strings.Count(rendered, "namespace:") >= 2,
		"RBAC template should have namespace field for both Role and RoleBinding")
}

func TestNamespaceFieldInSecrets(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":                      support.DigestTag,
			"login.enabled":                  "true",
			"zitadel.masterkey":              "01234567890123456789012345678901",
			"zitadel.selfSignedCert.enabled": "true",
			"zitadel.dbSslCaCrt":             "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t",
		},
	}

	releaseName := "secret-namespace-test"

	secrets := []string{
		"templates/secret_db-ssl-ca-crt.yaml",
		"templates/secret_self-signed.yaml",
		"templates/secret_zitadel-masterkey.yaml",
	}

	for _, tmpl := range secrets {
		tmpl := tmpl
		t.Run(tmpl, func(t *testing.T) {
			rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{tmpl})

			if rendered == "" {
				t.Skip("template did not render (conditional)")
				return
			}

			require.True(t, strings.Contains(rendered, "namespace:"),
				"template %s should contain namespace field", tmpl)
		})
	}
}

func TestNamespaceFieldInSecretConfig(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":         support.DigestTag,
			"login.enabled":     "true",
			"zitadel.masterkey": "01234567890123456789012345678901",
			"zitadel.secretConfig.Database.Postgres.User.Password": "test-password",
		},
	}

	releaseName := "secret-config-namespace-test"

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{
		"templates/secret_zitadel-secrets.yaml",
	})

	require.NotEmpty(t, rendered, "secret_zitadel-secrets should render when secretConfig is set")
	require.Contains(t, rendered, "namespace:", "secret_zitadel-secrets should have namespace field")
}

func TestNamespaceFieldInServiceMonitor(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	t.Run("default-namespace", func(t *testing.T) {
		options := &helm.Options{
			SetValues: map[string]string{
				"image.tag":                      support.DigestTag,
				"zitadel.masterkey":              "01234567890123456789012345678901",
				"metrics.enabled":                "true",
				"metrics.serviceMonitor.enabled": "true",
			},
		}

		releaseName := "servicemonitor-namespace-test"
		rendered := helm.RenderTemplate(t, options, chartPath, releaseName,
			[]string{"templates/servicemonitor.yaml"})

		require.Contains(t, rendered, "namespace:", "ServiceMonitor should have namespace field")
		require.Contains(t, rendered, "namespace: default",
			"ServiceMonitor should use release namespace")
	})

	t.Run("custom-namespace", func(t *testing.T) {
		options := &helm.Options{
			SetValues: map[string]string{
				"image.tag":                        support.DigestTag,
				"zitadel.masterkey":                "01234567890123456789012345678901",
				"metrics.enabled":                  "true",
				"metrics.serviceMonitor.enabled":   "true",
				"metrics.serviceMonitor.namespace": "custom-namespace",
			},
		}

		releaseName := "servicemonitor-custom-ns-test"
		rendered := helm.RenderTemplate(t, options, chartPath, releaseName,
			[]string{"templates/servicemonitor.yaml"})

		require.Contains(t, rendered, "namespace: custom-namespace",
			"ServiceMonitor should use custom namespace when specified")
	})
}

func TestNamespaceFieldInDebugReplicaSet(t *testing.T) {
	t.Parallel()

	chartPath := support.ChartPath(t)

	options := &helm.Options{
		SetValues: map[string]string{
			"image.tag":             support.DigestTag,
			"zitadel.masterkey":     "01234567890123456789012345678901",
			"zitadel.debug.enabled": "true",
		},
	}

	releaseName := "debug-namespace-test"

	rendered := helm.RenderTemplate(t, options, chartPath, releaseName, []string{
		"templates/debug_replicaset.yaml",
	})

	require.NotEmpty(t, rendered, "debug_replicaset should render when debug.enabled is true")
	require.Contains(t, rendered, "namespace:", "debug_replicaset should have namespace field")
	require.Contains(t, rendered, "namespace: default", "debug_replicaset should use release namespace")
}
