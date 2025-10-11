package acceptance_test

import (
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
)

// BeforeTest runs before each individual test case.
// It executes test-specific setup hooks, installs the database chart
// in the test's namespace, and runs post-database setup hooks.
// The helm repository is already cached from SetupSuite, so only
// the chart package is downloaded on first use and then cached.
func (s *ConfigurationTest) BeforeTest(_, _ string) {
	if s.beforeFunc != nil {
		s.beforeFunc(s)
	}

	options := &helm.Options{
		KubectlOptions: s.KubeOptions,
		Version:        s.dbChart.version,
		SetValues:      s.dbChart.testValues,
		ExtraArgs:      map[string][]string{"install": {"--wait", "--timeout", "10m", "--hide-notes"}},
	}
	if s.dbChart.valuesFile != "" {
		options.ValuesFiles = []string{s.dbChart.valuesFile}
	}

	helm.Install(s.T(), options, s.dbRepoName+"/"+s.dbChart.name, s.dbRelease)
	s.T().Log("Waiting 30 seconds for PostgreSQL to become fully ready...")
	time.Sleep(1 * time.Minute)

	if s.afterDBFunc != nil {
		s.afterDBFunc(s)
	}
}
