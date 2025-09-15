package acceptance_test

import (
	"fmt"
	"path/filepath"
	"strings"
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
		version: "12.10.0",
	}
	CloudNativePG = databaseChart{
		repoUrl: "https://cloudnative-pg.github.io/charts",
		name:    "cloudnative-pg",
		version: "0.21.6",
	}
)

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
	dbRepoName := fmt.Sprintf("crdb-%s", strings.TrimPrefix(namespace, "zitadel-helm-"))
	kubeOptions := k8s.NewKubectlOptions("", "", namespace)
	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, kubeOptions)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &ConfigurationTest{
		log:              logger.New(logger.Terratest),
		KubeOptions:      kubeOptions,
		KubeClient:       clientset,
		zitadelValues:    zitadelValues,
		zitadelChartPath: chartPath,
		zitadelRelease:   "zitadel-test",
		dbChart:          dbChart,
		dbRepoName:       dbRepoName,
		dbRelease:        "db",
		beforeFunc:       before,
		afterDBFunc:      afterDB,
		afterZITADELFunc: afterZITADEL,
		ApiBaseUrl:       "https://" + externalDomain,
	}
	cfg.SetT(t)
	return cfg
}
