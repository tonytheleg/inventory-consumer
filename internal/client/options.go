package kessel

import (
	"fmt"

	"github.com/spf13/pflag"
)

type Options struct {
	Enabled        bool   `mapstructure:"enabled"`
	InventoryURL   string `mapstructure:"url"`
	Insecure       bool   `mapstructure:"insecure-client"`
	EnableOidcAuth bool   `mapstructure:"enable-oidc-auth"`
	ClientId       string `mapstructure:"client-id"`
	ClientSecret   string `mapstructure:"client-secret"`
	TokenEndpoint  string `mapstructure:"sso-token-endpoint"`
}

func NewOptions() *Options {
	return &Options{
		Enabled:        true,
		Insecure:       true,
		EnableOidcAuth: false,
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}
	fs.BoolVar(&o.Enabled, prefix+"enabled", o.Enabled, "enable the kessel inventory grpc client")
	fs.StringVar(&o.InventoryURL, prefix+"url", o.InventoryURL, "HTTP endpoint of the kessel inventory service.")
	fs.StringVar(&o.ClientId, prefix+"sa-client-id", o.ClientId, "service account client id")
	fs.StringVar(&o.ClientSecret, prefix+"sa-client-secret", o.ClientSecret, "service account secret")
	fs.StringVar(&o.TokenEndpoint, prefix+"sso-token-endpoint", o.TokenEndpoint, "sso token endpoint for authentication")
	fs.BoolVar(&o.EnableOidcAuth, prefix+"enable-oidc-auth", o.EnableOidcAuth, "enable oidc token auth to connect with Inventory API service")
	fs.BoolVar(&o.Insecure, prefix+"insecure-client", o.Insecure, "the http client that connects to kessel should not verify certificates.")
}

func (o *Options) Validate() []error {
	var errs []error

	if len(o.InventoryURL) == 0 {
		errs = append(errs, fmt.Errorf("kessel url may not be empty"))
	}

	return errs
}

func (o *Options) Complete() []error {
	var errs []error

	return errs
}
