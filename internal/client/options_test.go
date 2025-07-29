package kessel

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
			Enabled:        true,
			Insecure:       true,
			EnableOidcAuth: false,
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
	prefix := "client"
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	test.options.AddFlags(fs, prefix)

	// the below logic ensures that every possible option defined in the Options type
	// has a defined flag for that option
	common.AllOptionsHaveFlags(t, prefix, fs, *test.options, nil)
}

func TestOptions_Validate(t *testing.T) {
	tests := []struct {
		name        string
		options     *Options
		expectError bool
	}{
		{
			name: "inventory url is empty and the client is enabled",
			options: &Options{
				Enabled:      true,
				InventoryURL: "",
			},
			expectError: true,
		},
		{
			name: "inventory url is set and the client is enabled",
			options: &Options{
				Enabled:      true,
				InventoryURL: "inventory-api:9000",
			},
			expectError: false,
		},
		{
			name: "inventory url is empty and the client is disabled",
			options: &Options{
				Enabled:      false,
				InventoryURL: "",
			},
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
