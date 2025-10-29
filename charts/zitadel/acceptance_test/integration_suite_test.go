package acceptance_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
)

type hookFunc func(*IntegrationSuite)

type IntegrationSuite struct {
	suite.Suite
	Log                                                                     *logger.Logger
	KubeOptions                                                             *k8s.KubectlOptions
	KubeClient                                                              *kubernetes.Clientset
	SetValues                                                               map[string]string
	DbChart                                                                 databaseChart
	ApiBaseURL, ZitadelChartPath, ZitadelRelease, DbRepoName, DbReleaseName string
	BeforeFunc, AfterDBFunc, AfterZITADELFunc                               hookFunc
}

type databaseChart struct {
	valuesFile, repoURL, name, version string
	testValues                         map[string]string
}

var (
	Postgres = databaseChart{
		repoURL: "https://charts.bitnami.com/bitnami",
		name:    "postgresql",
		testValues: map[string]string{
			"auth.postgresPassword":              "postgres",
			"image.repository":                   "bitnamilegacy/postgresql",
			"volumePermissions.image.repository": "bitnamilegacy/os-shell",
			"fullnameOverride":                   "db-postgresql",
		},
	}
)

const helmReleaseMaxLen = 53

func (d *databaseChart) WithValues(valuesFile string) databaseChart {
	d.valuesFile = valuesFile
	return *d
}

func Configure(
	t *testing.T,
	namespace, externalDomain string,
	dbChart databaseChart,
	valuesFile string,
	setValues map[string]string,
	before, afterDB, afterZITADEL hookFunc,
) *IntegrationSuite {
	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	kubeOptions := k8s.NewKubectlOptions("", "", namespace)
	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, kubeOptions)
	require.NoError(t, err)

	dbReleaseCandidate := "db-" + namespace
	dbRelease := truncateString(dbReleaseCandidate, helmReleaseMaxLen)

	mergedSetValues := make(map[string]string)

	if valuesFile != "" {
		fileValues := readValuesAsMap(t, valuesFile)
		for k, v := range fileValues {
			mergedSetValues[k] = v
		}
	}

	for k, v := range setValues {
		mergedSetValues[k] = v
	}

	integrationSuite := &IntegrationSuite{
		Log:              logger.New(logger.Terratest),
		KubeOptions:      kubeOptions,
		KubeClient:       clientset,
		SetValues:        mergedSetValues,
		ZitadelChartPath: chartPath,
		ZitadelRelease:   "zitadel-test",
		DbChart:          dbChart,
		DbRepoName:       "postgres",
		DbReleaseName:    dbRelease,
		BeforeFunc:       before,
		AfterDBFunc:      afterDB,
		AfterZITADELFunc: afterZITADEL,
		ApiBaseURL:       "https://" + externalDomain,
	}
	integrationSuite.SetT(t)
	return integrationSuite
}

func (suite *IntegrationSuite) SetupSuite() {
	options := &helm.Options{
		KubectlOptions: suite.KubeOptions,
	}

	ctx := context.Background()
	Awaitf(ctx, suite.T(), 1*time.Minute, func(ctx context.Context) error {
		return helm.AddRepoE(suite.T(), options, suite.DbRepoName, suite.DbChart.repoURL)
	}, "adding helm repo %s with URL %s failed", suite.DbRepoName, suite.DbChart.repoURL)
}

func (suite *IntegrationSuite) SetupTest() {
	_, err := k8s.GetNamespaceE(suite.T(), suite.KubeOptions, suite.KubeOptions.Namespace)

	if errors.IsNotFound(err) {
		k8s.CreateNamespace(suite.T(), suite.KubeOptions, suite.KubeOptions.Namespace)
		return
	}

	require.NoError(suite.T(), err)
	suite.T().Logf("Namespace %s already exists", suite.KubeOptions.Namespace)
}

func (suite *IntegrationSuite) BeforeTest(_, _ string) {
	if suite.BeforeFunc != nil {
		suite.BeforeFunc(suite)
	}

	options := &helm.Options{
		KubectlOptions: suite.KubeOptions,
		Version:        suite.DbChart.version,
		SetValues:      suite.DbChart.testValues,
		ExtraArgs:      map[string][]string{"install": {"--wait", "--timeout", "10m", "--hide-notes"}},
	}

	if suite.DbChart.valuesFile != "" {
		options.ValuesFiles = []string{suite.DbChart.valuesFile}
	}

	helm.Install(suite.T(), options, suite.DbRepoName+"/"+suite.DbChart.name, suite.DbReleaseName)

	if suite.AfterDBFunc != nil {
		suite.AfterDBFunc(suite)
	}
}

func (suite *IntegrationSuite) AfterTest(_, _ string) {
	if suite.AfterZITADELFunc == nil || suite.T().Failed() {
		return
	}
	suite.AfterZITADELFunc(suite)
}

func (suite *IntegrationSuite) TearDownTest() {
	if suite.T().Failed() {
		suite.T().Logf("Test failed - namespace %s preserved for debugging", suite.KubeOptions.Namespace)
		return
	}
	k8s.DeleteNamespace(suite.T(), suite.KubeOptions, suite.KubeOptions.Namespace)
}
