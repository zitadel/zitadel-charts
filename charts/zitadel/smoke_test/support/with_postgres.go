package support

import (
	"fmt"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
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
		ExtraArgs: map[string][]string{
			"upgrade": {
				"--install",
				"--wait", "--timeout", "10m",
				"--set-string", "primary.persistence.enabled=false",
				"--set-string", "primary.pgHbaConfiguration=host all all all trust",
				"--set-string", "primary.extendedConfiguration=max_connections = 500",
			},
		},
	}

	chartName := fmt.Sprintf("%s/postgresql", repoAlias)
	releaseName := "db"

	require.NoError(testing, helm.UpgradeE(testing, helmOptions, chartName, releaseName))

	k8s.WaitUntilServiceAvailable(testing, env.Kube, "db-postgresql", 60, 5*time.Second)
	k8s.WaitUntilPodAvailable(testing, env.Kube, "db-postgresql-0", 60, 5*time.Second)
}
