package installation

import (
	"context"
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
	Ctx              context.Context
	log              *logger.Logger
	KubeOptions      *k8s.KubectlOptions
	KubeClient       *kubernetes.Clientset
	zitadelValues    []string
	zitadelChartPath string
	zitadelRelease   string
	dbChart          databaseChart
	dbRepoName       string
	dbRelease        string
	beforeFunc       hookFunc
	afterFunc        hookFunc
}

type databaseChart struct {
	values  map[string]string
	repoUrl string
	name    string
	version string
}

var (
	Cockroach = databaseChart{
		repoUrl: "https://charts.cockroachdb.com/",
		name:    "cockroachdb",
		version: "11.0.1",
	}
	Postgres = databaseChart{
		repoUrl: "https://charts.bitnami.com/bitnami",
		name:    "postgresql",
		version: "12.6.4",
	}
)

func WithValues(chart databaseChart, values map[string]string) databaseChart {
	chart.values = values
	return chart
}

func Configure(
	t *testing.T,
	namespace string,
	dbChart databaseChart,
	zitadelValues []string,
	before, after hookFunc,
) *ConfigurationTest {
	chartPath, err := filepath.Abs("../../")
	require.NoError(t, err)
	dbRepoName := fmt.Sprintf("crdb-%s", strings.TrimPrefix(namespace, "zitadel-helm-"))
	kubeOptions := k8s.NewKubectlOptions("", "", namespace)
	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, kubeOptions)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &ConfigurationTest{
		Ctx:              context.Background(),
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
		afterFunc:        after,
	}
	cfg.SetT(t)
	return cfg
}
