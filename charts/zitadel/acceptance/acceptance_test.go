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
	workDir, valuesFile, values := readConfig(t, example)
	cfg := values.Zitadel.ConfigmapConfig
	suite.Run(t, acceptance.Configure(
		t,
		newNamespaceIdentifier(example),
		acceptance.Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		[]string{valuesFile},
		cfg.ExternalDomain,
		cfg.ExternalPort,
		cfg.ExternalSecure,
		nil,
		nil,
		nil,
	))
}

func TestPostgresSecure(t *testing.T) {
	t.Parallel()
	example := "2-postgres-secure"
	workDir, valuesFile, values := readConfig(t, example)
	cfg := values.Zitadel.ConfigmapConfig
	suite.Run(t, acceptance.Configure(
		t,
		newNamespaceIdentifier(example),
		acceptance.Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		[]string{valuesFile},
		cfg.ExternalDomain,
		cfg.ExternalPort,
		cfg.ExternalSecure,
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
	workDir, valuesFile, values := readConfig(t, example)
	cfg := values.Zitadel.ConfigmapConfig
	suite.Run(t, acceptance.Configure(
		t,
		newNamespaceIdentifier(example),
		acceptance.Cockroach.WithValues(filepath.Join(workDir, "cockroach-values.yaml")),
		[]string{valuesFile},
		cfg.ExternalDomain,
		cfg.ExternalPort,
		cfg.ExternalSecure,
		nil,
		nil,
		nil,
	))
}

func TestCockroachSecure(t *testing.T) {
	t.Parallel()
	example := "4-cockroach-secure"
	workDir, valuesFile, values := readConfig(t, example)
	cfg := values.Zitadel.ConfigmapConfig
	suite.Run(t, acceptance.Configure(
		t,
		newNamespaceIdentifier(example),
		acceptance.Cockroach.WithValues(filepath.Join(workDir, "cockroach-values.yaml")),
		[]string{valuesFile},
		cfg.ExternalDomain,
		cfg.ExternalPort,
		cfg.ExternalSecure,
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
	workDir, valuesFile, values := readConfig(t, example)
	cfg := values.Zitadel.ConfigmapConfig
	suite.Run(t, acceptance.Configure(
		t,
		newNamespaceIdentifier(example),
		acceptance.Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		[]string{valuesFile},
		cfg.ExternalDomain,
		cfg.ExternalPort,
		cfg.ExternalSecure,
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
	workDir, valuesFile, values := readConfig(t, example)
	cfg := values.Zitadel.ConfigmapConfig
	saUsername := cfg.FirstInstance.Org.Machine.Machine.Username
	suite.Run(t, acceptance.Configure(
		t,
		newNamespaceIdentifier(example),
		acceptance.Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		[]string{valuesFile},
		cfg.ExternalDomain,
		cfg.ExternalPort,
		cfg.ExternalSecure,
		nil,
		nil,
		testAuthenticatedAPI(saUsername, fmt.Sprintf("%s.json", saUsername))),
	)
}

func TestSelfSigned(t *testing.T) {
	t.Parallel()
	example := "7-self-signed"
	workDir, valuesFile, values := readConfig(t, example)
	cfg := values.Zitadel.ConfigmapConfig
	suite.Run(t, acceptance.Configure(
		t,
		newNamespaceIdentifier(example),
		acceptance.Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		[]string{valuesFile},
		cfg.ExternalDomain,
		cfg.ExternalPort,
		cfg.ExternalSecure,
		nil,
		nil,
		nil,
	))
}
