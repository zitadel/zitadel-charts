package acceptance

import (
	"fmt"
	"github.com/zitadel/zitadel-go/v2/pkg/client/management"
	"github.com/zitadel/zitadel-go/v2/pkg/client/middleware"
	"github.com/zitadel/zitadel-go/v2/pkg/client/zitadel"
)

func OpenGRPCConnection(cfg *ConfigurationTest, key []byte) *management.Client {
	t := cfg.T()
	conn, err := management.NewClient(
		cfg.APIBaseURL(),
		fmt.Sprintf("%s:%d", cfg.Domain, cfg.Port),
		[]string{zitadel.ScopeZitadelAPI()},
		zitadel.WithJWTProfileTokenSource(middleware.JWTProfileFromFileData(key)),
		zitadel.WithInsecure(),
	)
	if err != nil {
		t.Fatalf("couldn't create gRPC management client: %v", err)
	}
	return conn
}
