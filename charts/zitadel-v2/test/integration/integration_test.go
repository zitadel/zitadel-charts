//go:build integration
// +build integration

package integration

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration(t *testing.T) {
	kubecontext := "kind-kind"

	chartPath, err := filepath.Abs("../../")
	require.NoError(t, err)

	namespace := createNamespaceName()
	kubeOptions := k8s.NewKubectlOptions(kubecontext, "", namespace)

	it := &integrationTest{
		chartPath:   chartPath,
		release:     "zitadel-test",
		namespace:   namespace,
		kubeOptions: kubeOptions,
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
	namespace := "zitadel-helm-"
	if !exist {
		namespace += strings.ToLower(random.UniqueId())
	} else {
		namespace += commitSHA
	}

	// max namespace length is 63 characters
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
	return truncateString(namespace, 63)
}
