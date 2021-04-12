package main

import (
	flag "github.com/spf13/pflag"
	"fmt"
	"os"

	"github.com/sthussey/rcon/internal"
)

func main() {
	configPath := flag.String("config", "", "Path to context config JSON file.")
    dirtyExit := flag.Bool("dirty", false, "Don't clean up at exit to allow for debugging.")
	flag.Parse()
	//debug := flag.Bool("d", false, "Turn on debug logging.")
	rc, err := internal.NewRunConfig(*configPath)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	r := internal.NewLocalRunner(nil)

	val_msg := r.ValidateContext(rc.RunContext, true)

	if len(val_msg) > 0 {
		for m := range val_msg {
			fmt.Printf("%s\n", val_msg[m])
		}
		os.Exit(1)
	}

	res := r.ExecuteContext(rc.RunContext, *dirtyExit)

	if res {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
