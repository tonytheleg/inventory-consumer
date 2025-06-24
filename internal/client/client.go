package kessel

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/project-kessel/inventory-client-go/common"
	"github.com/project-kessel/inventory-client-go/v1beta1"
)

func New(c CompletedConfig, logger *log.Helper) (*v1beta1.InventoryClient, error) {

	logger.Info("Setting up Inventory API client")
	client, err := v1beta1.New(common.NewConfig(
		common.WithgRPCUrl(c.InventoryURL),
		common.WithTLSInsecure(c.Insecure),
		common.WithAuthEnabled(c.ClientId, c.ClientSecret, c.TokenEndpoint),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create Inventory API gRPC client: %v", err)
	}
	return client, nil
}
