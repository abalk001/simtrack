package main

import (
	"fmt"
	"os"

	"github.com/abalk001/simtrack/internal/simtrack"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "run":
		os.Exit(cmdRun(os.Args[2:]))
	case "-h", "--help", "help":
		usage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: simtrack <command> [args]

Commands:
  run [--] <command> [args...]   Execute a command and record a manifest
  init <name>                    Create a new SimTrack project       (PR 2)
  list                           List runs in the current project   (PR 3)
  show <run-id>                  Show a run's manifest              (PR 3)
  reproduce <run-id>             Print reproduction instructions    (PR 3)`)
}

func cmdRun(args []string) int {
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: simtrack run [--] <command> [args...]")
		return 2
	}
	code, err := simtrack.Run(simtrack.RunOptions{Args: args})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return code
}
