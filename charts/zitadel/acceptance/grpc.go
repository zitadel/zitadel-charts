package acceptance

import (
	oidc_client "github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/management"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
	"strconv"
)

func OpenGRPCConnection(cfg *ConfigurationTest, key []byte) (management.ManagementServiceClient, error) {
	var clientOptions []client.Option
	if key != nil {
		keyFile, err := oidc_client.ConfigFromKeyFileData(key)
		if err != nil {
			return nil, err
		}
		clientOptions = append(clientOptions, client.WithAuth(client.JWTAuthentication(keyFile, client.ScopeZitadelAPI())))
	}
	zitadelOptions := []zitadel.Option{
		zitadel.WithPort(cfg.Port),
		zitadel.WithInsecureSkipVerifyTLS(),
	}
	if cfg.Scheme != "https" {
		zitadelOptions = append(zitadelOptions, zitadel.WithInsecure(strconv.Itoa(int(cfg.Port))))
	}
	c, err := client.New(cfg.Ctx, zitadel.New(cfg.Domain, zitadelOptions...), clientOptions...)
	if err != nil {
		return nil, err
	}
	return c.ManagementService(), nil
}
