package smoke_test_test

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/zitadel/zitadel-charts/test/assert"
	setup "github.com/zitadel/zitadel-charts/test/smoke/support"
	"github.com/zitadel/zitadel-charts/test/support"
)

func TestSecretsMatrix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		setValues        map[string]string
		masterkey        *assert.SecretAssertion
		machineKey       *assert.SecretAssertion
		machineKeyName   string
		machinePat       *assert.SecretAssertion
		machinePatName   string
		loginClient      *assert.SecretAssertion
		loginClientName  string
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
			masterkey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("masterkey", gomega.Not(gomega.BeEmpty())),
				),
			},
			machineKeyName: "iam-admin",
			machineKey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("iam-admin.json", gomega.Not(gomega.BeEmpty())),
				),
			},
			machinePatName: "iam-admin-pat",
			machinePat: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("pat", gomega.Not(gomega.BeEmpty())),
				),
			},
			loginClientName: "login-client",
			loginClient: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("pat", gomega.Not(gomega.BeEmpty())),
				),
			},
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
			masterkey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("masterkey", gomega.Not(gomega.BeEmpty())),
				),
			},
			machineKeyName: "my-machine",
			machineKey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("my-machine.json", gomega.Not(gomega.BeEmpty())),
				),
			},
			loginClientName: "login-client",
			loginClient: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("pat", gomega.Not(gomega.BeEmpty())),
				),
			},
		},
		{
			name: "login-client-only",
			setValues: map[string]string{
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Username":   "login-client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Machine.Name":       "Login Client",
				"zitadel.configmapConfig.FirstInstance.Org.LoginClient.Pat.ExpirationDate": "2029-01-01T00:00:00Z",
			},
			masterkey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("masterkey", gomega.Not(gomega.BeEmpty())),
				),
			},
			loginClientName: "login-client",
			loginClient: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("pat", gomega.Not(gomega.BeEmpty())),
				),
			},
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
			masterkey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("masterkey", gomega.Not(gomega.BeEmpty())),
				),
			},
			machineKeyName: "custom-admin",
			machineKey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("custom-admin.json", gomega.Not(gomega.BeEmpty())),
				),
			},
			machinePatName: "custom-admin-pat",
			machinePat: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("pat", gomega.Not(gomega.BeEmpty())),
				),
			},
			loginClientName: "myapp-login-client",
			loginClient: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("pat", gomega.Not(gomega.BeEmpty())),
				),
			},
		},
		{
			name:      "minimal-no-setup",
			setValues: map[string]string{},
			masterkey: &assert.SecretAssertion{
				Data: assert.Matching[map[string][]byte](
					gomega.HaveKeyWithValue("masterkey", gomega.Not(gomega.BeEmpty())),
				),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			support.WithNamespace(t, func(env *support.Env) {
				releaseName := setup.InstallZitadel(t, env, tc.name, tc.setValues)

				if tc.masterkey != nil {
					env.AssertPartial(t, releaseName+"-masterkey", *tc.masterkey)
				}

				if tc.machineKey != nil {
					env.AssertPartial(t, tc.machineKeyName, *tc.machineKey)
				} else if tc.machineKeyName != "" {
					env.AssertNone(t, tc.machineKeyName, assert.SecretAssertion{})
				}

				if tc.machinePat != nil {
					env.AssertPartial(t, tc.machinePatName, *tc.machinePat)
				} else if tc.machinePatName != "" {
					env.AssertNone(t, tc.machinePatName, assert.SecretAssertion{})
				}

				if tc.loginClient != nil {
					env.AssertPartial(t, tc.loginClientName, *tc.loginClient)
				} else if tc.loginClientName != "" {
					env.AssertNone(t, tc.loginClientName, assert.SecretAssertion{})
				}
			})
		})
	}
}
