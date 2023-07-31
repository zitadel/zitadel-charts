package installation

import "github.com/gruntwork-io/terratest/modules/helm"

func (s *ConfigurationTest) BeforeTest(_, _ string) {
	helm.AddRepo(s.T(), &helm.Options{}, s.dbRepoName, s.dbChart.repoUrl)
	options := &helm.Options{
		KubectlOptions: s.KubeOptions,
		Version:        s.dbChart.version,
		SetValues:      s.dbChart.testValues,
	}
	if s.dbChart.valuesFile != "" {
		options.ValuesFiles = []string{s.dbChart.valuesFile}
	}
	helm.Install(s.T(), options, s.dbRepoName+"/"+s.dbChart.name, s.dbRelease)
	if s.beforeFunc == nil || s.T().Failed() {
		return
	}
	s.beforeFunc(s)
}
