package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"githerd/internal/config"
	"githerd/internal/worker"
)

// Version information - populated at build time by GoReleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

// buildVersion returns a formatted version string
func buildVersion() string {
	if version == "dev" {
		return fmt.Sprintf("%s (built from source)", version)
	}
	return fmt.Sprintf("%s (commit: %s, built: %s, by: %s)", version, commit, date, builtBy)
}

func main() {
	cfg := config.DefaultConfig()

	rootCmd := &cobra.Command{
		Use:     "githerd [path]",
		Short:   "Bulk git operations on multiple repositories",
		Long:    `GitHerd performs git operations (fetch/pull) on all git repositories 
found in the specified directory and its subdirectories.`,
		Version: buildVersion(),
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Setup signal handling for graceful shutdown
			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			// Add timeout if specified
			if cfg.Timeout > 0 {
				ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
				defer cancel()
			}

			// Determine root path
			rootPath := "."
			if len(args) > 0 {
				rootPath = args[0]
			}

			// Validate path
			if _, err := os.Stat(rootPath); os.IsNotExist(err) {
				return fmt.Errorf("path does not exist: %s", rootPath)
			}

			// Create and execute manager
			manager := worker.New(cfg)
			return manager.Execute(ctx, rootPath)
		},
	}

	// Setup configuration flags and viper
	config.SetupFlags(rootCmd, cfg)
	config.SetupViper(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}