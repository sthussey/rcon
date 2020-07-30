package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sthussey/rcon/internal"
)

func main() {
	configPath := flag.String("c", "", "Path to context config JSON file.")
    flag.Parse();
    //debug := flag.Bool("d", false, "Turn on debug logging.")
	rc := internal.NewRunConfig(*configPath)

	r := internal.NewLocalRunner(nil)

	val_msg := r.ValidateContext(rc.RunContext, true)

	if len(val_msg) > 0 {
		for m := range val_msg {
			fmt.Printf("%s\n", val_msg[m])
		}
		os.Exit(1)
	}

	res := r.ExecuteContext(rc.RunContext)

	if res {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
