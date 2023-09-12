package acceptance

func (s *ConfigurationTest) AfterTest(_, _ string) {
	if s.afterZITADELFunc == nil || s.T().Failed() {
		return
	}
	s.afterZITADELFunc(s)
}
