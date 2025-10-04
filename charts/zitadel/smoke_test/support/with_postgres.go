package support

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
)

// WithPostgres installs a lightweight Bitnami PostgreSQL release named "db"
// into the test environment's namespace. It disables persistence and
// authentication for fast, ephemeral test environments suitable for automated
// testing scenarios.
func WithPostgres(testing *testing.T, env *Env) {
	testing.Helper()

	repoAlias := "crdb-" + env.Namespace
	chartRepository := "https://charts.bitnami.com/bitnami"

	_, _ = helm.RunHelmCommandAndGetOutputE(testing, &helm.Options{},
		"repo", "add", repoAlias, chartRepository)

	helmOptions := &helm.Options{
		KubectlOptions: env.Kube,
		SetValues: map[string]string{
			"image.repository":                   "bitnamilegacy/postgresql",
			"volumePermissions.image.repository": "bitnamilegacy/os-shell",
		},
		ExtraArgs: map[string][]string{
			"upgrade": {
				"--install",
				"--hide-notes",
				"--no-update",
				"--set-string", "primary.persistence.enabled=false",
				"--set-string", "primary.pgHbaConfiguration=host all all all trust",
				"--set-string", "primary.extendedConfiguration=max_connections = 500",
			},
		},
	}

	chartName := fmt.Sprintf("%s/postgresql", repoAlias)
	releaseName := "db"

	require.NoError(testing, helm.UpgradeE(testing, helmOptions, chartName, releaseName))
}
