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
	machineUsername := "zitadel-admin-sa"

	testcluster.WithNamespace(t, func(ctx context.Context, k *k8s.KubectlOptions) {
		InstallPostgres(t, k)
		InstallZitadel(t, k,
			WithExternalDomain(domain),
			WithExternalPort(httpsPort),
			WithMachineUser("Admin", machineUsername),
		)

		t.Run("accessibility", func(t *testing.T) { CheckAccessibility(ctx, t, k, apiBaseURL) })
		t.Run("metrics", func(t *testing.T) { CheckMetrics(ctx, t, k, false) })
		t.Run("login", func(t *testing.T) { CheckLogin(t, apiBaseURL) })
		t.Run("authenticated-api", func(t *testing.T) {
			CheckAuthenticatedAPI(ctx, t, k, apiBaseURL, machineUsername, machineUsername+".json")
		})
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
	machineUsername := "zitadel-admin-sa"

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
			WithMachineUser("Admin", machineUsername),
		)

		t.Run("accessibility", func(t *testing.T) { CheckAccessibility(ctx, t, k, apiBaseURL) })
		t.Run("metrics", func(t *testing.T) { CheckMetrics(ctx, t, k, false) })
		t.Run("login", func(t *testing.T) { CheckLogin(t, apiBaseURL) })
		t.Run("authenticated-api", func(t *testing.T) {
			CheckAuthenticatedAPI(ctx, t, k, apiBaseURL, machineUsername, machineUsername+".json")
		})
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
	machineUsername := "zitadel-admin-sa"

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
			WithMachineUser("Admin", machineUsername),
		)

		t.Run("accessibility", func(t *testing.T) { CheckAccessibility(ctx, t, k, apiBaseURL) })
		t.Run("metrics", func(t *testing.T) { CheckMetrics(ctx, t, k, false) })
		t.Run("login", func(t *testing.T) { CheckLogin(t, apiBaseURL) })
		t.Run("authenticated-api", func(t *testing.T) {
			CheckAuthenticatedAPI(ctx, t, k, apiBaseURL, machineUsername, machineUsername+".json")
		})
		t.Run("uninstall", func(t *testing.T) {
			CheckUninstall(ctx, t, k, nil)
		})
	})
}

// TestMachineUser validates the machine user provisioning feature which creates
// a service account during ZITADEL setup. Machine users enable machine-to-
// machine authentication via JWT profile assertions. This test configures a
// machine user, verifies that the service account key is created as a
// Kubernetes secret, and then uses that key to authenticate against both the
// HTTP and gRPC management APIs. This validates the complete M2M auth flow.
//
//goland:noinspection DuplicatedCode
func TestMachineUser(t *testing.T) {
	domain := "machine.127.0.0.1.sslip.io"
	apiBaseURL := BuildAPIBaseURL(domain, httpsPort, true)
	machineUsername := "zitadel-admin-sa"

	testcluster.WithNamespace(t, func(ctx context.Context, k *k8s.KubectlOptions) {
		InstallPostgres(t, k)
		InstallZitadel(t, k,
			WithExternalDomain(domain),
			WithExternalPort(httpsPort),
			WithMachineUser("Admin", machineUsername),
		)

		t.Run("accessibility", func(t *testing.T) { CheckAccessibility(ctx, t, k, apiBaseURL) })
		t.Run("metrics", func(t *testing.T) { CheckMetrics(ctx, t, k, false) })
		t.Run("login", func(t *testing.T) { CheckLogin(t, apiBaseURL) })
		t.Run("authenticated-api", func(t *testing.T) {
			CheckAuthenticatedAPI(ctx, t, k, apiBaseURL, machineUsername, machineUsername+".json")
		})
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
	machineUsername := "zitadel-admin-sa"

	testcluster.WithNamespace(t, func(ctx context.Context, k *k8s.KubectlOptions) {
		InstallPostgres(t, k)
		InstallZitadel(t, k,
			WithExternalDomain(domain),
			WithExternalPort(httpsPort),
			WithSelfSignedCert(domain),
			WithMachineUser("Admin", machineUsername),
		)

		t.Run("accessibility", func(t *testing.T) { CheckAccessibility(ctx, t, k, apiBaseURL) })
		t.Run("metrics", func(t *testing.T) { CheckMetrics(ctx, t, k, true) })
		t.Run("login", func(t *testing.T) { CheckLogin(t, apiBaseURL) })
		t.Run("authenticated-api", func(t *testing.T) {
			CheckAuthenticatedAPI(ctx, t, k, apiBaseURL, machineUsername, machineUsername+".json")
		})
		t.Run("uninstall", func(t *testing.T) {
			CheckUninstall(ctx, t, k, nil)
		})
	})
}

// TestGatewayAPI validates that ZITADEL is accessible through Gateway API
// HTTPRoute resources instead of traditional Kubernetes Ingress. This test
// verifies that the chart's Gateway API integration works correctly with
// Traefik as the Gateway controller, including proper appProtocol handling
// for HTTP/2 backends (issue #580).
//
//goland:noinspection DuplicatedCode
func TestGatewayAPI(t *testing.T) {
	domain := "gateway.127.0.0.1.sslip.io"
	apiBaseURL := BuildAPIBaseURL(domain, httpPort, false)
	machineUsername := "zitadel-admin-sa"

	testcluster.WithNamespace(t, func(ctx context.Context, k *k8s.KubectlOptions) {
		InstallPostgres(t, k)
		InstallZitadel(t, k,
			WithExternalDomain(domain),
			WithExternalPort(httpPort),
			WithExternalSecure(false),
			WithGateway("traefik-gateway", "kube-system"),
			WithMachineUser("Admin", machineUsername),
		)

		t.Run("accessibility", func(t *testing.T) { CheckAccessibility(ctx, t, k, apiBaseURL) })
		t.Run("metrics", func(t *testing.T) { CheckMetrics(ctx, t, k, false) })
		t.Run("authenticated-api", func(t *testing.T) {
			CheckAuthenticatedAPI(ctx, t, k, apiBaseURL, machineUsername, machineUsername+".json")
		})
		t.Run("uninstall", func(t *testing.T) {
			CheckUninstall(ctx, t, k, nil)
		})
	})
}
