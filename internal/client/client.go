package kessel

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/project-kessel/kessel-sdk-go/kessel/errors"
	"github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
)

type ClientProvider interface {
	CreateOrUpdateResource(request *v1beta2.ReportResourceRequest) (*v1beta2.ReportResourceResponse, error)
	DeleteResource(request *v1beta2.DeleteResourceRequest) (*v1beta2.DeleteResourceResponse, error)
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
		client, err = v1beta2.NewInventoryGRPCClientBuilder().
			WithEndpoint(c.InventoryURL).
			WithOAuth2(c.ClientId, c.ClientSecret, c.TokenEndpoint).
			WithInsecure(c.Insecure).
			WithMaxReceiveMessageSize(8 * 1024 * 1024).
			WithMaxSendMessageSize(4 * 1024 * 1024).
			Build()
	} else {
		client, err = v1beta2.NewInventoryGRPCClientBuilder().
			WithEndpoint(c.InventoryURL).
			WithInsecure(c.Insecure).
			WithMaxReceiveMessageSize(8 * 1024 * 1024).
			WithMaxSendMessageSize(4 * 1024 * 1024).
			Build()
	}
	if err != nil {
		if errors.IsConnectionError(err) {
			return &KesselClient{}, fmt.Errorf("failed to establish connection: %w", err)
		} else if errors.IsTokenError(err) {
			return &KesselClient{}, fmt.Errorf("oauth2 token configuration failed: %w", err)
		} else {
			return &KesselClient{}, fmt.Errorf("failed to create Inventory API gRPC client: %w", err)
		}
	}
	return &KesselClient{
		InventoryClient: client,
		Enabled:         c.Enabled,
		AuthEnabled:     c.EnableOidcAuth,
	}, nil
}

func (k *KesselClient) CreateOrUpdateResource(request *v1beta2.ReportResourceRequest) (*v1beta2.ReportResourceResponse, error) {
	resp, err := k.ReportResource(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("failed to report resource: %w", err)
	}
	return resp, nil
}

func (k *KesselClient) DeleteResource(request *v1beta2.DeleteResourceRequest) (*v1beta2.DeleteResourceResponse, error) {
	resp, err := k.KesselInventoryServiceClient.DeleteResource(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("failed to delete resource: %w", err)
	}
	return resp, nil
}

func (k *KesselClient) IsEnabled() bool {
	return k.Enabled
}
