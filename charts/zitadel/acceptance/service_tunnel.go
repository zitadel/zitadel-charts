package acceptance

import (
	"context"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"net"
)

type CloseFunc func()

func (c CloseFunc) Close() {
	c()
}

// ServiceTunnel must be closed using the returned close function
func ServiceTunnel(cfg *ConfigurationTest) func() {
	serviceTunnel := k8s.NewTunnel(cfg.KubeOptions, k8s.ResourceTypeService, cfg.zitadelRelease, int(cfg.Port), 8080)
	awaitServicePortForward(cfg, serviceTunnel)
	return serviceTunnel.Close
}

func awaitServicePortForward(cfg *ConfigurationTest, tunnel *k8s.Tunnel) {
	t := cfg.T()
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%d", cfg.Port))
	if err != nil {
		t.Fatal(err)
	}
	Await(cfg.Ctx, t, nil, 600, func(ctx context.Context) error {
		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return err
		}
		if err := l.Close(); err != nil {
			panic(err)
		}
		return tunnel.ForwardPortE(cfg.T())
	})
}
