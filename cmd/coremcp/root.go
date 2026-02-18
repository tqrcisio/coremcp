package main

import (
	"os"
	"path/filepath"

	"github.com/corebasehq/coremcp/pkg/config"
	"github.com/spf13/cobra"
)

var cfg *config.Config

// defaultConfigPath returns the path to coremcp.yaml next to the running binary.
// Falls back to the plain filename (CWD-relative) if the executable path cannot be determined.
func defaultConfigPath() string {
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "coremcp.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return "coremcp.yaml"
}

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
	rootCmd.PersistentFlags().StringP("config", "c", defaultConfigPath(), "config file path")
}
