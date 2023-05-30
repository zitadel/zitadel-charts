package installation

import "github.com/gruntwork-io/terratest/modules/k8s"

type CloseFunc func()

func (c CloseFunc) Close() {
	c()
}

// ServiceTunnel must be closed using the returned close function
func ServiceTunnel(cfg *ConfigurationTest) func() {
	serviceTunnel := k8s.NewTunnel(cfg.KubeOptions, k8s.ResourceTypeService, cfg.zitadelRelease, 8080, 8080)
	serviceTunnel.ForwardPort(cfg.T())
	return serviceTunnel.Close
}
