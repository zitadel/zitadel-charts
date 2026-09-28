package smoke_test_test

import (
	"maps"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
)

const (
	sessionCookieSecretEnv = "ZITADEL_SESSION_COOKIE_SECRET"
	loginContainerName     = "zitadel-login"
)

// TestLoginSessionCookieSecretRendering verifies how the chart wires the
// Login UI session cookie secret. The chart never generates that secret: it
// only references a Secret the user names in login.sessionCookieSecretName,
// and it stays out of the way when the user sets the variable in login.env.
// These checks are render-only to keep them fast and to avoid adding to the
// K3s cluster install budget.
func TestLoginSessionCookieSecretRendering(t *testing.T) {
	t.Parallel()

	chartPath := setup.ChartPath(t)

	// minValues are the minimum values required for the chart to render
	// the login templates without errors.
	minValues := map[string]string{
		"zitadel.masterkey":                      "x123456789012345678901234567891y",
		"zitadel.configmapConfig.ExternalDomain": "auth.example.com",
		"zitadel.configmapConfig.ExternalPort":   "443",
		"zitadel.configmapConfig.TLS.Enabled":    "false",
		"login.enabled":                          "true",
	}

	// userEnv is a login.env entry that sets the variable itself, from a
	// Secret and key of the user's own choosing.
	userEnv := map[string]string{
		"login.env[0].name":                        sessionCookieSecretEnv,
		"login.env[0].valueFrom.secretKeyRef.name": "user-managed",
		"login.env[0].valueFrom.secretKeyRef.key":  "custom-key",
	}

	testCases := []struct {
		name string
		// values are applied on top of minValues.
		values []map[string]string
		// wantSecret is the Secret the single env entry must reference.
		// Empty means the variable must not be set at all.
		wantSecret string
		wantKey    string
	}{
		{
			// Leaving the value empty must not change the deployment: the
			// login keeps deriving the key from the login service key.
			name: "not-set-by-default",
		},
		{
			name:       "from-existing-secret",
			values:     []map[string]string{{"login.sessionCookieSecretName": "my-cookie"}},
			wantSecret: "my-cookie",
			wantKey:    sessionCookieSecretEnv,
		},
		{
			// Kubernetes rejects duplicate env names, so the chart must not
			// add its own entry next to the user's. The user's entry wins.
			name:       "no-duplicate-when-user-provides",
			values:     []map[string]string{{"login.sessionCookieSecretName": "my-cookie"}, userEnv},
			wantSecret: "user-managed",
			wantKey:    "custom-key",
		},
		{
			name:       "user-env-without-secret-name",
			values:     []map[string]string{userEnv},
			wantSecret: "user-managed",
			wantKey:    "custom-key",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			values := maps.Clone(minValues)
			for _, extra := range tc.values {
				maps.Copy(values, extra)
			}

			output, err := helm.RenderTemplateE(t,
				&helm.Options{SetValues: values},
				chartPath, "login-session-cookie",
				[]string{"templates/deployment_login.yaml"})
			require.NoError(t, err)

			var dep appsv1.Deployment
			require.NoError(t, yaml.Unmarshal([]byte(output), &dep))

			envVars := findEnvVars(dep.Spec.Template.Spec.Containers, loginContainerName, sessionCookieSecretEnv)
			if tc.wantSecret == "" {
				require.Empty(t, envVars, "the chart must not set %s unless a Secret is named", sessionCookieSecretEnv)
				return
			}

			require.Len(t, envVars, 1, "expected exactly one %s env var on the login container", sessionCookieSecretEnv)
			require.Empty(t, envVars[0].Value, "the secret must be referenced, never rendered as a plain value")
			require.NotNil(t, envVars[0].ValueFrom)
			require.NotNil(t, envVars[0].ValueFrom.SecretKeyRef)
			require.Equal(t, tc.wantSecret, envVars[0].ValueFrom.SecretKeyRef.Name)
			require.Equal(t, tc.wantKey, envVars[0].ValueFrom.SecretKeyRef.Key)
			require.Nil(t, envVars[0].ValueFrom.SecretKeyRef.Optional,
				"a missing Secret must fail the pod instead of silently falling back")
		})
	}

	// t.Run: login-disabled
	// With the login disabled nothing in the chart may reference the variable,
	// even if a Secret name is configured.
	t.Run("login-disabled", func(t *testing.T) {
		t.Parallel()

		values := maps.Clone(minValues)
		values["login.enabled"] = "false"
		values["login.sessionCookieSecretName"] = "my-cookie"

		output, err := helm.RenderTemplateE(t,
			&helm.Options{SetValues: values},
			chartPath, "login-session-cookie", nil)
		require.NoError(t, err)
		require.NotContains(t, output, sessionCookieSecretEnv)
	})

	// t.Run: zitadel-deployment-untouched
	// The secret belongs to the Login UI only. The ZITADEL API deployment must
	// neither receive the variable nor restart because of it.
	t.Run("zitadel-deployment-untouched", func(t *testing.T) {
		t.Parallel()

		values := maps.Clone(minValues)
		values["login.sessionCookieSecretName"] = "my-cookie"

		output, err := helm.RenderTemplateE(t,
			&helm.Options{SetValues: values},
			chartPath, "login-session-cookie",
			[]string{"templates/deployment_zitadel.yaml"})
		require.NoError(t, err)
		require.NotContains(t, output, sessionCookieSecretEnv)
		require.NotContains(t, output, "my-cookie")
	})
}

// findEnvVars returns every env entry named envName on the named container.
// Unlike countEnvVar it returns the full entries, so valueFrom references can
// be asserted as well.
func findEnvVars(containers []corev1.Container, containerName, envName string) []corev1.EnvVar {
	var found []corev1.EnvVar
	for _, c := range containers {
		if c.Name != containerName {
			continue
		}
		for _, e := range c.Env {
			if e.Name == envName {
				found = append(found, e)
			}
		}
	}
	return found
}
