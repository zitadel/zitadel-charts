package acceptance_test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/suite"
	"github.com/zitadel/zitadel-charts/charts/zitadel/acceptance"
)

func TestPostgresInsecure(t *testing.T) {
	t.Parallel()
	example := "1-postgres-insecure"
	workDir, values := workingDirectory(example)
	suite.Run(t, acceptance.Configure(
		t,
		newNamespaceIdentifier(example),
		acceptance.Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		[]string{values},
		"",
		nil,
		nil,
		nil,
	))
}

func TestPostgresSecure(t *testing.T) {
	t.Parallel()
	example := "2-postgres-secure"
	workDir, values := workingDirectory(example)
	suite.Run(t, acceptance.Configure(
		t,
		newNamespaceIdentifier(example),
		acceptance.Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		[]string{values},
		"",
		func(cfg *acceptance.ConfigurationTest) {
			k8s.KubectlApply(t, cfg.KubeOptions, filepath.Join(workDir, "certs-job.yaml"))
			k8s.WaitUntilJobSucceed(t, cfg.KubeOptions, "create-certs", 120, 3*time.Second)
		},
		nil,
		nil,
	))
}

func TestCockroachInsecure(t *testing.T) {
	t.Parallel()
	example := "3-cockroach-insecure"
	workDir, values := workingDirectory(example)
	suite.Run(t, acceptance.Configure(
		t,
		newNamespaceIdentifier(example),
		acceptance.Cockroach.WithValues(filepath.Join(workDir, "cockroach-values.yaml")),
		[]string{values},
		"",
		nil,
		nil,
		nil,
	))
}

func TestCockroachSecure(t *testing.T) {
	t.Parallel()
	example := "4-cockroach-secure"
	workDir, values := workingDirectory(example)
	suite.Run(t, acceptance.Configure(
		t,
		newNamespaceIdentifier(example),
		acceptance.Cockroach.WithValues(filepath.Join(workDir, "cockroach-values.yaml")),
		[]string{values},
		"",
		nil,
		func(cfg *acceptance.ConfigurationTest) {
			k8s.KubectlApply(t, cfg.KubeOptions, filepath.Join(workDir, "zitadel-cert-job.yaml"))
			k8s.WaitUntilJobSucceed(t, cfg.KubeOptions, "create-zitadel-cert", 120, 3*time.Second)
		},
		nil,
	))
}

func TestReferencedSecrets(t *testing.T) {
	t.Parallel()
	example := "5-referenced-secrets"
	workDir, values := workingDirectory(example)
	suite.Run(t, acceptance.Configure(
		t,
		newNamespaceIdentifier(example),
		acceptance.Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		[]string{values},
		"",
		nil,
		func(cfg *acceptance.ConfigurationTest) {
			k8s.KubectlApply(t, cfg.KubeOptions, filepath.Join(workDir, "zitadel-secrets.yaml"))
			k8s.KubectlApply(t, cfg.KubeOptions, filepath.Join(workDir, "zitadel-masterkey.yaml"))
		},
		nil,
	))
}

func TestMachineUser(t *testing.T) {
	t.Parallel()
	example := "6-machine-user"
	workDir, values := workingDirectory(example)
	saUserame := readValues(t, values).Zitadel.ConfigmapConfig.FirstInstance.Org.Machine.Machine.Username
	suite.Run(t, acceptance.Configure(
		t,
		newNamespaceIdentifier(example),
		acceptance.Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		[]string{values},
		"",
		nil,
		nil,
		testAuthenticatedAPI(saUserame, fmt.Sprintf("%s.json", saUserame))),
	)
}
