package acceptance_test

import (
	"context"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
)

func (s *ConfigurationTest) BeforeTest(_, _ string) {
	if s.beforeFunc != nil {
		s.beforeFunc(s)
	}
	options := &helm.Options{
		KubectlOptions: s.KubeOptions,
		Version:        s.dbChart.version,
		SetValues:      s.dbChart.testValues,
		ExtraArgs:      map[string][]string{"install": {"--wait", "--timeout", "10m"}},
	}
	Awaitf(context.Background(), s.T(), 1*time.Minute, func(ctx context.Context) error {
		err := helm.AddRepoE(s.T(), options, s.dbRepoName, s.dbChart.repoUrl)
		if err != nil {
			s.T().Log(err)
		}
		return err
	}, "adding helm repo %s with URL %s failed for a minute", s.dbRepoName, s.dbChart.repoUrl)
	if s.dbChart.valuesFile != "" {
		options.ValuesFiles = []string{s.dbChart.valuesFile}
	}
	helm.Install(s.T(), options, s.dbRepoName+"/"+s.dbChart.name, s.dbRelease)
	s.T().Log("Waiting 30 seconds for PostgreSQL to become fully ready...")
	time.Sleep(30 * time.Second)

	if s.afterDBFunc != nil {
		s.afterDBFunc(s)
	}
}
