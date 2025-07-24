package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/authzed/grpcutil"
	"github.com/go-kratos/kratos/v2/log"
	kesselv1 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1"
	kessel "github.com/project-kessel/inventory-consumer/internal/client"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func readyzCommand(clientOptions *kessel.Options) *cobra.Command {
	readyzCmd := &cobra.Command{
		Use:   "readyz",
		Short: "Check if the Inventory API service is ready",
		Long: `Check if the Inventory API service is ready by making a gRPC request
to the kessel.inventory.v1.KesselInventoryHealthService.GetLivez endpoint.
The InventoryURL from the client configuration is used as the gRPC endpoint.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate client configuration
			if errs := clientOptions.Complete(); errs != nil {
				return fmt.Errorf("failed to setup client options")
			}
			if errs := clientOptions.Validate(); errs != nil {
				return fmt.Errorf("client options validation error")
			}

			// Check if client is enabled
			if !clientOptions.Enabled {
				log.Info("inventory service is not enabled")
				return nil
			}

			// Check if InventoryURL is configured
			if clientOptions.InventoryURL == "" {
				return fmt.Errorf("inventory URL not configured")
			}

			fmt.Printf("Checking inventory service readiness at: %s\n", clientOptions.InventoryURL)

			// Set up gRPC connection options
			var opts []grpc.DialOption

			if clientOptions.Insecure {
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
			} else {
				tlsConfig, _ := grpcutil.WithSystemCerts(grpcutil.VerifyCA)
				opts = append(opts, tlsConfig)
			}

			// Create gRPC connection with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			conn, err := grpc.NewClient(clientOptions.InventoryURL, opts...)
			if err != nil {
				return fmt.Errorf("failed to connect to inventory service: %v", err)
			}
			defer conn.Close() //nolint:errcheck

			// Create health service client
			healthClient := kesselv1.NewKesselInventoryHealthServiceClient(conn)

			// Make gRPC call to GetLivez
			req := &kesselv1.GetLivezRequest{}
			resp, err := healthClient.GetLivez(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to check inventory service health: %v", err)
			}

			// Check if the response indicates the service is healthy
			// Typically, a successful gRPC call means the service is ready
			// For health checks, we expect HTTP-like status codes where 2xx indicates success
			code := resp.GetCode()
			if code < 200 || code >= 300 {
				return fmt.Errorf("inventory service not healthy, status: %s, code: %d", resp.GetStatus(), resp.GetCode())

			}
			// Return success
			fmt.Printf("Inventory service is ready! Status: %s\n", resp.GetStatus())
			return nil
		},
	}

	// Add client flags to the readyz command
	clientOptions.AddFlags(readyzCmd.Flags(), "client")

	return readyzCmd
}
