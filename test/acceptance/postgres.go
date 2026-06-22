package acceptance_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

const postgresRelease = "db"

// PostgresOption configures PostgreSQL installation.
type PostgresOption func(*postgresConfig)

type postgresConfig struct {
	tlsEnabled       bool
	tlsSecretName    string
	postgresPassword string
	database         string
}

// WithPostgresTLS enables TLS for PostgreSQL with the given secret name.
func WithPostgresTLS(secretName string) PostgresOption {
	return func(c *postgresConfig) {
		c.tlsEnabled = true
		c.tlsSecretName = secretName
	}
}

// WithPostgresDatabase sets the default database name created on init.
func WithPostgresDatabase(database string) PostgresOption {
	return func(c *postgresConfig) {
		c.database = database
	}
}

// WithPostgresPassword sets the postgres password.
func WithPostgresPassword(password string) PostgresOption {
	return func(c *postgresConfig) {
		c.postgresPassword = password
	}
}

// vendoredPostgresChart returns the path to the PostgreSQL chart vendored in
// the zitadel chart (charts/zitadel/charts/postgresql-*.tgz). The tests install
// PostgreSQL from this in-repo copy rather than pulling it from Bitnami's
// deprecated public registry, which prunes chart versions (and even individual
// blobs of versions that previously resolved) without notice and caused
// intermittent 404s in CI. Only the bitnamilegacy container images are fetched
// at pod runtime.
func vendoredPostgresChart(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to determine caller info for chart path resolution")
	repoRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	matches, err := filepath.Glob(filepath.Join(repoRoot, "charts", "zitadel", "charts", "postgresql-*.tgz"))
	require.NoError(t, err)
	require.NotEmpty(t, matches, "vendored postgresql chart tgz not found under charts/zitadel/charts")
	return matches[0]
}

// InstallPostgres installs PostgreSQL via Helm into the given namespace from the
// vendored PostgreSQL chart, using legacy images. Persistence is disabled for
// test environments.
func InstallPostgres(t *testing.T, k *k8s.KubectlOptions, opts ...PostgresOption) {
	t.Helper()

	cfg := &postgresConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	values := map[string]string{
		"image.repository":                   "bitnamilegacy/postgresql",
		"volumePermissions.image.repository": "bitnamilegacy/os-shell",
		"metrics.image.repository":           "bitnamilegacy/postgres-exporter",
		"primary.persistence.enabled":        "false",
	}

	if cfg.tlsEnabled {
		values["tls.enabled"] = "true"
		values["tls.certificatesSecret"] = cfg.tlsSecretName
		values["tls.certFilename"] = "tls.crt"
		values["tls.certKeyFilename"] = "tls.key"
		values["volumePermissions.enabled"] = "true"
	} else {
		values["primary.pgHbaConfiguration"] = "host all all all trust"
	}

	if cfg.database != "" {
		values["auth.database"] = cfg.database
	}

	if cfg.postgresPassword != "" {
		values["auth.postgresPassword"] = cfg.postgresPassword
	}

	options := &helm.Options{
		KubectlOptions: k,
		SetValues:      values,
		ExtraArgs:      map[string][]string{"install": {"--wait", "--timeout", "10m", "--hide-notes"}},
	}

	helm.Install(t, options, vendoredPostgresChart(t), postgresRelease)
}
