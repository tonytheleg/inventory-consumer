package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"
	"github.com/tonytheleg/inventory-consumer/consumer"
	"github.com/tonytheleg/inventory-consumer/internal/common"
	"github.com/tonytheleg/inventory-consumer/internal/storage"
)

func startCommand(consumerOptions *consumer.Options, storageOptions *storage.Options, loggerOptions common.LoggerOptions) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Starts the Inventory Resource Consumer",
		Long: `Starts the Inventory Resource Consumer in the specified environment,
subscribed to the provided topic`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, logger := common.InitLogger(common.GetLogLevel(), loggerOptions)
			logHelper := log.NewHelper(log.With(logger, "subsystem", "inventoryConsumer"))

			// configure storage
			if errs := storageOptions.Complete(); errs != nil {
				return fmt.Errorf("failed to setup storage options: %v", errs)
			}
			if errs := storageOptions.Validate(); errs != nil {
				return fmt.Errorf("storage options validation error: %v", errs)
			}
			storageConfig := storage.NewConfig(storageOptions).Complete()
			db, err := storage.New(storageConfig, log.NewHelper(log.With(logger, "subsystem", "storage")))
			if err != nil {
				return err
			}

			// configure consumer
			if errs = consumerOptions.Complete(); errs != nil {
				return fmt.Errorf("failed to setup consumer options: %v", errs)
			}
			if errs = consumerOptions.Validate(); errs != nil {
				return fmt.Errorf("consumer options validation error: %v", errs)
			}
			consumerConfig, errs = consumer.NewConfig(consumerOptions).Complete()
			if errs != nil {
				return fmt.Errorf("failed to setup consumer config: %v", errs)
			}

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

			srvErrs := make(chan error)
			go func() {
				srvErrs <- icrg.Run(consumerOptions, consumerConfig, db, logHelper)
			}()
			select {
			case <-quit:
				shutdown(&icrg, logHelper, fmt.Errorf("received signal \"quit\", shutting down"))
			case err := <-srvErrs:
				shutdown(&icrg, logHelper, err)
			}
			return nil

		},
	}
	consumerOptions.AddFlags(startCmd.Flags(), "consumer")
	storageOptions.AddFlags(startCmd.Flags(), "storage")
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
