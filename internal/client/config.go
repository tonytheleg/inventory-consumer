package kessel

type Config struct {
	*Options
}

func NewConfig(o *Options) *Config {
	return &Config{Options: o}
}

type tokenClientConfig struct {
	clientId       string
	clientSecret   string
	url            string
	enableOIDCAuth bool
	insecure       bool
}

type completedConfig struct {
	*Options
	tokenConfig *tokenClientConfig
}

type CompletedConfig struct {
	*completedConfig
}

func (c *Config) Complete() (CompletedConfig, []error) {
	tokenReq := &tokenClientConfig{
		clientId:       c.ClientId,
		clientSecret:   c.ClientSecret,
		url:            c.TokenEndpoint,
		enableOIDCAuth: c.EnableOidcAuth,
		insecure:       c.Insecure,
	}

	return CompletedConfig{
		&completedConfig{
			Options:     c.Options,
			tokenConfig: tokenReq,
		},
	}, nil
}
