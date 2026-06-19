package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"

	lx "github.com/theawakener0/zod/lexer"
	tk "github.com/theawakener0/zod/token"
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

		for tok := l.NextToken(); tok.Type != tk.EOF; tok = l.NextToken() {
			fmt.Printf("%+v\n", tok)
		}

	}
}

