package installation

func (s *ConfigurationTest) AfterTest(_, _ string) {
	if s.afterFunc == nil || s.T().Failed() {
		return
	}
	s.afterFunc(s)
}
