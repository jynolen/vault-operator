/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/go-logr/zapr"
	"github.com/jynolen/vault-operator/cmd/internal/cli/controller"
	initCmd "github.com/jynolen/vault-operator/cmd/internal/cli/init"
	"github.com/jynolen/vault-operator/cmd/internal/cli/unseal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	zapcore "go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vault-operator",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig(cmd)
	},
}
var cfgFile string
var zapOption = &zap.Options{
	Development: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func initializeConfig(cmd *cobra.Command) error {
	viper.SetEnvPrefix("VAULT_OPERATOR")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "*", "-", "*"))
	viper.AutomaticEnv()

	if cfgFile != "" {

		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("vault-operator")
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return err
		}
	}
	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return err
	}
	logger := zap.NewRaw(zap.UseFlagOptions(zapOption))
	zapcore.ReplaceGlobals(logger)
	zapr.NewLogger(logger).Info("Configuration initialized.", "configfile", viper.ConfigFileUsed())
	viper.WriteConfigAs(".coucou")
	return nil
}

func init() {
	flagSet := &flag.FlagSet{}
	zapOption.BindFlags(flagSet)
	rootCmd.PersistentFlags().AddGoFlagSet(flagSet)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default locations: .)")

	rootCmd.AddCommand(controller.NewCommand())
	rootCmd.AddCommand(initCmd.NewCommand())
	rootCmd.AddCommand(unseal.NewCommand())
}
