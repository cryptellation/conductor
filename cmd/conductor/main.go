package main

import (
	"context"
	"os"

	"github.com/lerenn/conductor/pkg/conductor"
	"github.com/lerenn/conductor/pkg/config"
	"github.com/lerenn/conductor/pkg/logging"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var configPath string

func main() {
	logging.Init()

	var rootCmd = &cobra.Command{
		Use:   "conductor",
		Short: "Conductor orchestrates your repositories",
		Run: func(_ *cobra.Command, _ []string) {
			cfg, err := config.Load(configPath)
			if err != nil {
				logging.L().Fatal("Failed to load config", zap.Error(err))
			}

			token := os.Getenv("GITHUB_TOKEN")
			if token == "" {
				logging.L().Fatal("GITHUB_TOKEN environment variable is not set")
			}

			c, err := conductor.New(cfg, token)
			if err != nil {
				logging.L().Fatal("Failed to create conductor", zap.Error(err))
			}
			defer c.Close()

			ctx := context.Background()
			c.RunWithLogging(ctx)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "configs/conductor.yaml", "Path to the config file")

	if err := rootCmd.Execute(); err != nil {
		logging.L().Error("Command execution failed", zap.Error(err))
		os.Exit(1)
	}
}
