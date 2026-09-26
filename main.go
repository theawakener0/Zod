package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

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

var version = "v0.6.0"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("zod %s\n", version)
		return
	}

	engine := "--eng=vm"
	fileIdx := 1
	if len(os.Args) > 1 && (os.Args[1] == "--eng=eval" || os.Args[1] == "--eng=vm") {
		engine = os.Args[1]
		fileIdx = 2
	} else if len(os.Args) > 1 && len(os.Args[1]) > 6 && os.Args[1][:6] == "--eng=" {
		fmt.Fprintln(os.Stderr, "unknown engine "+os.Args[1]+" (expected --eng=vm or --eng=eval)")
		os.Exit(1)
	}

	if len(os.Args) > fileIdx {
		source, err := os.ReadFile(os.Args[fileIdx])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		repl.Execute(string(source), os.Stdout, engine, filepath.Dir(os.Args[fileIdx]))
		return
	}

	if fileIdx == 2 {
		// `zod --eng=<x>` with no file: REPL with explicit engine.
		user, err := user.Current()
		if err != nil {
			panic(err)
		}

		fmt.Printf("\x1b[0;34m%s\x1b[0m\n", Banner)
		fmt.Printf("\nHello %s! Type the command here.\n", user.Username)

		repl.Start(os.Stdin, os.Stdout, engine)
		return
	}

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("\x1b[0;34m%s\x1b[0m\n", Banner)
	fmt.Printf("\nHello %s! Type the command here.\n", user.Username)

	repl.Start(os.Stdin, os.Stdout, "--eng=vm")

}
