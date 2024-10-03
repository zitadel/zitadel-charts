package acceptance

import "github.com/gruntwork-io/terratest/modules/helm"

func (s *ConfigurationTest) BeforeTest(_, _ string) {
	if s.beforeFunc != nil {
		s.beforeFunc(s)
	}
	options := &helm.Options{
		KubectlOptions: s.KubeOptions,
		Version:        s.dbChart.version,
		SetValues:      s.dbChart.testValues,
		ExtraArgs:      map[string][]string{"install": {"--wait"}},
	}
	if s.dbChart.valuesFile != "" {
		options.ValuesFiles = []string{s.dbChart.valuesFile}
	}
	helm.Install(s.T(), options, s.dbChart.name+"/"+s.dbChart.name, s.dbRelease)
	if s.afterDBFunc != nil {
		s.afterDBFunc(s)
	}
}
