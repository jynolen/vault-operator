/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package init

import (
	"math"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	zapcore "go.uber.org/zap"
)

type InitCliConfig struct {
	VaultAddr           string
	PGPKeysPath         string
	RootTokenPgpKey     string
	RecoveryPGPKeysPath string
	UseAutoSeal         bool
	SecretShared        int
	SecretThreshold     int
	RecoveryShared      int
	RecoveryThreshold   int
}

func (c *InitCliConfig) VaultConfig() *api.Config {
	return &api.Config{
		Address: c.VaultAddr,
	}
}

// initCmd represents the init command
var (
	config  = &InitCliConfig{}
	initlog logr.Logger
)

func initCmd(cmd *cobra.Command) {
	flags := cmd.Flags()
	commonOpts, unsealOpts := &pflag.FlagSet{}, &pflag.FlagSet{}

	flags.BoolVarP(&config.UseAutoSeal, "use-auto-seal", "", false, "_TODO_.")
	flags.StringVarP(&config.VaultAddr, "vault-addr", "", "", "_TODO_.")

	commonOpts.StringVarP(&config.PGPKeysPath, "pgp-key-path", "", "", "_TODO_.")
	commonOpts.StringVarP(&config.RootTokenPgpKey, "root-token-pgp-key", "", "", "_TODO_.")
	commonOpts.IntVarP(&config.SecretShared, "secret-shared", "", 5, "_TODO_.")
	commonOpts.IntVarP(&config.SecretThreshold, "secret-threshold", "", -1, "_TODO_.")

	unsealOpts.StringVarP(&config.RecoveryPGPKeysPath, "recovery-pgp-key-path", "", "", "_TODO_.")
	unsealOpts.IntVarP(&config.RecoveryShared, "recovery-shared", "", 5, "_TODO_.")
	unsealOpts.IntVarP(&config.RecoveryThreshold, "recovery-threshold", "", -1, "_TODO_.")

	flags.AddFlagSet(commonOpts)
	flags.AddFlagSet(unsealOpts)
}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Init Command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: initMain,
	}
	initCmd(cmd)

	return cmd
}

func initMain(cmd *cobra.Command, args []string) {
	initlog = zapr.NewLogger(zapcore.L())
	if config.SecretThreshold == -1 {
		config.SecretThreshold = int(math.Ceil(float64(config.SecretShared) / 2))
	}
	if config.RecoveryThreshold == -1 {
		config.RecoveryThreshold = int(math.Ceil(float64(config.RecoveryShared) / 2))
	}

	client, err := api.NewClient(config.VaultConfig())
	if err != nil {
		initlog.Error(err, "to initialize vault client", "error", err)
		os.Exit(1)
	}

	inited, err := client.Sys().InitStatus()
	if err != nil {
		initlog.Error(err, "to check init status", "error", err)
		os.Exit(1)
	}
	if inited {
		initlog.Info("Vault has already being initialized.")
		os.Exit(0)
	}
}
