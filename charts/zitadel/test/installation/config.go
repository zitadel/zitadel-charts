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
	zitadelValues    map[string]string
	crdbValues       map[string]string
	zitadelChartPath string
	zitadelRelease   string
	crdbRepoName     string
	crdbRepoURL      string
	crdbChart        string
	crdbVersion      string
	crdbRelease      string
	beforeFunc       hookFunc
	afterFunc        hookFunc
}

func Configure(t *testing.T, namespace string, zitadelValues map[string]string, before, after hookFunc) *ConfigurationTest {
	chartPath, err := filepath.Abs("../../")
	require.NoError(t, err)
	crdbRepoName := fmt.Sprintf("crdb-%s", strings.TrimPrefix(namespace, "zitadel-helm-"))
	kubeOptions := k8s.NewKubectlOptions("", "", namespace)
	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, kubeOptions)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &ConfigurationTest{
		Ctx:           context.Background(),
		log:           logger.New(logger.Terratest),
		KubeOptions:   kubeOptions,
		KubeClient:    clientset,
		zitadelValues: zitadelValues,
		crdbValues: map[string]string{
			"fullnameOverride": "crdb",
		},
		zitadelChartPath: chartPath,
		zitadelRelease:   "zitadel-test",
		crdbRepoURL:      "https://charts.cockroachdb.com/",
		crdbRepoName:     crdbRepoName,
		crdbChart:        fmt.Sprintf("%s/cockroachdb", crdbRepoName),
		crdbRelease:      "crdb",
		crdbVersion:      "11.0.1",
		beforeFunc:       before,
		afterFunc:        after,
	}
	cfg.SetT(t)
	return cfg
}
