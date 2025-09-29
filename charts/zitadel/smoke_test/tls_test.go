package smoke_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zitadel/zitadel-charts/charts/zitadel/smoke_test/support"
)

// TestTLSHelpers validates that the test infrastructure components work
// correctly together before running full integration tests. This test
// verifies that namespaces are created, PostgreSQL is deployed, and TLS
// secrets are generated without errors. It does not deploy Zitadel itself.
func TestTLSHelpers(t *testing.T) {
	t.Parallel()

	cluster := support.ConnectCluster(t)

	support.WithNamespace(t, cluster, func(env *support.Env) {
		env.Logger.Logf(t, "namespace %q created successfully", env.Namespace)

		support.WithPostgres(t, env)
		env.Logger.Logf(t, "PostgreSQL installed successfully")

		support.WithTLSSecret(t, env, "zitadel-tls", "zitadel.dev.mrida.ng")
		env.Logger.Logf(t, "zitadel-tls secret created successfully")

		support.WithTLSSecret(t, env, "login-tls", "login.dev.mrida.ng")
		env.Logger.Logf(t, "login-tls secret created successfully")

		require.NotEmpty(t, env.Namespace)
		env.Logger.Logf(t, "test infrastructure validated successfully")
	})
}
