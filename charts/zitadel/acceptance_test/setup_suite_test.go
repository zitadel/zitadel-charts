package acceptance_test

import (
	"context"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
)

// SetupSuite runs once before all tests in the suite.
// It adds the database helm repository to cache the repository index,
// preventing repeated downloads across test executions.
func (s *ConfigurationTest) SetupSuite() {
	options := &helm.Options{
		KubectlOptions: s.KubeOptions,
	}
	Awaitf(context.Background(), s.T(), 1*time.Minute, func(ctx context.Context) error {
		err := helm.AddRepoE(s.T(), options, s.dbRepoName, s.dbChart.repoUrl)
		if err != nil {
			s.T().Log(err)
		}
		return err
	}, "adding helm repo %s with URL %s failed for a minute", s.dbRepoName, s.dbChart.repoUrl)
}
