package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sthussey/rcon/internal"
)

func main() {
	configPath := flag.String("c", "", "Path to context config YAML or JSON file.")
	rc := internal.NewRunConfig(*configPath)
	err := rc.Run()

	if err != nil {
		fmt.Printf("Error running context: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}
