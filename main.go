package main

import (
	"fmt"
	"os"
	"os/user"

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

var version = "v0.4.5"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("zod %s\n", version)
		return
	}


	if len(os.Args) > 1 && os.Args[1] == "--eng=vm" {
		if len(os.Args) > 2 {
			source, err := os.ReadFile(os.Args[2])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			repl.Execute(string(source), os.Stdout, os.Args[1])
			return
		}

		user, err := user.Current()
		if err != nil {
			panic(err)
		}

		fmt.Printf("\x1b[0;34m%s\x1b[0m\n", Banner)
		fmt.Printf("\nHello %s! Type the command here.\n", user.Username)

		repl.Start(os.Stdin, os.Stdout, os.Args[1])
	}

	if len(os.Args) > 1 {
		source, err := os.ReadFile(os.Args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		repl.Execute(string(source), os.Stdout, "--eng=eval")
		return
	}

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("\x1b[0;34m%s\x1b[0m\n", Banner)
	fmt.Printf("\nHello %s! Type the command here.\n", user.Username)

	repl.Start(os.Stdin, os.Stdout, "--eng=eval")

}
