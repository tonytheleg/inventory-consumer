package kessel

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
	"github.com/project-kessel/inventory-client-go/common"
	v1beta2 "github.com/project-kessel/inventory-client-go/v1beta2"
	"google.golang.org/grpc"
)

type ClientProvider interface {
	CreateOrUpdateResource(request *kesselv2.ReportResourceRequest) (*kesselv2.ReportResourceResponse, error)
	DeleteResource(request *kesselv2.DeleteResourceRequest) (*kesselv2.DeleteResourceResponse, error)
	IsEnabled() bool
}

type KesselClient struct {
	*v1beta2.InventoryClient
	Enabled     bool
	AuthEnabled bool
}

func New(c CompletedConfig, logger *log.Helper) (*KesselClient, error) {
	logger.Info("Setting up Inventory API client")
	var client *v1beta2.InventoryClient
	var err error

	if !c.Enabled {
		logger.Info("ClientProvider enabled: ", c.Enabled)
		return &KesselClient{Enabled: false}, nil
	}

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
		return &KesselClient{}, fmt.Errorf("failed to create Inventory API gRPC client: %w", err)
	}

	return &KesselClient{
		InventoryClient: client,
		Enabled:         c.Enabled,
		AuthEnabled:     c.EnableOidcAuth,
	}, nil
}

func (k *KesselClient) CreateOrUpdateResource(request *kesselv2.ReportResourceRequest) (*kesselv2.ReportResourceResponse, error) {
	var opts []grpc.CallOption
	var err error

	if k.AuthEnabled {
		opts, err = k.GetTokenCallOption()
		if err != nil {
			return nil, fmt.Errorf("failed to get token option: %w", err)
		}
	}

	resp, err := k.KesselInventoryService.ReportResource(context.Background(), request, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to report resource: %w", err)
	}
	return resp, nil
}

func (k *KesselClient) DeleteResource(request *kesselv2.DeleteResourceRequest) (*kesselv2.DeleteResourceResponse, error) {
	var opts []grpc.CallOption
	var err error

	if k.AuthEnabled {
		opts, err = k.GetTokenCallOption()
		if err != nil {
			return nil, fmt.Errorf("failed to get token option: %w", err)
		}
	}

	resp, err := k.KesselInventoryService.DeleteResource(context.Background(), request, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to delete resource: %w", err)
	}
	return resp, nil
}

func (k *KesselClient) IsEnabled() bool {
	return k.Enabled
}
