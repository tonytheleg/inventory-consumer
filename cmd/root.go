package cmd

import (
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tonytheleg/inventory-consumer/consumer"
)

var (
	logger         *log.Helper
	consumerConfig consumer.CompletedConfig
	icrg           consumer.InventoryConsumer
	err            error
	errs           []error
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "inventory-consumer",
	Short: "A consumer group for replicating resources to Kessel Inventory",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	options := consumer.NewOptions()
	startCmd := startCommand(options)
	rootCmd.AddCommand(startCmd)
	err = viper.BindPFlags(startCmd.Flags())
	if err != nil {
		panic(err)
	}
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.consumer-test-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
