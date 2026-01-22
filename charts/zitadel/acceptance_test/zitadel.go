package acceptance_test

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
)

const (
	zitadelRelease   = "zitadel-test"
	defaultMasterkey = "x123456789012345678901234567891y"
)

// ZitadelOption configures ZITADEL installation.
type ZitadelOption func(*zitadelConfig)

type zitadelConfig struct {
	externalDomain      string
	externalPort        string
	tlsEnabled          bool
	selfSignedCert      bool
	masterkeySecretName string
	configSecretName    string
	configSecretKey     string
	dbSSLMode           string
	dbHost              string
	skipDBHost          bool
	dbUser              string
	dbAdminUser         string
	dbPassword          string
	dbAdminPassword     string
	dbSslCaCrtSecret    string
	dbSslAdminCrtSecret string
	dbSslUserCrtSecret  string
	machineUserName     string
	machineUserUsername string
	additionalValues    map[string]string
}

// WithExternalDomain sets the external domain for ZITADEL.
func WithExternalDomain(domain string) ZitadelOption {
	return func(c *zitadelConfig) {
		c.externalDomain = domain
	}
}

// WithExternalPort sets the external port for ZITADEL.
func WithExternalPort(port string) ZitadelOption {
	return func(c *zitadelConfig) {
		c.externalPort = port
	}
}

// WithSelfSignedCert enables self-signed certificate generation for ZITADEL.
// This also enables TLS and configures the additional DNS name for the cert.
func WithSelfSignedCert(additionalDNSName string) ZitadelOption {
	return func(c *zitadelConfig) {
		c.selfSignedCert = true
		c.tlsEnabled = true
		if additionalDNSName != "" {
			c.additionalValues["zitadel.selfSignedCert.additionalDnsName"] = additionalDNSName
		}
	}
}

// WithMasterkeySecret references an existing secret for the ZITADEL masterkey.
func WithMasterkeySecret(secretName string) ZitadelOption {
	return func(c *zitadelConfig) {
		c.masterkeySecretName = secretName
	}
}

// WithConfigSecret references an existing secret for ZITADEL configuration.
func WithConfigSecret(secretName, key string) ZitadelOption {
	return func(c *zitadelConfig) {
		c.configSecretName = secretName
		c.configSecretKey = key
	}
}

// WithDBSSLMode sets the SSL mode for database connections.
func WithDBSSLMode(mode string) ZitadelOption {
	return func(c *zitadelConfig) {
		c.dbSSLMode = mode
	}
}

// WithDBCredentials sets database user and admin credentials.
func WithDBCredentials(user, password, adminUser, adminPassword string) ZitadelOption {
	return func(c *zitadelConfig) {
		c.dbUser = user
		c.dbPassword = password
		c.dbAdminUser = adminUser
		c.dbAdminPassword = adminPassword
	}
}

// WithDBTLSSecrets sets the TLS secrets for database connections.
func WithDBTLSSecrets(caCrtSecret, adminCrtSecret, userCrtSecret string) ZitadelOption {
	return func(c *zitadelConfig) {
		c.dbSslCaCrtSecret = caCrtSecret
		c.dbSslAdminCrtSecret = adminCrtSecret
		c.dbSslUserCrtSecret = userCrtSecret
	}
}

// WithMachineUser configures a machine user for API access.
func WithMachineUser(name, username string) ZitadelOption {
	return func(c *zitadelConfig) {
		c.machineUserName = name
		c.machineUserUsername = username
	}
}

// WithoutDBHost skips setting the database host, useful when the host is
// provided via a config secret instead of Helm values.
func WithoutDBHost() ZitadelOption {
	return func(c *zitadelConfig) {
		c.skipDBHost = true
	}
}

// InstallZitadel installs ZITADEL via Helm with the provided options. The chart
// is installed from the local filesystem relative to this test file location.
// The install blocks until all resources are ready (--wait --timeout 10m).
func InstallZitadel(t *testing.T, k *k8s.KubectlOptions, opts ...ZitadelOption) {
	t.Helper()

	cfg := &zitadelConfig{
		externalPort:     "443",
		dbSSLMode:        "disable",
		dbHost:           "db-postgresql",
		dbUser:           "postgres",
		dbAdminUser:      "postgres",
		additionalValues: make(map[string]string),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	_, filename, _, _ := runtime.Caller(0)
	chartPath := filepath.Join(filename, "..", "..")

	values := map[string]string{
		"replicaCount":          "1",
		"login.replicaCount":    "1",
		"pdb.enabled":           "true",
		"ingress.enabled":       "true",
		"login.ingress.enabled": "true",
	}

	if cfg.externalDomain != "" {
		values["zitadel.configmapConfig.ExternalDomain"] = cfg.externalDomain
	}
	if cfg.externalPort != "" {
		values["zitadel.configmapConfig.ExternalPort"] = cfg.externalPort
	}

	if cfg.tlsEnabled {
		values["zitadel.configmapConfig.TLS.Enabled"] = "true"
	} else {
		values["zitadel.configmapConfig.TLS.Enabled"] = "false"
	}

	if cfg.selfSignedCert {
		values["zitadel.selfSignedCert.enabled"] = "true"
		values["service.annotations.traefik\\.ingress\\.kubernetes\\.io/service\\.serversscheme"] = "https"
	}

	if cfg.masterkeySecretName != "" {
		values["zitadel.masterkeySecretName"] = cfg.masterkeySecretName
	} else {
		values["zitadel.masterkey"] = defaultMasterkey
	}

	if cfg.configSecretName != "" {
		values["zitadel.configSecretName"] = cfg.configSecretName
		if cfg.configSecretKey != "" {
			values["zitadel.configSecretKey"] = cfg.configSecretKey
		}
	}

	if !cfg.skipDBHost {
		values["zitadel.configmapConfig.Database.Postgres.Host"] = cfg.dbHost
	}
	values["zitadel.configmapConfig.Database.Postgres.Port"] = "5432"
	values["zitadel.configmapConfig.Database.Postgres.Database"] = "zitadel"
	values["zitadel.configmapConfig.Database.Postgres.MaxOpenConns"] = "20"
	values["zitadel.configmapConfig.Database.Postgres.MaxIdleConns"] = "10"
	values["zitadel.configmapConfig.Database.Postgres.MaxConnLifetime"] = "30m"
	values["zitadel.configmapConfig.Database.Postgres.MaxConnIdleTime"] = "5m"
	values["zitadel.configmapConfig.Database.Postgres.User.Username"] = cfg.dbUser
	values["zitadel.configmapConfig.Database.Postgres.User.SSL.Mode"] = cfg.dbSSLMode
	values["zitadel.configmapConfig.Database.Postgres.Admin.Username"] = cfg.dbAdminUser
	values["zitadel.configmapConfig.Database.Postgres.Admin.SSL.Mode"] = cfg.dbSSLMode

	if cfg.dbPassword != "" {
		values["zitadel.secretConfig.Database.Postgres.User.Password"] = cfg.dbPassword
	}
	if cfg.dbAdminPassword != "" {
		values["zitadel.secretConfig.Database.Postgres.Admin.Password"] = cfg.dbAdminPassword
	}

	if cfg.dbSslCaCrtSecret != "" {
		values["zitadel.dbSslCaCrtSecret"] = cfg.dbSslCaCrtSecret
	}
	if cfg.dbSslAdminCrtSecret != "" {
		values["zitadel.dbSslAdminCrtSecret"] = cfg.dbSslAdminCrtSecret
	}
	if cfg.dbSslUserCrtSecret != "" {
		values["zitadel.dbSslUserCrtSecret"] = cfg.dbSslUserCrtSecret
	}

	if cfg.machineUserUsername != "" {
		values["zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Username"] = cfg.machineUserUsername
		values["zitadel.configmapConfig.FirstInstance.Org.Machine.Machine.Name"] = cfg.machineUserName
		values["zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.ExpirationDate"] = "2029-01-01T00:00:00Z"
		values["zitadel.configmapConfig.FirstInstance.Org.Machine.MachineKey.Type"] = "1"
		values["zitadel.configmapConfig.Log.Level"] = "debug"
	}

	for k, v := range cfg.additionalValues {
		values[k] = v
	}

	options := &helm.Options{
		KubectlOptions: k,
		SetValues:      values,
		ExtraArgs:      map[string][]string{"install": {"--wait", "--timeout", "10m"}},
	}

	helm.Install(t, options, chartPath, zitadelRelease)
}

// BuildAPIBaseURL constructs the API base URL from domain and port. It uses
// HTTPS when TLS is enabled or when the port is 443.
func BuildAPIBaseURL(domain, port string, useTLS bool) string {
	scheme := "http"
	if useTLS || port == "443" {
		scheme = "https"
	}
	if port != "" {
		return fmt.Sprintf("%s://%s:%s", scheme, domain, port)
	}
	return fmt.Sprintf("%s://%s", scheme, domain)
}
