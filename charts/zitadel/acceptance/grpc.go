package acceptance

import (
	"fmt"
	"github.com/zitadel/zitadel-go/v2/pkg/client/management"
	"github.com/zitadel/zitadel-go/v2/pkg/client/middleware"
	"github.com/zitadel/zitadel-go/v2/pkg/client/zitadel"
)

func OpenGRPCConnection(cfg *ConfigurationTest, key []byte) (*management.Client, error) {
	conn, err := management.NewClient(
		cfg.APIBaseURL(),
		fmt.Sprintf("%s:%d", cfg.Domain, cfg.Port),
		[]string{zitadel.ScopeZitadelAPI()},
		zitadel.WithJWTProfileTokenSource(middleware.JWTProfileFromFileData(key)),
		zitadel.WithInsecure(),
	)
	return conn, err
}
