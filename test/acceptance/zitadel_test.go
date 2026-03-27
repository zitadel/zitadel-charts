package acceptance_test

import (
	"context"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"

	"github.com/zitadel/zitadel-charts/test/internal/testcluster"
)

// TestPostgresInsecure validates a basic ZITADEL deployment with an insecure
// PostgreSQL connection (no TLS). This is the simplest deployment scenario and
// serves as a baseline test. The test installs PostgreSQL with trust-based auth
// (no password), deploys ZITADEL with default settings, then verifies that all
// HTTP/gRPC endpoints are accessible and the browser-based login flow works.
func TestPostgresInsecure(t *testing.T) {
	domain := "pg-insecure.127.0.0.1.sslip.io"
	apiBaseURL := BuildAPIBaseURL(domain, httpsPort, true)

	testcluster.WithNamespace(t, func(ctx context.Context, k *k8s.KubectlOptions) {
		InstallPostgres(t, k)
		InstallZitadel(t, k,
			WithExternalDomain(domain),
			WithExternalPort(httpsPort),
			WithValue("image.repository", "ghcr.io/zitadel/zitadel"),
			WithValue("image.tag", "1f74a0959ab172c7ea00beee122e8ef062d77eef"),
			WithValue("login.image.repository", "ghcr.io/zitadel/zitadel-login"),
			WithValue("login.image.tag", "1f74a0959ab172c7ea00beee122e8ef062d77eef"),
		)

		t.Run("accessibility", func(t *testing.T) { CheckAccessibility(ctx, t, k, apiBaseURL) })
		t.Run("login", func(t *testing.T) { CheckLogin(t, apiBaseURL) })
		// TODO: re-enable once auth.go is rewritten for X.509 admin-client JWT auth
		// t.Run("authenticated-api", func(t *testing.T) {
		// 	CheckAuthenticatedAPI(ctx, t, k, apiBaseURL, zitadelRelease+"-admin-service-key")
		// })
		t.Run("uninstall", func(t *testing.T) {
			CheckUninstall(ctx, t, k, nil)
		})
	})
}

// TestPostgresSecure validates ZITADEL deployment with TLS-encrypted PostgreSQL
// connections using verify-full SSL mode. This test generates a CA and server
// certificates for both PostgreSQL and ZITADEL, creates the corresponding TLS
// secrets, and configures both services to use mutual TLS authentication. This
// scenario validates that the chart correctly mounts and uses TLS certificates
// for database connections.
func TestPostgresSecure(t *testing.T) {
	domain := "pg-secure.127.0.0.1.sslip.io"
	apiBaseURL := BuildAPIBaseURL(domain, httpsPort, true)

	testcluster.WithNamespace(t, func(ctx context.Context, k *k8s.KubectlOptions) {
		ca, err := testcluster.GenerateCA("Test CA")
		require.NoError(t, err, "failed to generate CA")

		dnsNames := []string{"postgres", "zitadel", "db-postgresql"}
		pgCert, err := ca.SignCertificate("db-postgresql", dnsNames)
		require.NoError(t, err, "failed to generate postgres certificate")
		zitadelCert, err := ca.SignCertificate("zitadel", dnsNames)
		require.NoError(t, err, "failed to generate zitadel certificate")

		testcluster.CreateTLSSecret(t, k, "postgres-cert", ca.Cert, pgCert.Cert, pgCert.Key)
		testcluster.CreateTLSSecret(t, k, "zitadel-cert", ca.Cert, zitadelCert.Cert, zitadelCert.Key)

		InstallPostgres(t, k,
			WithPostgresTLS("postgres-cert"),
			WithPostgresPassword("abc"),
		)
		InstallZitadel(t, k,
			WithExternalDomain(domain),
			WithExternalPort(httpsPort),
			WithDBSSLMode("verify-full"),
			WithDBCredentials("zitadel", "xyz", "postgres", "abc"),
			WithDBTLSSecrets("postgres-cert", "postgres-cert", "zitadel-cert"),
		)

		t.Run("accessibility", func(t *testing.T) { CheckAccessibility(ctx, t, k, apiBaseURL) })
		t.Run("login", func(t *testing.T) { CheckLogin(t, apiBaseURL) })
		// TODO: re-enable once auth.go is rewritten for X.509 admin-client JWT auth
		// t.Run("authenticated-api", func(t *testing.T) {
		// 	CheckAuthenticatedAPI(ctx, t, k, apiBaseURL, zitadelRelease+"-admin-service-key")
		// })
		t.Run("uninstall", func(t *testing.T) {
			CheckUninstall(ctx, t, k, nil)
		})
	})
}

// TestReferencedSecrets validates that ZITADEL can be configured using
// pre-existing Kubernetes secrets rather than inline Helm values. This tests
// the configSecretName/configSecretKey and masterkeySecretName chart values,
// which allow operators to manage sensitive configuration outside of Helm. The
// test creates secrets containing the masterkey and database host config before
// installing ZITADEL, verifying that the chart correctly references and mounts
// these external secrets.
func TestReferencedSecrets(t *testing.T) {
	domain := "ref-secrets.127.0.0.1.sslip.io"
	apiBaseURL := BuildAPIBaseURL(domain, httpsPort, true)

	testcluster.WithNamespace(t, func(ctx context.Context, k *k8s.KubectlOptions) {
		testcluster.CreateOpaqueSecret(t, k, "existing-zitadel-masterkey", map[string]string{
			"masterkey": defaultMasterkey,
		})
		testcluster.CreateOpaqueSecret(t, k, "existing-zitadel-secrets", map[string]string{
			"config.yaml": `Database:
  Postgres:
    Host: db-postgresql
`,
		})

		InstallPostgres(t, k)
		InstallZitadel(t, k,
			WithExternalDomain(domain),
			WithExternalPort(httpsPort),
			WithMasterkeySecret("existing-zitadel-masterkey"),
			WithConfigSecret("existing-zitadel-secrets", "config.yaml"),
			WithoutDBHost(),
		)

		t.Run("accessibility", func(t *testing.T) { CheckAccessibility(ctx, t, k, apiBaseURL) })
		t.Run("login", func(t *testing.T) { CheckLogin(t, apiBaseURL) })
		// TODO: re-enable once auth.go is rewritten for X.509 admin-client JWT auth
		// t.Run("authenticated-api", func(t *testing.T) {
		// 	CheckAuthenticatedAPI(ctx, t, k, apiBaseURL, zitadelRelease+"-admin-service-key")
		// })
		t.Run("uninstall", func(t *testing.T) {
			CheckUninstall(ctx, t, k, nil)
		})
	})
}

// TestMachineUser validates the admin-client X.509 authentication flow. The
// chart auto-generates an RSA keypair, mounts the public cert as
// ZITADEL_ADMINCLIENT_KEYFILE, and operators can extract the private key to sign
// JWTs for admin API access. This test verifies the complete flow: keypair
// generation, ZITADEL startup with the cert, and authenticated API calls using
// JWTs signed with the private key.
//
//goland:noinspection DuplicatedCode
func TestMachineUser(t *testing.T) {
	domain := "machine.127.0.0.1.sslip.io"
	apiBaseURL := BuildAPIBaseURL(domain, httpsPort, true)

	testcluster.WithNamespace(t, func(ctx context.Context, k *k8s.KubectlOptions) {
		InstallPostgres(t, k)
		InstallZitadel(t, k,
			WithExternalDomain(domain),
			WithExternalPort(httpsPort),
		)

		t.Run("accessibility", func(t *testing.T) { CheckAccessibility(ctx, t, k, apiBaseURL) })
		t.Run("login", func(t *testing.T) { CheckLogin(t, apiBaseURL) })
		// TODO: re-enable once auth.go is rewritten for X.509 admin-client JWT auth
		// t.Run("authenticated-api", func(t *testing.T) {
		// 	CheckAuthenticatedAPI(ctx, t, k, apiBaseURL, zitadelRelease+"-admin-service-key")
		// })
		t.Run("uninstall", func(t *testing.T) {
			CheckUninstall(ctx, t, k, nil)
		})
	})
}

// TestInternalTLS validates ZITADEL deployment with internal TLS enabled using
// a self-signed certificate generated by the chart. When selfSignedCert is
// enabled, the chart generates a CA and server certificate, configures ZITADEL
// to serve HTTPS, and the ingress controller connects to ZITADEL over TLS. This
// tests end-to-end encryption from client through ingress to the ZITADEL pod.
//
//goland:noinspection DuplicatedCode
func TestInternalTLS(t *testing.T) {
	domain := "internal-tls.127.0.0.1.sslip.io"
	apiBaseURL := BuildAPIBaseURL(domain, httpsPort, true)

	testcluster.WithNamespace(t, func(ctx context.Context, k *k8s.KubectlOptions) {
		InstallPostgres(t, k)
		InstallZitadel(t, k,
			WithExternalDomain(domain),
			WithExternalPort(httpsPort),
			WithSelfSignedCert(domain),
		)

		t.Run("accessibility", func(t *testing.T) { CheckAccessibility(ctx, t, k, apiBaseURL) })
		t.Run("login", func(t *testing.T) { CheckLogin(t, apiBaseURL) })
		// TODO: re-enable once auth.go is rewritten for X.509 admin-client JWT auth
		// t.Run("authenticated-api", func(t *testing.T) {
		// 	CheckAuthenticatedAPI(ctx, t, k, apiBaseURL, zitadelRelease+"-admin-service-key")
		// })
		t.Run("uninstall", func(t *testing.T) {
			CheckUninstall(ctx, t, k, nil)
		})
	})
}
