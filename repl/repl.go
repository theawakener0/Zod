package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"

	ev "github.com/theawakener0/Zod/evaluator"
	"github.com/theawakener0/Zod/compiler"
	lx "github.com/theawakener0/Zod/lexer"
	obj "github.com/theawakener0/Zod/object"
	ps "github.com/theawakener0/Zod/parser"
	"github.com/theawakener0/Zod/vm"
)

const PROMPT = "\x1b[0;32m>>\x1b[0m "


func Start(in io.Reader, out io.Writer, engine string) {
	scanner := bufio.NewScanner(in)
	env := obj.NewEnviroment()

	constants := []obj.Object{}
	globals := make([]obj.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()
	for i, v := range obj.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

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

		if engine == "--eng=vm" {
			constants = runVM(line, out, constants, globals, symbolTable)
			continue
		}

		if engine == "--eng=eval" {
			runEval(line, env, out)
			continue
		}

	}
}

func runVM(source string, out io.Writer, constants []obj.Object, globals []obj.Object, symbolTable *compiler.SymbolTable) []obj.Object {
	l := lx.New(source)
	p := ps.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParseErrors(out, p.Errors())
		return constants
	}

	comp := compiler.NewWithState(symbolTable, constants)
	err0 := comp.Compile(program)
	if err0 != nil {
		fmt.Fprintf(out, "Oh shit here we go again! Compilation failed:\n %s\n", err0)
		return constants
	}

	bytecode := comp.Bytecode()

	machine := vm.NewWithGlobalsStore(bytecode, globals)
	err1 := machine.Run()
	if err1 != nil {
		fmt.Fprintf(out, "Oh shit here we go again! Executing bytecode failed:\n %s\n", err1)
		return bytecode.Constant
	}

	LastPopped := machine.LastPoppedStackElem()
	io.WriteString(out, LastPopped.Inspect())
	io.WriteString(out, "\n")

	return bytecode.Constant
}

func Execute(source string, out io.Writer, engine string) {
	if engine == "--eng=vm" {
		constants := []obj.Object{}
		globals := make([]obj.Object, vm.GlobalsSize)
		symbolTable := compiler.NewSymbolTable()
		for i, v := range obj.Builtins {
			symbolTable.DefineBuiltin(i, v.Name)
		}

		runVM(source, out, constants, globals, symbolTable)
		return
	}
	if engine == "--eng=eval" {
		env := obj.NewEnviroment()

		runEval(source, env, out)
		return
	}
}

func runEval(source string, env *obj.Enviroment, out io.Writer) {
	l := lx.New(source)
	p := ps.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParseErrors(out, p.Errors())
		return
	}

	eval := ev.Eval(program, env)
	if eval != nil && eval.Type() != obj.NULL_OBJ {
		io.WriteString(out, eval.Inspect())
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

