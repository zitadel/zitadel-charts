package smoke_test_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/zitadel/zitadel-charts/test/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

func TestSecretsMatrix(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)
	chartPath := setup.ChartPath(t)

	testCases := []struct {
		name               string
		setValues          map[string]string
		machineKeySecret   bool
		machinePatSecret   bool
		loginClientSecret  bool
		machineKeyName     string
		machinePatName     string
		loginClientName    string
		machineKeyContent  string
		machinePatContent  string
		loginClientContent string
	}{
		{
			name: "default-all-enabled",
			setValues: map[string]string{
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Username":          "iam-admin",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Name":              "Admin Machine",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.ExpirationDate": "2029-01-01T00:00:00Z",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.Type":           "1",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Pat.ExpirationDate":        "2029-01-01T00:00:00Z",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Username":      "login-client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Name":          "Login Client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Pat.ExpirationDate":    "2029-01-01T00:00:00Z",
			},
			machineKeySecret:   true,
			machinePatSecret:   true,
			loginClientSecret:  true,
			machineKeyName:     "iam-admin",
			machinePatName:     "iam-admin-pat",
			loginClientName:    "login-client",
			machineKeyContent:  "iam-admin.json",
			machinePatContent:  "pat",
			loginClientContent: "pat",
		},
		{
			name: "machine-only-no-pat",
			setValues: map[string]string{
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Username":          "my-machine",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Name":              "My Machine",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.ExpirationDate": "2029-01-01T00:00:00Z",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.Type":           "1",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Username":      "login-client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Name":          "Login Client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Pat.ExpirationDate":    "2029-01-01T00:00:00Z",
			},
			machineKeySecret:   true,
			machinePatSecret:   false,
			loginClientSecret:  true,
			machineKeyName:     "my-machine",
			machinePatName:     "",
			loginClientName:    "login-client",
			machineKeyContent:  "my-machine.json",
			machinePatContent:  "",
			loginClientContent: "pat",
		},
		{
			name: "login-client-only",
			setValues: map[string]string{
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Username":   "login-client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Name":       "Login Client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Pat.ExpirationDate": "2029-01-01T00:00:00Z",
			},
			machineKeySecret:   false,
			machinePatSecret:   false,
			loginClientSecret:  true,
			machineKeyName:     "",
			machinePatName:     "",
			loginClientName:    "login-client",
			machineKeyContent:  "",
			machinePatContent:  "",
			loginClientContent: "pat",
		},
		{
			name: "custom-names-with-prefix",
			setValues: map[string]string{
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Username":          "custom-admin",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Name":              "Custom Admin",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.ExpirationDate": "2029-01-01T00:00:00Z",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.Type":           "1",
				"zitadel.configmapConfig.FirstInstance.Org.Machine.Pat.ExpirationDate":        "2029-01-01T00:00:00Z",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Username":      "login-client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Name":          "Login Client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Pat.ExpirationDate":    "2029-01-01T00:00:00Z",
				"login.loginClientSecretPrefix":                                               "myapp-",
			},
			machineKeySecret:   true,
			machinePatSecret:   true,
			loginClientSecret:  true,
			machineKeyName:     "custom-admin",
			machinePatName:     "custom-admin-pat",
			loginClientName:    "myapp-login-client",
			machineKeyContent:  "custom-admin.json",
			machinePatContent:  "pat",
			loginClientContent: "pat",
		},
		{
			name:               "minimal-no-setup",
			setValues:          map[string]string{},
			machineKeySecret:   false,
			machinePatSecret:   false,
			loginClientSecret:  false,
			machineKeyName:     "",
			machinePatName:     "",
			loginClientName:    "",
			machineKeyContent:  "",
			machinePatContent:  "",
			loginClientContent: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			support.WithNamespace(t, cluster, func(env *support.Env) {
				releaseName := setup.InstallZitadel(t, env, chartPath, tc.name, tc.setValues)

				// Verify masterkey secret
				masterkeyName := releaseName + "-masterkey"
				masterkeySecret := env.GetSecret(t, masterkeyName)
				assert.AssertPartial(t, masterkeySecret, assert.SecretAssertion{
					ObjectMeta: assert.ObjectMetaAssertion{
						Labels: assert.Some(map[string]string{
							"app.kubernetes.io/name":       "zitadel",
							"app.kubernetes.io/instance":   releaseName,
							"app.kubernetes.io/version":    "v4.10.1",
							"app.kubernetes.io/managed-by": "Helm",
						}),
					},
				}, masterkeyName)
				require.Contains(t, masterkeySecret.Data, "masterkey", "secret %q should contain key %q", masterkeyName, "masterkey")
				require.NotEmpty(t, masterkeySecret.Data["masterkey"], "secret %q key %q should not be empty", masterkeyName, "masterkey")

				if tc.machineKeySecret {
					secret := env.GetSecret(t, tc.machineKeyName)
					require.Contains(t, secret.Data, tc.machineKeyContent)
					require.NotEmpty(t, secret.Data[tc.machineKeyContent])
				} else if tc.machineKeyName != "" {
					_, err := env.GetSecretE(t, tc.machineKeyName)
					require.True(t, apierrors.IsNotFound(err), "secret %q should not exist", tc.machineKeyName)
				}

				if tc.machinePatSecret {
					secret := env.GetSecret(t, tc.machinePatName)
					require.Contains(t, secret.Data, tc.machinePatContent)
					require.NotEmpty(t, secret.Data[tc.machinePatContent])
				} else if tc.machinePatName != "" {
					_, err := env.GetSecretE(t, tc.machinePatName)
					require.True(t, apierrors.IsNotFound(err), "secret %q should not exist", tc.machinePatName)
				}

				if tc.loginClientSecret {
					secret := env.GetSecret(t, tc.loginClientName)
					require.Contains(t, secret.Data, tc.loginClientContent)
					require.NotEmpty(t, secret.Data[tc.loginClientContent])
				} else if tc.loginClientName != "" {
					_, err := env.GetSecretE(t, tc.loginClientName)
					require.True(t, apierrors.IsNotFound(err), "secret %q should not exist", tc.loginClientName)
				}
			})
		})
	}
}
