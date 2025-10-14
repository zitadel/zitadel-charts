package acceptance_test

import (
	"context"
	"crypto/tls"
	"net/url"

	oidcclient "github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/management"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func OpenGRPCConnection(ctx context.Context, cfg *ConfigurationTest, key []byte) (management.ManagementServiceClient, error) {
	clientOptions := []client.Option{
		client.WithGRPCDialOptions(grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}))),
	}
	if key != nil {
		//goland:noinspection GoDeprecation
		keyFile, err := oidcclient.ConfigFromKeyFileData(key)
		if err != nil {
			return nil, err
		}
		clientOptions = append(clientOptions, client.WithAuth(client.JWTAuthentication(keyFile, client.ScopeZitadelAPI())))
	}
	apiBaseUrl, err := url.Parse(cfg.ApiBaseUrl)
	if err != nil {
		return nil, err
	}
	c, err := client.New(ctx, zitadel.New(apiBaseUrl.Hostname(), zitadel.WithInsecureSkipVerifyTLS()), clientOptions...)
	if err != nil {
		return nil, err
	}
	return c.ManagementService(), nil
}
