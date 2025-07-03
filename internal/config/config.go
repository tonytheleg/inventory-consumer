package config

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/project-kessel/inventory-consumer/consumer"
	kessel "github.com/project-kessel/inventory-consumer/internal/client"
	"github.com/project-kessel/inventory-consumer/internal/common"
	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
)

// OptionsConfig contains the settings for each configuration option
type OptionsConfig struct {
	Consumer *consumer.Options
	Client   *kessel.Options
}

// NewOptionsConfig returns a new OptionsConfig with default options set
func NewOptionsConfig() *OptionsConfig {
	return &OptionsConfig{
		consumer.NewOptions(),
		kessel.NewOptions(),
	}
}

// LogConfigurationInfo outputs connection details to logs when in debug for testing (no secret data is output)
func LogConfigurationInfo(options *OptionsConfig) {
	log.Debugf("Consumer Configuration: Bootstrap Server: %s, Topics: %s, Consumer Max Retries: %d, Operation Max Retries: %d, Backoff Factor: %d, Max Backoff Seconds: %d",
		options.Consumer.BootstrapServers,
		options.Consumer.Topics,
		options.Consumer.RetryOptions.ConsumerMaxRetries,
		options.Consumer.RetryOptions.OperationMaxRetries,
		options.Consumer.RetryOptions.BackoffFactor,
		options.Consumer.RetryOptions.MaxBackoffSeconds,
	)

	log.Debugf("Consumer Auth Settings: Enabled: %v, Security Protocol: %s, Mechanism: %s, Username: %s",
		options.Consumer.AuthOptions.Enabled,
		options.Consumer.AuthOptions.SecurityProtocol,
		options.Consumer.AuthOptions.SASLMechanism,
		options.Consumer.AuthOptions.SASLUsername)

	if options.Client.Enabled {
		log.Debugf("Client Configuration: URL: %s, Insecure?: %t, Token Endpoint?: %s",
			options.Client.InventoryURL,
			options.Client.Insecure,
			options.Client.TokenEndpoint,
		)
	}
}

// InjectClowdAppConfig updates service options based on values in the ClowdApp AppConfig
func (o *OptionsConfig) InjectClowdAppConfig(appconfig *clowder.AppConfig) error {
	// check for consumer config
	if !common.IsNil(appconfig.Kafka) {
		o.ConfigureConsumer(appconfig)
	}
	return nil
}

// ConfigureConsumer updates Consumer settings based on ClowdApp AppConfig
func (o *OptionsConfig) ConfigureConsumer(appconfig *clowder.AppConfig) {
	var brokers []string
	for _, broker := range appconfig.Kafka.Brokers {
		brokers = append(brokers, fmt.Sprintf("%s:%d", broker.Hostname, *broker.Port))
	}
	o.Consumer.BootstrapServers = brokers

	if len(appconfig.Kafka.Brokers) > 0 && appconfig.Kafka.Brokers[0].SecurityProtocol != nil {
		o.Consumer.AuthOptions.SecurityProtocol = *appconfig.Kafka.Brokers[0].SecurityProtocol

		if appconfig.Kafka.Brokers[0].Sasl != nil {
			o.Consumer.AuthOptions.SASLMechanism = *appconfig.Kafka.Brokers[0].Sasl.SaslMechanism
			o.Consumer.AuthOptions.SASLUsername = *appconfig.Kafka.Brokers[0].Sasl.Username
			o.Consumer.AuthOptions.SASLPassword = *appconfig.Kafka.Brokers[0].Sasl.Password
		}
	}
}
