package main

import (
	"fmt"
	"log"

	"conductor/pkg/config"
)

func main() {
	cfg, err := config.Load("configs")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	fmt.Printf("Loaded configuration: %+v\n", cfg)
}
