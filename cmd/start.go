package cmd

import (
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"
	"github.com/tonytheleg/inventory-consumer/common"
	"github.com/tonytheleg/inventory-consumer/consumer"
)

func startCommand(options *consumer.Options) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Starts the Inventory Resource Consumer",
		Long: `Starts the Inventory Resource Consumer in the specified environment,
subscribed to the provided topic`,
		Run: func(cmd *cobra.Command, args []string) {
			_, loggr := common.InitLogger("info", common.LoggerOptions{})

			if errs = options.Complete(); errs != nil {
				logger.Errorf("failed to complete options: %v", errs)
			}
			if errs = options.Validate(); errs != nil {
				logger.Errorf("failed to validate options: %v", errs)
			}
			if options.Enabled {
				consumerConfig, errs = consumer.NewConfig(options).Complete()
				if errs != nil {
					logger.Errorf("failed to setup consumer config: %v", errs)
				}
			}

			if options.Enabled {
				go func() {
					retries := 0
					for options.RetryOptions.ConsumerMaxRetries == -1 || retries < options.RetryOptions.ConsumerMaxRetries {
						// If the consumer cannot process a message, the consumer loop is restarted
						// This is to ensure we re-read the message and prevent it being dropped and moving to next message.
						// To re-read the current message, we have to recreate the consumer connection so that the earliest offset is used
						icrg, err = consumer.New(consumerConfig, nil, log.NewHelper(log.With(loggr, "subsystem", "ircg")))
						if err != nil {
							shutdown(&icrg, log.NewHelper(loggr), err)
						}
						err = icrg.Consume()
						if errors.Is(err, consumer.ErrClosed) {
							icrg.Logger.Errorf("consumer unable to process current message -- restarting consumer")
							retries++
							if options.RetryOptions.ConsumerMaxRetries == -1 || retries < options.RetryOptions.ConsumerMaxRetries {
								backoff := min(time.Duration(icrg.RetryOptions.BackoffFactor*retries*300)*time.Millisecond, time.Duration(options.RetryOptions.MaxBackoffSeconds)*time.Second)
								icrg.Logger.Errorf("retrying in %v", backoff)
								time.Sleep(backoff)
							}
							continue
						} else {
							icrg.Logger.Errorf("consumer unable to process messages: %v", err)
							shutdown(&icrg, log.NewHelper(loggr), err)
						}
					}
					shutdown(&icrg, log.NewHelper(loggr), err)
				}()
			}
		},
	}
	options.AddFlags(startCmd.Flags(), "consumer")
	return startCmd
}
