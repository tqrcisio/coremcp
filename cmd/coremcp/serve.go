package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/corebasehq/coremcp/pkg/adapter"
	"github.com/corebasehq/coremcp/pkg/config"
	"github.com/corebasehq/coremcp/pkg/security"
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

		// Configure security features
		log.Println("Configuring security features...")
		piiPatterns := convertPIIPatterns(cfg.Security.PIIPatterns)
		if err := mcpSrv.ConfigureSecurity(
			cfg.Security.MaxRowLimit,
			cfg.Security.EnablePIIMasking,
			piiPatterns,
			cfg.Security.AllowedKeywords,
			cfg.Security.BlockedKeywords,
		); err != nil {
			log.Fatalf("CRITICAL: Failed to configure security: %v", err)
		}
		log.Printf("Security configured: MaxRowLimit=%d, PIIMasking=%v",
			cfg.Security.MaxRowLimit, cfg.Security.EnablePIIMasking)

		for _, sourceCfg := range cfg.Sources {
			src, err := adapter.NewSource(sourceCfg.Type, sourceCfg.DSN, sourceCfg.NoLock, sourceCfg.NormalizeTurkish)
			if err != nil {
				log.Printf("ERROR: Failed to create source %s: %v\n", sourceCfg.Name, err)
				continue
			}

			if err := src.Connect(cmd.Context()); err != nil {
				log.Printf("ERROR: Failed to connect to database %s: %v — skipping source", sourceCfg.Name, err)
				continue
			}

			mcpSrv.AddSource(sourceCfg.Name, src, sourceCfg.IsReadOnly())
			log.Printf("Source ready: %s (%s) [ReadOnly: %v, NoLock: %v, NormalizeTurkish: %v]", sourceCfg.Name, sourceCfg.Type, sourceCfg.IsReadOnly(), sourceCfg.NoLock, sourceCfg.NormalizeTurkish)
		}

		// Load database schemas for AI context in background
		// so the MCP server can respond to initialize immediately
		go func() {
			log.Println("Loading database schemas for AI context (background)...")
			time.Sleep(500 * time.Millisecond) // let stdio start first
			if err := mcpSrv.LoadSchemas(cmd.Context()); err != nil {
				log.Printf("WARNING: Failed to load schemas: %v", err)
			} else {
				log.Println("Database schemas loaded successfully!")
			}
		}()

		// Register custom tools from config
		if len(cfg.CustomTools) > 0 {
			log.Printf("Registering %d custom tool(s)...", len(cfg.CustomTools))
			for _, toolCfg := range cfg.CustomTools {
				params := make([]server.ToolParam, len(toolCfg.Parameters))
				for i, p := range toolCfg.Parameters {
					params[i] = server.ToolParam{
						Name:     p.Name,
						Type:     p.Type,
						Required: p.Required,
						Default:  p.Default,
					}
				}

				if err := mcpSrv.AddCustomTool(
					toolCfg.Name,
					toolCfg.Description,
					toolCfg.Source,
					toolCfg.Query,
					params,
				); err != nil {
					log.Printf("WARNING: Failed to register custom tool %s: %v", toolCfg.Name, err)
				} else {
					log.Printf("Custom tool registered: %s", toolCfg.Name)
				}
			}
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

// convertPIIPatterns converts config PII patterns to security PII patterns.
func convertPIIPatterns(configPatterns []config.PIIMaskPattern) []security.MaskPattern {
	patterns := make([]security.MaskPattern, len(configPatterns))
	for i, p := range configPatterns {
		patterns[i] = security.MaskPattern{
			Name:        p.Name,
			Pattern:     p.Pattern,
			Replacement: p.Replacement,
			Enabled:     p.Enabled,
		}
	}
	return patterns
}
