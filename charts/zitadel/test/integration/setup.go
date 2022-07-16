package integration

import (
	"context"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/suite"
)

type integrationTest struct {
	suite.Suite
	context     context.Context
	log         *logger.Logger
	chartPath   string
	release     string
	namespace   string
	kubeOptions *k8s.KubectlOptions
}
