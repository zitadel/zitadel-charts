package support

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"

	testsupport "github.com/zitadel/zitadel-charts/test/support"
)

// WithPostgres installs a lightweight Bitnami PostgreSQL release named "db"
// into the test environment's namespace. It disables persistence and
// authentication for fast, ephemeral test environments suitable for automated
// testing scenarios.
func WithPostgres(testing *testing.T, env *testsupport.Env) {
	testing.Helper()

	chartRepository := "https://charts.bitnami.com/bitnami"

	_, _ = helm.RunHelmCommandAndGetOutputE(testing, &helm.Options{},
		"repo", "add", "bitnami", chartRepository)

	helmOptions := &helm.Options{
		KubectlOptions: kubectlOptions(env),
		SetValues: map[string]string{
			"image.repository":                   "bitnamilegacy/postgresql",
			"volumePermissions.image.repository": "bitnamilegacy/os-shell",
			// Pre-create the "zitadel" database. Configmap-mode tests don't
			// need this (ZITADEL's admin connection creates it), but DSN-mode
			// tests do: the DSN must point to an existing database because
			// ZITADEL's init command uses it directly without CREATE DATABASE.
			"auth.database": "zitadel",
		},
		ExtraArgs: map[string][]string{
			"upgrade": {
				"--install",
				"--hide-notes",
				// Pin the chart version: Bitnami prunes versions from its
				// (deprecated) public registry without notice, so installing the
				// latest can suddenly 404. 18.5.13 is the version vendored by the
				// zitadel chart and is known to still resolve.
				"--version", "18.5.13",
				"--set-string", "primary.persistence.enabled=false",
				"--set-string", "primary.pgHbaConfiguration=host all all all trust",
				"--set-string", "primary.extendedConfiguration=max_connections = 500",
			},
		},
	}

	chartName := fmt.Sprintf("%s/postgresql", "bitnami")
	require.NoError(testing, helm.UpgradeE(testing, helmOptions, chartName, "db"))
}
