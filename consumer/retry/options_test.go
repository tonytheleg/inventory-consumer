package retry

import (
	"testing"

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
			ConsumerMaxRetries:  2,
			OperationMaxRetries: 3,
			BackoffFactor:       5,
			MaxBackoffSeconds:   30,
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
	prefix := "consumer.retry-options"
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	test.options.AddFlags(fs, prefix)

	common.AllOptionsHaveFlags(t, prefix, fs, *test.options, nil)
}
