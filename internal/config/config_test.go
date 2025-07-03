package config

import (
	"testing"

	. "github.com/project-kessel/inventory-api/cmd/common"
	"github.com/project-kessel/inventory-consumer/consumer"
	"github.com/project-kessel/inventory-consumer/consumer/auth"
	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestConfigureConsumer(t *testing.T) {
	tests := []struct {
		name        string
		appconfig   *clowder.AppConfig
		options     *OptionsConfig
		authEnabled bool
		expected    []string
	}{
		{
			name: "ensure boostrap server is set properly when only one is provided - no auth settings",
			appconfig: &clowder.AppConfig{
				Kafka: &clowder.KafkaConfig{
					Brokers: []clowder.BrokerConfig{
						{
							Hostname: "test-kafka-server",
							Port:     ToPointer(9092),
						},
					},
				},
			},
			options:     NewOptionsConfig(),
			authEnabled: false,
			expected:    []string{"test-kafka-server:9092"},
		},
		{
			name: "ensure boostrap server is set properly when multiple are provided - no auth settings",
			appconfig: &clowder.AppConfig{
				Kafka: &clowder.KafkaConfig{
					Brokers: []clowder.BrokerConfig{
						{
							Hostname: "test-kafka-server-01",
							Port:     ToPointer(9092),
						},
						{
							Hostname: "test-kafka-server-02",
							Port:     ToPointer(9092),
						},
						{
							Hostname: "test-kafka-server-03",
							Port:     ToPointer(9092),
						},
					},
				},
			},
			options:     NewOptionsConfig(),
			authEnabled: false,
			expected:    []string{"test-kafka-server-01:9092", "test-kafka-server-02:9092", "test-kafka-server-03:9092"},
		},
		{
			name: "ensure sasl settings are configured when present",
			appconfig: &clowder.AppConfig{
				Kafka: &clowder.KafkaConfig{
					Brokers: []clowder.BrokerConfig{
						{
							Hostname:         "test-kafka-server-01",
							Port:             ToPointer(9092),
							SecurityProtocol: ToPointer("SASL_SSL"),
							Sasl: &clowder.KafkaSASLConfig{
								Password:         ToPointer("test-password"),
								SaslMechanism:    ToPointer("SCRAM-SHA-512"),
								SecurityProtocol: ToPointer("SASL_SSL"),
								Username:         ToPointer("test-user"),
							},
						},
					},
				},
			},
			options:     NewOptionsConfig(),
			authEnabled: true,
			expected:    []string{"test-kafka-server-01:9092"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.options.Consumer.AuthOptions.Enabled = test.authEnabled
			test.options.ConfigureConsumer(test.appconfig)
			assert.Equal(t, test.expected, test.options.Consumer.BootstrapServers)
			if test.authEnabled {
				assert.Equal(t, test.options.Consumer.AuthOptions.SecurityProtocol, *test.appconfig.Kafka.Brokers[0].SecurityProtocol)
				assert.Equal(t, test.options.Consumer.AuthOptions.SASLMechanism, *test.appconfig.Kafka.Brokers[0].Sasl.SaslMechanism)
				assert.Equal(t, test.options.Consumer.AuthOptions.SASLUsername, *test.appconfig.Kafka.Brokers[0].Sasl.Username)
				assert.Equal(t, test.options.Consumer.AuthOptions.SASLPassword, *test.appconfig.Kafka.Brokers[0].Sasl.Password)
			} else {
				assert.Equal(t, test.options.Consumer.AuthOptions.SecurityProtocol, "")
				assert.Equal(t, test.options.Consumer.AuthOptions.SASLMechanism, "")
				assert.Equal(t, test.options.Consumer.AuthOptions.SASLUsername, "")
				assert.Equal(t, test.options.Consumer.AuthOptions.SASLPassword, "")
			}
		})
	}
}

func TestInjectClowdAppConfig(t *testing.T) {
	consumerTest := struct {
		name      string
		appconfig *clowder.AppConfig
		options   *OptionsConfig
		expected  *OptionsConfig
	}{
		name: "Consumer is configured and injected with no auth settings, authz and storage are ignored",
		appconfig: &clowder.AppConfig{
			Endpoints: []clowder.DependencyEndpoint{},
			Kafka: &clowder.KafkaConfig{
				Brokers: []clowder.BrokerConfig{
					{
						Hostname: "test-kafka-server",
						Port:     ToPointer(9092),
					},
				},
			},
		},
		options: NewOptionsConfig(),
		expected: &OptionsConfig{
			Consumer: &consumer.Options{
				BootstrapServers: []string{"test-kafka-server:9092"},
			},
		},
	}
	t.Run(consumerTest.name, func(t *testing.T) {
		err := consumerTest.options.InjectClowdAppConfig(consumerTest.appconfig)
		assert.NoError(t, err)
		assert.Equal(t, consumerTest.expected.Consumer.BootstrapServers, consumerTest.options.Consumer.BootstrapServers)
	})

	consumerAuthTest := struct {
		name      string
		appconfig *clowder.AppConfig
		options   *OptionsConfig
		expected  *OptionsConfig
	}{
		name: "Consumer is configured and injected with auth settings, authz and storage are ignored",
		appconfig: &clowder.AppConfig{
			Endpoints: []clowder.DependencyEndpoint{},
			Kafka: &clowder.KafkaConfig{
				Brokers: []clowder.BrokerConfig{
					{
						Hostname:         "test-kafka-server-01",
						Port:             ToPointer(9092),
						SecurityProtocol: ToPointer("SASL_SSL"),
						Sasl: &clowder.KafkaSASLConfig{
							Password:         ToPointer("test-password"),
							SaslMechanism:    ToPointer("SCRAM-SHA-512"),
							SecurityProtocol: ToPointer("SASL_SSL"),
							Username:         ToPointer("test-user"),
						},
					},
				},
			},
		},
		options: NewOptionsConfig(),
		expected: &OptionsConfig{
			Consumer: &consumer.Options{
				BootstrapServers: []string{"test-kafka-server-01:9092"},
				AuthOptions: &auth.Options{
					SecurityProtocol: "SASL_SSL",
					SASLMechanism:    "SCRAM-SHA-512",
					SASLUsername:     "test-user",
					SASLPassword:     "test-password",
				},
			},
		},
	}
	t.Run(consumerAuthTest.name, func(t *testing.T) {
		err := consumerAuthTest.options.InjectClowdAppConfig(consumerAuthTest.appconfig)
		assert.NoError(t, err)
		assert.Equal(t, consumerAuthTest.expected.Consumer.BootstrapServers, consumerAuthTest.options.Consumer.BootstrapServers)
		assert.Equal(t, consumerAuthTest.expected.Consumer.AuthOptions.SecurityProtocol, consumerAuthTest.options.Consumer.AuthOptions.SecurityProtocol)
		assert.Equal(t, consumerAuthTest.expected.Consumer.AuthOptions.SASLMechanism, consumerAuthTest.options.Consumer.AuthOptions.SASLMechanism)
		assert.Equal(t, consumerAuthTest.expected.Consumer.AuthOptions.SASLUsername, consumerAuthTest.options.Consumer.AuthOptions.SASLUsername)
		assert.Equal(t, consumerAuthTest.expected.Consumer.AuthOptions.SASLPassword, consumerAuthTest.options.Consumer.AuthOptions.SASLPassword)
	})

	noConfigTest := struct {
		name      string
		appconfig *clowder.AppConfig
		options   *OptionsConfig
		expected  *OptionsConfig
	}{
		name:      "No values found in AppConfig -- Clowder changes nothing",
		appconfig: &clowder.AppConfig{},
		options:   NewOptionsConfig(),
		expected:  NewOptionsConfig(),
	}

	t.Run(noConfigTest.name, func(t *testing.T) {
		err := noConfigTest.options.InjectClowdAppConfig(noConfigTest.appconfig)
		assert.NoError(t, err)
		assert.Equal(t, noConfigTest.expected.Consumer, noConfigTest.options.Consumer)
	})
}
