package acceptance_test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/suite"
)

func TestPostgresInsecure(t *testing.T) {
	t.Parallel()
	example := "1-postgres-insecure"
	workDir, valuesFile, values := readConfig(t, example)
	suite.Run(t, Configure(
		t,
		newNamespaceIdentifier(example),
		values.Zitadel.ConfigmapConfig.ExternalDomain,
		Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		[]string{valuesFile},
		nil,
		nil,
		nil,
	))
}

func TestPostgresSecure(t *testing.T) {
	t.Parallel()
	example := "2-postgres-secure"
	workDir, valuesFile, values := readConfig(t, example)
	suite.Run(t, Configure(
		t,
		newNamespaceIdentifier(example),
		values.Zitadel.ConfigmapConfig.ExternalDomain,
		Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		[]string{valuesFile},
		func(cfg *ConfigurationTest) {
			k8s.KubectlApply(t, cfg.KubeOptions, filepath.Join(workDir, "certs-job.yaml"))
			k8s.WaitUntilJobSucceed(t, cfg.KubeOptions, "create-certs", 120, 3*time.Second)
		},
		nil,
		nil,
	))
}

func TestReferencedSecrets(t *testing.T) {
	t.Parallel()
	example := "3-referenced-secrets"
	workDir, valuesFile, values := readConfig(t, example)
	suite.Run(t, Configure(
		t,
		newNamespaceIdentifier(example),
		values.Zitadel.ConfigmapConfig.ExternalDomain,
		Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		[]string{valuesFile},
		nil,
		func(cfg *ConfigurationTest) {
			k8s.KubectlApply(t, cfg.KubeOptions, filepath.Join(workDir, "zitadel-secrets.yaml"))
			k8s.KubectlApply(t, cfg.KubeOptions, filepath.Join(workDir, "zitadel-masterkey.yaml"))
		},
		nil,
	))
}

func TestMachineUser(t *testing.T) {
	t.Parallel()
	example := "4-machine-user"
	workDir, valuesFile, values := readConfig(t, example)
	cfg := values.Zitadel.ConfigmapConfig
	saUsername := cfg.FirstInstance.Org.Machine.Machine.Username
	suite.Run(t, Configure(
		t,
		newNamespaceIdentifier(example),
		values.Zitadel.ConfigmapConfig.ExternalDomain,
		Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		[]string{valuesFile},
		nil,
		nil,
		testAuthenticatedAPI(saUsername, fmt.Sprintf("%s.json", saUsername))),
	)
}

func TestInternalTLS(t *testing.T) {
	t.Parallel()
	example := "5-internal-tls"
	workDir, valuesFile, values := readConfig(t, example)
	suite.Run(t, Configure(
		t,
		newNamespaceIdentifier(example),
		values.Zitadel.ConfigmapConfig.ExternalDomain,
		Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		[]string{valuesFile},
		nil,
		nil,
		nil,
	))
}
