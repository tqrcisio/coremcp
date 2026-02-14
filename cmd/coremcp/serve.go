package main

import (
	"fmt"
	"log"
	"os"

	"github.com/corebasehq/coremcp/pkg/adapter"
	"github.com/corebasehq/coremcp/pkg/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the MCP server",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(os.Stderr)

		if cfg == nil {
			log.Fatal("Failed to load configuration")
		}

		fmt.Fprintf(os.Stderr, "Starting CoreMCP Server via %s...\n", cfg.Server.Transport)

		mcpSrv := server.NewMCPServer(cfg.Server.Name, cfg.Server.Version)

		for _, sourceCfg := range cfg.Sources {
			src, err := adapter.NewSource(sourceCfg.Type, sourceCfg.DSN)
			if err != nil {
				log.Printf("ERROR: Failed to create source %s: %v\n", sourceCfg.Name, err)
				continue
			}

			if err := src.Connect(cmd.Context()); err != nil {
				log.Fatalf("CRITICAL: Failed to connect to database %s: %v", sourceCfg.Name, err)
			}

			mcpSrv.AddSource(sourceCfg.Name, src, sourceCfg.ReadOnly)
			log.Printf("Source ready: %s (%s) [ReadOnly: %v]", sourceCfg.Name, sourceCfg.Type, sourceCfg.ReadOnly)
		}

		transport, _ := cmd.Flags().GetString("transport")
		if transport == "stdio" {
			log.Println("CoreMCP started on Stdio. Waiting for MCP client...")
			if err := mcpSrv.StartStdio(); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal("HTTP transport is not supported yet.")
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringP("transport", "t", "stdio", "Transport type: stdio or http")
}
