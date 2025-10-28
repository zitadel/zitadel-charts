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

// IntegrationSuite is the main test suite for ZITADEL acceptance tests.
// It manages the lifecycle of test resources including Kubernetes namespaces,
// database installations, and ZITADEL deployments.
type IntegrationSuite struct {
	suite.Suite
	Log                                                                     *logger.Logger
	KubeOptions                                                             *k8s.KubectlOptions
	KubeClient                                                              *kubernetes.Clientset
	ZitadelValues                                                           []string
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

// Configure creates a new IntegrationSuite instance with the specified
// configuration. It sets up the Kubernetes client, generates unique resource
// names, and registers optional hooks for customizing test behavior.
func Configure(
	t *testing.T,
	namespace, externalDomain string,
	dbChart databaseChart,
	zitadelValues []string,
	before, afterDB, afterZITADEL hookFunc,
) *IntegrationSuite {
	chartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	kubeOptions := k8s.NewKubectlOptions("", "", namespace)
	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, kubeOptions)
	require.NoError(t, err)

	dbReleaseCandidate := "db-" + namespace
	dbRelease := truncateString(dbReleaseCandidate, helmReleaseMaxLen)

	integrationSuite := &IntegrationSuite{
		Log:              logger.New(logger.Terratest),
		KubeOptions:      kubeOptions,
		KubeClient:       clientset,
		ZitadelValues:    zitadelValues,
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

// SetupSuite runs once before all tests in the suite. It adds the database
// Helm repository to cache the repository index, preventing repeated downloads
// across test executions. Retries for up to one minute if the operation fails.
func (suite *IntegrationSuite) SetupSuite() {
	options := &helm.Options{
		KubectlOptions: suite.KubeOptions,
	}

	ctx := context.Background()
	Awaitf(ctx, suite.T(), 1*time.Minute, func(ctx context.Context) error {
		return helm.AddRepoE(suite.T(), options, suite.DbRepoName, suite.DbChart.repoURL)
	}, "adding helm repo %s with URL %s failed", suite.DbRepoName, suite.DbChart.repoURL)
}

// SetupTest runs before each test. It creates the test namespace if it does
// not already exist. If the namespace already exists, it logs a warning but
// continues execution to support test recovery scenarios.
func (suite *IntegrationSuite) SetupTest() {
	_, err := k8s.GetNamespaceE(suite.T(), suite.KubeOptions, suite.KubeOptions.Namespace)

	if errors.IsNotFound(err) {
		k8s.CreateNamespace(suite.T(), suite.KubeOptions, suite.KubeOptions.Namespace)
		return
	}

	require.NoError(suite.T(), err)
	suite.T().Logf("Namespace %s already exists", suite.KubeOptions.Namespace)
}

// BeforeTest runs before each test. It executes the optional beforeFunc hook,
// installs the database Helm chart in the test namespace, and runs the optional
// afterDBFunc hook. The database chart waits up to 10 minutes for successful
// deployment before proceeding.
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

// AfterTest runs after each test. It executes the optional afterZITADELFunc
// hook if the test passed. This hook is typically used for post-deployment
// validation such as connectivity checks or smoke tests.
func (suite *IntegrationSuite) AfterTest(_, _ string) {
	if suite.AfterZITADELFunc == nil || suite.T().Failed() {
		return
	}
	suite.AfterZITADELFunc(suite)
}

// TearDownTest runs after each test. It deletes the test namespace if the
// test passed, cleaning up all resources created during the test. If the test
// failed, the namespace is preserved for debugging purposes.
func (suite *IntegrationSuite) TearDownTest() {
	if suite.T().Failed() {
		suite.T().Logf("Test failed - namespace %s preserved for debugging", suite.KubeOptions.Namespace)
		return
	}
	k8s.DeleteNamespace(suite.T(), suite.KubeOptions, suite.KubeOptions.Namespace)
}
