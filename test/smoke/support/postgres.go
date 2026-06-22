package support

import (
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"

	testsupport "github.com/zitadel/zitadel-charts/test/support"
)

// WithPostgres installs a lightweight PostgreSQL release named "db" into the
// test environment's namespace. It disables persistence and authentication for
// fast, ephemeral test environments.
//
// The PostgreSQL chart is installed from the copy vendored in the zitadel chart
// (charts/zitadel/charts/postgresql-*.tgz) rather than pulled from Bitnami's
// deprecated public registry. Bitnami prunes chart versions (and even
// individual blobs of versions that previously resolved) without notice, which
// caused intermittent 404s in CI. The chart content ships in-repo; only the
// bitnamilegacy container images are fetched at pod runtime.
func WithPostgres(testing *testing.T, env *testsupport.Env) {
	testing.Helper()

	matches, err := filepath.Glob(filepath.Join(ChartPath(testing), "charts", "postgresql-*.tgz"))
	require.NoError(testing, err)
	require.NotEmpty(testing, matches, "vendored postgresql chart tgz not found under charts/zitadel/charts")
	chartPath := matches[0]

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
				"--set-string", "primary.persistence.enabled=false",
				"--set-string", "primary.pgHbaConfiguration=host all all all trust",
				"--set-string", "primary.extendedConfiguration=max_connections = 500",
			},
		},
	}

	require.NoError(testing, helm.UpgradeE(testing, helmOptions, chartPath, "db"))
}
