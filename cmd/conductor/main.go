package main

import (
	"context"
	"fmt"
	"os"

	"github.com/lerenn/conductor/pkg/conductor"
	"github.com/lerenn/conductor/pkg/config"
	"github.com/spf13/cobra"
)

var configPath string

func main() {
	var rootCmd = &cobra.Command{
		Use:   "conductor",
		Short: "Conductor orchestrates your repositories",
		Run: func(_ *cobra.Command, _ []string) {
			cfg, err := config.Load(configPath)
			if err != nil {
				panic(err)
			}

			token := os.Getenv("GITHUB_TOKEN")
			if token == "" {
				panic("GITHUB_TOKEN environment variable is not set")
			}

			c := conductor.New(cfg, token)

			ctx := context.Background()
			c.RunWithLogging(ctx)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "configs/conductor.yaml", "Path to the config file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
