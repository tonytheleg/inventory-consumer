package consumer

import (
	"fmt"

	"github.com/project-kessel/inventory-consumer/consumer/auth"
	"github.com/project-kessel/inventory-consumer/consumer/retry"
	"github.com/spf13/pflag"
)

type Options struct {
	BootstrapServers   []string       `mapstructure:"bootstrap-servers"`
	ConsumerGroupID    string         `mapstructure:"consumer-group-id"`
	Topics             []string       `mapstructure:"topics"`
	SessionTimeout     string         `mapstructure:"session-timeout"`
	HeartbeatInterval  string         `mapstructure:"heartbeat-interval"`
	MaxPollInterval    string         `mapstructure:"max-poll-interval"`
	EnableAutoCommit   string         `mapstructure:"enable-auto-commit"`
	AutoOffsetReset    string         `mapstructure:"auto-offset-reset"`
	StatisticsInterval string         `mapstructure:"statistics-interval-ms"`
	Debug              string         `mapstructure:"debug"`
	RetryOptions       *retry.Options `mapstructure:"retry-options"`
	AuthOptions        *auth.Options  `mapstructure:"auth"`
}

func NewOptions() *Options {
	return &Options{
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
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}
	fs.StringSliceVar(&o.BootstrapServers, prefix+"bootstrap-servers", o.BootstrapServers, "sets the bootstrap server address and port for Kafka")
	fs.StringVar(&o.ConsumerGroupID, prefix+"consumer-group-id", o.ConsumerGroupID, "sets the Kafka consumer group name (default: inventory-consumer)")
	fs.StringArrayVar(&o.Topics, prefix+"topics", o.Topics, "Kafka topic to monitor for events")
	fs.StringVar(&o.SessionTimeout, prefix+"session-timeout", o.SessionTimeout, "time a consumer can live without sending heartbeat (default: 45000ms)")
	fs.StringVar(&o.HeartbeatInterval, prefix+"heartbeat-interval", o.HeartbeatInterval, "interval between heartbeats sent to Kafka (default: 3000ms, must be lower then session-timeout)")
	fs.StringVar(&o.MaxPollInterval, prefix+"max-poll-interval", o.MaxPollInterval, "length of time consumer can go without polling before considered dead (default: 300000ms)")
	fs.StringVar(&o.EnableAutoCommit, prefix+"enable-auto-commit", o.EnableAutoCommit, "enables auto commit on consumer when messages are consumed (default: false)")
	fs.StringVar(&o.AutoOffsetReset, prefix+"auto-offset-reset", o.AutoOffsetReset, "action to take when there is no initial offset in offset store (default: earliest)")
	fs.StringVar(&o.StatisticsInterval, prefix+"statistics-interval-ms", o.StatisticsInterval, "librdkafka statistics emit interval (default: 30000ms)")
	fs.StringVar(&o.Debug, prefix+"debug", o.Debug, "a comma-separated list of debug contexts to enable (default: \"\"")

	o.AuthOptions.AddFlags(fs, prefix+"auth")
	o.RetryOptions.AddFlags(fs, prefix+"retry-options")
}

func (o *Options) Validate() []error {
	var errs []error

	if len(o.BootstrapServers) == 0 {
		errs = append(errs, fmt.Errorf("bootstrap servers can not be empty"))
	}

	if len(o.Topics) == 0 {
		errs = append(errs, fmt.Errorf("topic value can not be empty"))
	}
	return errs
}

func (o *Options) Complete() []error {
	return nil
}
