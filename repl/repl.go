package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"

	lx "github.com/theawakener0/zod/lexer"
	ps "github.com/theawakener0/zod/parser"
)

const PROMPT = "\x1b[0;32m>>\x1b[0m "


func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf(PROMPT)

		scanned := scanner.Scan()
		if !scanned {
			return
		}
		
		line := scanner.Text()
		
		switch line {
		case "/clear":
			fmt.Printf("\x1b[2J\x1b[H")
			continue
		case "/exit":
			os.Exit(0)
		}

		l := lx.New(line)
		p := ps.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParseErrors(out, p.Errors())
			continue
		}
		
		io.WriteString(out, program.String())
		io.WriteString(out, "\n")
	}
}

func printParseErrors(out io.Writer, error []string) {
	io.WriteString(out, "We ran into some problems while parsing your program.\n")
	io.WriteString(out, "Parse errors:\n")
	for _, msg := range error {
		io.WriteString(out, "\t" + msg + "\n")
	}
}

