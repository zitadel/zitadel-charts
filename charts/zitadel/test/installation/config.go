package installation

import (
	"context"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type beforeFunc func(ctx context.Context, namespace string, k8sClient *kubernetes.Clientset) error

type configurationTest struct {
	suite.Suite
	context    context.Context
	log        *logger.Logger
	chartPath  string
	release    string
	namespace  string
	options    *helm.Options
	beforeFunc beforeFunc
}

func TestConfiguration(t *testing.T, before beforeFunc, values map[string]string) {

	chartPath, err := filepath.Abs("../../")
	require.NoError(t, err)

	namespace := createNamespaceName()
	kubeOptions := k8s.NewKubectlOptions("", "", namespace)

	it := &configurationTest{
		context:   context.Background(),
		log:       logger.New(logger.Terratest),
		chartPath: chartPath,
		release:   "zitadel-test",
		namespace: namespace,
		options: &helm.Options{
			KubectlOptions: kubeOptions,
			SetValues:      values,
		},
		beforeFunc: before,
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
