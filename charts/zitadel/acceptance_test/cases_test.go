package acceptance_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/suite"
)

func TestPostgresInsecure(t *testing.T) {
	t.Parallel()
	example := "1-postgres-insecure"
	workDir, valuesFile := workingDirectory(example)

	suite.Run(t, Configure(
		t,
		newNamespaceIdentifier(example),
		readExternalDomain(t, valuesFile),
		Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		valuesFile,
		nil,
		nil,
		nil,
		nil,
	))
}

func TestPostgresSecure(t *testing.T) {
	t.Parallel()
	example := "2-postgres-secure"
	workDir, valuesFile := workingDirectory(example)

	suite.Run(t, Configure(
		t,
		newNamespaceIdentifier(example),
		readExternalDomain(t, valuesFile),
		Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		valuesFile,
		nil,
		func(cfg *IntegrationSuite) {
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
	workDir, valuesFile := workingDirectory(example)

	suite.Run(t, Configure(
		t,
		newNamespaceIdentifier(example),
		readExternalDomain(t, valuesFile),
		Postgres.WithValues(filepath.Join(workDir, "postgres-values.yaml")),
		valuesFile,
		nil,
		nil,
		func(cfg *IntegrationSuite) {
			k8s.KubectlApply(t, cfg.KubeOptions, filepath.Join(workDir, "zitadel-secrets.yaml"))
			k8s.KubectlApply(t, cfg.KubeOptions, filepath.Join(workDir, "zitadel-masterkey.yaml"))
		},
		nil,
	))
}

func readExternalDomain(t *testing.T, valuesFile string) string {
	valuesMap := readValuesAsMap(t, valuesFile)
	return valuesMap["zitadel.configmapConfig.ExternalDomain"]
}
