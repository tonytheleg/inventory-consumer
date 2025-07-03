package consumer

import (
	"testing"

	"github.com/project-kessel/inventory-consumer/consumer/auth"
	"github.com/project-kessel/inventory-consumer/consumer/retry"
	"github.com/project-kessel/inventory-consumer/internal/common"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestNewOptions(t *testing.T) {
	test := struct {
		options         *Options
		expectedOptions *Options
	}{
		options: NewOptions(),
		expectedOptions: &Options{
			ConsumerGroupID:    "kic",
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
			name: "bootstrap servers and topic are both set",
			options: &Options{
				BootstrapServers: []string{
					"test-server:9092",
				},
				Topics: []string{"test-topic"},
			},
			expectError: false,
		},
		{
			name: "bootstrap servers is empty and topic is set",
			options: &Options{
				BootstrapServers: []string{},
				Topics:           []string{"test-topic"},
			},
			expectError: true,
		},
		{
			name: "bootstrap servers is empty and topic is empty",
			options: &Options{
				BootstrapServers: []string{},
				Topics:           []string{},
			},
			expectError: true,
		},
		{
			name: "bootstrap servers is set and topic is empty",
			options: &Options{
				BootstrapServers: []string{
					"test-server:9092",
				},
				Topics: []string{},
			},
			expectError: true,
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
