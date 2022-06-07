package integration

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/suite"
)

type integrationTest struct {
	suite.Suite
	chartPath   string
	release     string
	namespace   string
	kubeOptions *k8s.KubectlOptions
}
