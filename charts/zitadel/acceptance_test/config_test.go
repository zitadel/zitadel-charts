package acceptance_test

import (
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes"
)

type hookFunc func(*ConfigurationTest)

type ConfigurationTest struct {
	suite.Suite
	log                                                                 *logger.Logger
	KubeOptions                                                         *k8s.KubectlOptions
	KubeClient                                                          *kubernetes.Clientset
	zitadelValues                                                       []string
	dbChart                                                             databaseChart
	ApiBaseUrl, zitadelChartPath, zitadelRelease, dbRepoName, dbRelease string
	beforeFunc, afterDBFunc, afterZITADELFunc                           hookFunc
}

type databaseChart struct {
	valuesFile, repoUrl, name, version string
	testValues                         map[string]string
}

var (
	Postgres = databaseChart{
		repoUrl: "https://charts.bitnami.com/bitnami",
		name:    "postgresql",
		testValues: map[string]string{
			"image.repository":                   "bitnamilegacy/postgresql",
			"volumePermissions.image.repository": "bitnamilegacy/os-shell",
			// FIX: Override Bitnami's naming to ensure the Service name is short (db-postgresql) and consistent.
			"fullnameOverride": "db-postgresql",
		},
	}
)

const helmReleaseMaxLen = 53 // Helm limit is 53 characters

func (d *databaseChart) WithValues(valuesFile string) databaseChart {
	d.valuesFile = valuesFile
	return *d
}

func Configure(
	t *testing.T,
	namespace, externalDomain string,
	dbChart databaseChart,
	zitadelValues []string,
	before, afterDB, afterZITADEL hookFunc,
) *ConfigurationTest {
	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)
	dbRepoName := "postgres"
	kubeOptions := k8s.NewKubectlOptions("", "", namespace)
	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, kubeOptions)
	if err != nil {
		t.Fatal(err)
	}

	// Create a unique database Helm release name by prepending "db-" to the namespace
	// and truncating it to fit the Helm release name limit of 53 characters.
	dbReleaseCandidate := "db-" + namespace
	dbRelease := truncateString(dbReleaseCandidate, helmReleaseMaxLen)

	cfg := &ConfigurationTest{
		log:              logger.New(logger.Terratest),
		KubeOptions:      kubeOptions,
		KubeClient:       clientset,
		zitadelValues:    zitadelValues,
		zitadelChartPath: chartPath,
		zitadelRelease:   "zitadel-test",
		dbChart:          dbChart,
		dbRepoName:       dbRepoName,
		dbRelease:        dbRelease, // Use the truncated, unique release name
		beforeFunc:       before,
		afterDBFunc:      afterDB,
		afterZITADELFunc: afterZITADEL,
		ApiBaseUrl:       "https://" + externalDomain,
	}
	cfg.SetT(t)
	return cfg
}
