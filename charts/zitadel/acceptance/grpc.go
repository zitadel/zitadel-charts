package acceptance

import (
	client2 "github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/management"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
	"strconv"
)

func OpenGRPCConnection(cfg *ConfigurationTest, key []byte) (management.ManagementServiceClient, error) {
	var options []client.Option
	if key != nil {
		keyFile, err := client2.ConfigFromKeyFileData(key)
		if err != nil {
			return nil, err
		}
		options = append(options, client.WithAuth(client.JWTAuthentication(keyFile, client.ScopeZitadelAPI())))
	}
	c, err := client.New(cfg.Ctx, zitadel.New(cfg.Domain, zitadel.WithInsecure(strconv.Itoa(int(cfg.Port)))), options...)
	if err != nil {
		return nil, err
	}
	return c.ManagementService(), nil
}
