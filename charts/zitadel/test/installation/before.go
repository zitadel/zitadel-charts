package installation

func (s *ConfigurationTest) BeforeTest(_, _ string) {
	if s.beforeFunc == nil || s.T().Failed() {
		return
	}
	s.beforeFunc(s)
}
