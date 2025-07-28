package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/project-kessel/inventory-consumer/consumer"
	kessel "github.com/project-kessel/inventory-consumer/internal/client"
	"github.com/project-kessel/inventory-consumer/internal/common"
	metricscollector "github.com/project-kessel/inventory-consumer/metrics"
	"github.com/spf13/cobra"
)

func startCommand(consumerOptions *consumer.Options, clientOptions *kessel.Options, loggerOptions common.LoggerOptions) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Starts the Inventory Resource Consumer",
		Long: `Starts the Inventory Resource Consumer in the specified environment,
subscribed to the provided topic`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, logger := common.InitLogger(common.GetLogLevel(), loggerOptions)
			logHelper := log.NewHelper(log.With(logger, "subsystem", "inventoryConsumer"))

			var client *kessel.KesselClient

			// configure consumer
			if errs = consumerOptions.Complete(); errs != nil {
				return fmt.Errorf("failed to setup consumer options: %v", errs)
			}
			if errs = consumerOptions.Validate(); errs != nil {
				return fmt.Errorf("consumer options validation error: %v", errs)
			}
			consumerConfig, errs := consumer.NewConfig(consumerOptions).Complete()
			if errs != nil {
				return fmt.Errorf("failed to setup consumer config: %v", errs)
			}

			// configure inventory client
			if errs = clientOptions.Complete(); errs != nil {
				return fmt.Errorf("failed to setup client options: %v", errs)
			}
			if errs = clientOptions.Validate(); errs != nil {
				return fmt.Errorf("client options validation error: %v", errs)
			}
			clientConfig, errs := kessel.NewConfig(clientOptions).Complete()
			if errs != nil {
				return fmt.Errorf("failed to setup client config: %v", errs)
			}
			client, err = kessel.New(clientConfig, log.NewHelper(log.With(logger, "subsystem", "client")))
			if err != nil {
				return fmt.Errorf("failed to instantiate client: %v", errs)
			}

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

			log.Info("starting metrics server on port :9000")
			go metricscollector.ServeMetrics()

			srvErrs := make(chan error)
			if consumerConfig.Enabled {
				go func() {
					srvErrs <- kic.Run(consumerOptions, consumerConfig, client, logHelper)
				}()
			} else {
				log.Info("Consumer disabled -- running in Standby mode")
			}
			select {
			case <-quit:
				shutdown(&kic, logHelper, fmt.Errorf("received signal \"quit\", shutting down"))
			case err := <-srvErrs:
				shutdown(&kic, logHelper, err)
			}
			return nil

		},
	}
	consumerOptions.AddFlags(startCmd.Flags(), "consumer")
	clientOptions.AddFlags(startCmd.Flags(), "client")
	return startCmd
}

func shutdown(cm *consumer.InventoryConsumer, logger *log.Helper, reason error) {
	log.Info(fmt.Sprintf("Consumer Shutdown: %s", reason))

	if cm != nil {
		defer func() {
			err := cm.Shutdown()
			if err != nil {
				if errors.Is(err, consumer.ErrClosed) {
					logger.Warn("error shutting down consumer, consumer already closed")
				} else {
					logger.Error(fmt.Sprintf("Error Gracefully Shutting Down Consumer: %v", err))
				}
			}
		}()
	}
}
