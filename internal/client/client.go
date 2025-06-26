package kessel

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
	"github.com/project-kessel/inventory-client-go/common"
	v1beta2 "github.com/project-kessel/inventory-client-go/v1beta2"
)

func New(c CompletedConfig, logger *log.Helper) (*v1beta2.InventoryClient, error) {
	logger.Info("Setting up Inventory API client")
	var client *v1beta2.InventoryClient
	var err error

	if c.EnableOidcAuth {
		client, err = v1beta2.New(common.NewConfig(
			common.WithgRPCUrl(c.InventoryURL),
			common.WithTLSInsecure(c.Insecure),
			common.WithAuthEnabled(c.ClientId, c.ClientSecret, c.TokenEndpoint),
		))
	} else {
		client, err = v1beta2.New(common.NewConfig(
			common.WithgRPCUrl(c.InventoryURL),
			common.WithTLSInsecure(c.Insecure),
		))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Inventory API gRPC client: %v", err)
	}
	return client, nil
}

func CreateOrUpdateResource(client *v1beta2.InventoryClient, request *kesselv2.ReportResourceRequest) (*kesselv2.ReportResourceResponse, error) {
	resp, err := client.KesselInventoryService.ReportResource(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func DeleteResource(client *v1beta2.InventoryClient, request *kesselv2.DeleteResourceRequest) (*kesselv2.DeleteResourceResponse, error) {
	resp, err := client.KesselInventoryService.DeleteResource(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
