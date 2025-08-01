package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/project-kessel/inventory-consumer/consumer"
	"github.com/project-kessel/inventory-consumer/internal/common"
	"github.com/project-kessel/inventory-consumer/internal/config"
	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Version = "0.1.0"
	Name    = "inventory-consumer"
	cfgFile string

	logger *log.Helper
	kic    consumer.InventoryConsumer
	err    error
	errs   []error

	rootCmd = &cobra.Command{
		Use:     Name,
		Version: Version,
		Short:   "A consumer group for replicating resources to Kessel Inventory",
	}

	options = config.NewOptionsConfig()
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

// Initialize ensures the config file is read first, ensuring log level is set from config before setting up logger
func Initialize() {
	initConfig()

	// for troubleshoot, when set to debug, configuration info is logged in more detail to stdout
	logLevel := common.GetLogLevel()
	logger, _ = common.InitLogger(logLevel, common.LoggerOptions{
		ServiceName:    Name,
		ServiceVersion: Version,
	})
	if logLevel == "debug" {
		config.LogConfigurationInfo(options)
	}
}

func init() {
	cobra.OnInitialize(Initialize)

	configHelp := fmt.Sprintf("config file (default is $PWD/.%s.yaml)", Name)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", configHelp)

	err := viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		panic(err)
	}

	if clowder.IsClowderEnabled() {
		err := options.InjectClowdAppConfig(clowder.LoadedConfig)
		if err != nil {
			panic(err)
		}
	}

	loggerOptions := common.LoggerOptions{
		ServiceName:    Name,
		ServiceVersion: Version,
	}

	startCmd := startCommand(options.Consumer, options.Client, loggerOptions)
	rootCmd.AddCommand(startCmd)
	err = viper.BindPFlags(startCmd.Flags())
	if err != nil {
		panic(err)
	}

	readyzCmd := readyzCommand(options.Client)
	rootCmd.AddCommand(readyzCmd)
	err = viper.BindPFlags(readyzCmd.Flags())
	if err != nil {
		panic(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		if configFilePath, exists := os.LookupEnv("INVENTORY_CONSUMER_CONFIG"); exists {
			absPath, err := filepath.Abs(configFilePath)
			if err != nil {
				log.Fatalf("Failed to resolve absolute path for config file: %v", err)
			}
			// Set the config file path
			viper.SetConfigFile(absPath)
			if err := viper.ReadInConfig(); err != nil {
				log.Fatalf("Error reading INVENTORY_CONSUMER_CONFIG file, %s", err)
			}
		} else {
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)

			viper.AddConfigPath(".")
			viper.AddConfigPath(home)
			viper.SetConfigType("yaml")

			viper.SetConfigName("." + Name)
		}
	}

	viper.SetEnvPrefix(Name)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	} else {
		log.Infof("Using config file: %s", viper.ConfigFileUsed())
	}

	// put the values into the options struct.
	if err := viper.Unmarshal(&options); err != nil {
		panic(err)
	}
}
