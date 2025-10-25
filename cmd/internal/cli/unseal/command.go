/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package unseal

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
)

type InitCliConfig struct {
	PGPKeysPath          string
	RootTokenPgpKey      string
	RecioveryPGPKeysPath string
	SecretShared         int
	SecretThreshold      int
	StoreShared          int
	RecoveryShared       int
	RecoveryThreshold    int
}

// initCmd represents the init command
var (
	unsealLog = logr.Logger{}
	config    = &InitCliConfig{}
)

func initCmd(cmd *cobra.Command) {
	// _ := cmd.Flags()
}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unseal",
		Short: "Unseal Command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("init called")
		},
	}
	initCmd(cmd)
	return cmd
}
