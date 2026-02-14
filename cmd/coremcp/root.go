package main

import (
	"os"

	"github.com/corebasehq/coremcp/pkg/config"
	"github.com/spf13/cobra"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "coremcp",
	Short: "CoreMCP by CoreBaseHQ",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		loadedCfg, err := config.LoadConfig(configPath)
		if err != nil {
			return err
		}
		cfg = loadedCfg

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags here
	rootCmd.PersistentFlags().StringP("config", "c", "coremcp.yaml", "config file path")
}
