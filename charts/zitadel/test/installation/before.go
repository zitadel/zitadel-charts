package installation

import "github.com/gruntwork-io/terratest/modules/helm"

func (s *ConfigurationTest) BeforeTest(_, _ string) {
	helm.AddRepo(s.T(), &helm.Options{}, s.dbRepoName, s.dbChart.repoUrl)
	helm.Install(s.T(), &helm.Options{
		KubectlOptions: s.KubeOptions,
		SetValues:      s.dbChart.values,
		Version:        s.dbChart.version,
	}, s.dbRepoName+"/"+s.dbChart.name, s.dbRelease)
	if s.beforeFunc == nil || s.T().Failed() {
		return
	}
	s.beforeFunc(s)
}
