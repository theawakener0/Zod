package main

import (
	"fmt"
	"os"
	"os/user"

	"github.com/theawakener0/zod/repl"
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

func main() {
	if len(os.Args) > 1 {
		source, err := os.ReadFile(os.Args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		repl.Execute(string(source), os.Stdout)
		return
	}

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("\x1b[0;34m%s\x1b[0m\n", Banner)
	fmt.Printf("\nHello %s! Type the command here.\n", user.Username)

	repl.Start(os.Stdin, os.Stdout)

}
