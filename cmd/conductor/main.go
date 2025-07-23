package main

import (
	"context"

	"github.com/lerenn/conductor/pkg/conductor"
	"github.com/lerenn/conductor/pkg/config"
)

func main() {
	cfg, err := config.Load("configs")
	if err != nil {
		panic(err)
	}

	token := "" // TODO: load from env or config
	c := conductor.New(cfg, token)

	ctx := context.Background()
	c.RunWithLogging(ctx)
}
