package acceptance_test

import (
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	postgresRepoURL  = "https://charts.bitnami.com/bitnami"
	postgresRepoName = "bitnami"
	postgresChart    = "postgresql"
	postgresRelease  = "db"
)

// PostgresOption configures PostgreSQL installation.
type PostgresOption func(*postgresConfig)

type postgresConfig struct {
	tlsEnabled       bool
	tlsSecretName    string
	postgresPassword string
}

// WithPostgresTLS enables TLS for PostgreSQL with the given secret name.
func WithPostgresTLS(secretName string) PostgresOption {
	return func(c *postgresConfig) {
		c.tlsEnabled = true
		c.tlsSecretName = secretName
	}
}

// WithPostgresPassword sets the postgres password.
func WithPostgresPassword(password string) PostgresOption {
	return func(c *postgresConfig) {
		c.postgresPassword = password
	}
}

// InstallPostgres installs PostgreSQL via Helm into the given namespace. It
// uses the Bitnami PostgreSQL chart with legacy images for compatibility with
// older Kubernetes versions. Persistence is disabled for test environments.
func InstallPostgres(t *testing.T, k *k8s.KubectlOptions, opts ...PostgresOption) {
	t.Helper()

	cfg := &postgresConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		err := helm.AddRepoE(t, &helm.Options{}, postgresRepoName, postgresRepoURL)
		if !assert.NoError(collect, err) {
			t.Logf("retrying helm add repo in a second")
		}
	}, 1*time.Minute, time.Second, "adding helm repo failed for a minute")

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

	if cfg.postgresPassword != "" {
		values["auth.postgresPassword"] = cfg.postgresPassword
	}

	options := &helm.Options{
		KubectlOptions: k,
		SetValues:      values,
		ExtraArgs:      map[string][]string{"install": {"--wait", "--timeout", "10m", "--hide-notes"}},
	}

	helm.Install(t, options, postgresRepoName+"/"+postgresChart, postgresRelease)
}
