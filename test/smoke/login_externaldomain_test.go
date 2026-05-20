package smoke_test_test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
)

// TestLoginExternalDomainRendering verifies the ExternalDomain runtime-env
// injection paths using helm template rendering only (no cluster install).
// These checks are deliberately render-only to keep them fast and to avoid
// adding to the K3s cluster install budget which is limited to 30 minutes.
func TestLoginExternalDomainRendering(t *testing.T) {
	t.Parallel()

	chartPath := setup.ChartPath(t)

	// minValues are the minimum values required for the chart to render
	// the login templates without errors.
	minValues := map[string]string{
		"zitadel.masterkey":                    "x123456789012345678901234567891y",
		"zitadel.configmapConfig.ExternalPort": "443",
		"zitadel.configmapConfig.TLS.Enabled":  "false",
		"login.enabled":                         "true",
	}

	// t.Run: configmap-placeholder
	// Asserts that the login .env ConfigMap contains the literal shell-variable
	// placeholder ${ZITADEL_EXTERNALDOMAIN} in its .env data, and that Helm did
	// NOT expand the configured ExternalDomain value at template time.
	t.Run("configmap-placeholder", func(t *testing.T) {
		t.Parallel()

		values := make(map[string]string)
		for k, v := range minValues {
			values[k] = v
		}
		values["zitadel.configmapConfig.ExternalDomain"] = "check.example.com"

		output, err := helm.RenderTemplateE(t,
			&helm.Options{SetValues: values},
			chartPath, "login-ext-domain",
			[]string{"templates/configmap_login.yaml"})
		require.NoError(t, err)

		var cm corev1.ConfigMap
		require.NoError(t, yaml.Unmarshal([]byte(output), &cm))

		envFile := cm.Data[".env"]
		require.Contains(t, envFile, "${ZITADEL_EXTERNALDOMAIN}",
			".env must contain the literal shell placeholder, not a Helm-expanded value")
		require.NotContains(t, envFile, "check.example.com",
			".env must not contain the Helm-expanded domain string")
	})

	// t.Run: deployment-auto-inject
	// When ExternalDomain is set via configmapConfig and login.env does not
	// include ZITADEL_EXTERNALDOMAIN, the chart must inject exactly one
	// ZITADEL_EXTERNALDOMAIN env var on the login container with that value.
	t.Run("deployment-auto-inject", func(t *testing.T) {
		t.Parallel()

		values := make(map[string]string)
		for k, v := range minValues {
			values[k] = v
		}
		values["zitadel.configmapConfig.ExternalDomain"] = "auth.example.com"

		output, err := helm.RenderTemplateE(t,
			&helm.Options{SetValues: values},
			chartPath, "login-ext-domain",
			[]string{"templates/deployment_login.yaml"})
		require.NoError(t, err)

		var dep appsv1.Deployment
		require.NoError(t, yaml.Unmarshal([]byte(output), &dep))

		count, value := countEnvVar(dep.Spec.Template.Spec.Containers, "zitadel-login", "ZITADEL_EXTERNALDOMAIN")
		require.Equal(t, 1, count, "expected exactly one ZITADEL_EXTERNALDOMAIN env var on the login container")
		require.Equal(t, "auth.example.com", value, "expected the configured ExternalDomain value")
	})

	// t.Run: deployment-no-duplicate-when-user-provides
	// When the user already supplies ZITADEL_EXTERNALDOMAIN in login.env, the
	// chart must not inject a second copy (Kubernetes rejects duplicate env var
	// names). The user-supplied value must be the one that wins.
	t.Run("deployment-no-duplicate-when-user-provides", func(t *testing.T) {
		t.Parallel()

		values := make(map[string]string)
		for k, v := range minValues {
			values[k] = v
		}
		values["zitadel.configmapConfig.ExternalDomain"] = "auth.example.com"
		values["login.env[0].name"] = "ZITADEL_EXTERNALDOMAIN"
		values["login.env[0].value"] = "custom.auth.example.com"

		output, err := helm.RenderTemplateE(t,
			&helm.Options{SetValues: values},
			chartPath, "login-ext-domain",
			[]string{"templates/deployment_login.yaml"})
		require.NoError(t, err)

		var dep appsv1.Deployment
		require.NoError(t, yaml.Unmarshal([]byte(output), &dep))

		count, value := countEnvVar(dep.Spec.Template.Spec.Containers, "zitadel-login", "ZITADEL_EXTERNALDOMAIN")
		require.Equal(t, 1, count, "chart must not inject a duplicate ZITADEL_EXTERNALDOMAIN when user already provides it")
		require.Equal(t, "custom.auth.example.com", value, "user-supplied value must win over the auto-injected one")
	})
}

// countEnvVar scans the named container's env list and returns how many times
// envName appears, plus the Value field of the last matching entry.
func countEnvVar(containers []corev1.Container, containerName, envName string) (int, string) {
	for _, c := range containers {
		if c.Name != containerName {
			continue
		}
		count := 0
		value := ""
		for _, e := range c.Env {
			if e.Name == envName {
				count++
				value = e.Value
			}
		}
		return count, value
	}
	return 0, ""
}
