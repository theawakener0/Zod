package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/theawakener0/Zod/compiler"
	lx "github.com/theawakener0/Zod/lexer"
	obj "github.com/theawakener0/Zod/object"
	ps "github.com/theawakener0/Zod/parser"
	"github.com/theawakener0/Zod/vm"
)

const PROMPT = "\x1b[0;32m>>\x1b[0m "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	constants := make([]obj.Object, 0, 256)
	globals := make([]obj.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()
	for i, v := range obj.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}
	cwd, _ := os.Getwd()

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

		var err error
		constants, err = runVM(line, out, constants, globals, symbolTable, cwd)
		if err != nil {
			fmt.Fprintf(out, "Oh shit here we go again! %s\n", err)
		}
	}
}

func runVM(source string, out io.Writer, constants []obj.Object, globals []obj.Object, symbolTable *compiler.SymbolTable, baseDir string) (constantsOut []obj.Object, err error) {
	constantsOut = constants
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("internal error: %v", r)
			constantsOut = constants
		}
	}()

	l := lx.New(source)
	p := ps.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		return constants, fmt.Errorf("%s", formatParseErrors(p.Errors()))
	}

	comp := compiler.NewWithState(symbolTable, constants)
	comp.BaseDir = baseDir
	err0 := comp.Compile(program)
	if err0 != nil {
		return constants, fmt.Errorf("compilation failed:\n %s", err0)
	}

	bytecode := comp.Bytecode()

	machine := vm.NewWithGlobalsStore(bytecode, globals)
	err1 := machine.Run()
	if err1 != nil {
		return bytecode.Constant, fmt.Errorf("executing bytecode failed:\n %s", err1)
	}

	LastPopped := machine.LastPoppedStackElem()
	if errObj, ok := LastPopped.(*obj.Error); ok {
		return bytecode.Constant, fmt.Errorf("%s", errObj.Message)
	}
	if LastPopped != nil && LastPopped.Type() != obj.NULL_OBJ {
		io.WriteString(out, LastPopped.Inspect())
		io.WriteString(out, "\n")
	}

	return bytecode.Constant, nil
}

func Execute(source string, out io.Writer, baseDir string) error {
	constants := make([]obj.Object, 0, 256)
	globals := make([]obj.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()
	for i, v := range obj.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	_, err := runVM(source, out, constants, globals, symbolTable, baseDir)
	return err
}

func formatParseErrors(errors []string) string {
	var sb strings.Builder
	sb.WriteString("We ran into some problems while parsing your program.\nParse errors:\n")
	for _, msg := range errors {
		sb.WriteString("\t" + msg + "\n")
	}
	return sb.String()
}
