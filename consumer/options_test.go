package consumer

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/tonytheleg/inventory-consumer/consumer/auth"
	"github.com/tonytheleg/inventory-consumer/consumer/retry"
	"github.com/tonytheleg/inventory-consumer/internal/common"
)

func TestNewOptions(t *testing.T) {
	test := struct {
		options         *Options
		expectedOptions *Options
	}{
		options: NewOptions(),
		expectedOptions: &Options{
			Enabled:            true,
			ConsumerGroupID:    "ircg",
			Topic:              "outbox.event.kessel.tuples",
			SessionTimeout:     "45000",
			HeartbeatInterval:  "3000",
			MaxPollInterval:    "300000",
			EnableAutoCommit:   "false",
			AutoOffsetReset:    "earliest",
			StatisticsInterval: "60000",
			Debug:              "",
			AuthOptions:        auth.NewOptions(),
			RetryOptions:       retry.NewOptions(),
		},
	}
	assert.Equal(t, test.expectedOptions, NewOptions())
}

func TestOptions_AddFlags(t *testing.T) {
	test := struct {
		options *Options
	}{
		options: NewOptions(),
	}
	prefix := "consumer"
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	test.options.AddFlags(fs, prefix)

	// the below logic ensures that every possible option defined in the Options type
	// has a defined flag for that option; auth and retry-options are skipped in favor of testing
	// in their own packages
	common.AllOptionsHaveFlags(t, prefix, fs, *test.options, []string{"auth", "retry-options"})
}

func TestOptions_Validate(t *testing.T) {
	tests := []struct {
		name        string
		options     *Options
		expectError bool
	}{
		{
			name: "bootstrap servers is set and consumer is enabled",
			options: &Options{
				Enabled: true,
				BootstrapServers: []string{
					"test-server:9092",
				}},
			expectError: false,
		},
		{
			name: "bootstrap servers is empty and consumer is enabled",
			options: &Options{
				Enabled:          true,
				BootstrapServers: []string{},
			},
			expectError: true,
		},
		{
			name: "bootstrap servers is empty and consumer is disabled",
			options: &Options{
				Enabled:          false,
				BootstrapServers: []string{},
			},
			expectError: false,
		},
		{
			name: "bootstrap servers is set but consumer is disabled",
			options: &Options{
				Enabled: false,
				BootstrapServers: []string{
					"test-server:9092",
				}},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errs := test.options.Validate()
			if test.expectError {
				assert.NotNil(t, errs)
			} else {
				assert.Nil(t, errs)
			}
		})
	}
}
