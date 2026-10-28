package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	obj "github.com/theawakener0/Zod/object"
	"github.com/theawakener0/Zod/repl"
)

const Banner = ` 
                                                 
▄▄▄▄▄▄▄▄▄          ▄▄ ▄▄▄                        
▀▀▀▀▀████          ██ ███                        
   ▄███▀  ▄███▄ ▄████ ███       ▀▀█▄ ████▄ ▄████ 
 ▄███▀    ██ ██ ██ ██ ███      ▄█▀██ ██ ██ ██ ██ 
█████████ ▀███▀ ▀████ ████████ ▀█▄██ ██ ██ ▀████ 
                                              ██ 
                                            ▀▀▀  
`

var version = "v0.8.0"

type cliOptions struct {
	file        string
	showVersion bool
}

func parseArgs(args []string) (cliOptions, error) {
	if len(args) == 0 {
		return cliOptions{}, nil
	}

	first := args[0]
	switch first {
	case "-v", "--version":
		return cliOptions{showVersion: true}, nil
	case "--eng=vm":
		if len(args) > 1 {
			return cliOptions{file: args[1]}, nil
		}
		return cliOptions{}, nil
	case "--eng=eval":
		return cliOptions{}, fmt.Errorf("the eval engine was removed in %s; use the VM (default) or --eng=vm", version)
	default:
		if strings.HasPrefix(first, "--eng=") {
			return cliOptions{}, fmt.Errorf("unknown engine %s (expected --eng=vm)", first)
		}
		return cliOptions{file: first}, nil
	}
}

func main() {
	installSignalCleanup()	
	os.Exit(run())
}

func run() int {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if opts.showVersion {
		fmt.Printf("zod %s\n", version)
		return 0
	}

	defer obj.TermCleanup()

	if opts.file != "" {
		source, err := os.ReadFile(opts.file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		err = repl.Execute(string(source), os.Stdout, filepath.Dir(opts.file))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	}

	printWelcome()
	repl.Start(os.Stdin, os.Stdout)
	return 0
}

func printWelcome() {
	current, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("\x1b[0;34m%s\x1b[0m\n", Banner)
	fmt.Printf("\nHello %s! Type the command here.\n", current.Username)
}
