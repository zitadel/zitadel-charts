package installation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes"
)

type beforeFunc func(ctx context.Context, namespace string, k8sClient *kubernetes.Clientset) error

type configurationTest struct {
	suite.Suite
	context          context.Context
	log              *logger.Logger
	kubeOptions      *k8s.KubectlOptions
	zitadelValues    map[string]string
	crdbValues       map[string]string
	zitadelChartPath string
	zitadelRelease   string
	crdbRepoName     string
	crdbRepoURL      string
	crdbChart        string
	crdbVersion      string
	crdbRelease      string
	beforeFunc       beforeFunc
}

func TestConfiguration(t *testing.T, before beforeFunc, zitadelValues map[string]string) {

	chartPath, err := filepath.Abs("../../")
	require.NoError(t, err)

	namespace := createNamespaceName()
	crdbRepoName := fmt.Sprintf("crdb-%s", strings.TrimPrefix(namespace, "zitadel-helm-"))
	kubeOptions := k8s.NewKubectlOptions("", "", namespace)

	it := &configurationTest{
		context:       context.Background(),
		log:           logger.New(logger.Terratest),
		kubeOptions:   kubeOptions,
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
		crdbVersion:      "10.0.0",
		beforeFunc:       before,
	}
	suite.Run(t, it)
}

func truncateString(str string, num int) string {
	shortenStr := str
	if len(str) > num {
		shortenStr = str[0:num]
	}
	return shortenStr
}

func createNamespaceName() string {
	// if triggered by a github action the environment variable is set
	// we use it to better identify the test
	commitSHA, exist := os.LookupEnv("GITHUB_SHA")
	namespace := "zitadel-helm-" + strings.ToLower(random.UniqueId())

	if exist {
		namespace += "-" + commitSHA
	}

	// max namespace length is 63 characters
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
	return truncateString(namespace, 63)
}
