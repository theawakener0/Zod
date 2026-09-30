package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/theawakener0/Zod/compiler"
	ev "github.com/theawakener0/Zod/evaluator"
	lx "github.com/theawakener0/Zod/lexer"
	obj "github.com/theawakener0/Zod/object"
	ps "github.com/theawakener0/Zod/parser"
	"github.com/theawakener0/Zod/vm"
)

const PROMPT = "\x1b[0;32m>>\x1b[0m "

func Start(in io.Reader, out io.Writer, engine string) {
	scanner := bufio.NewScanner(in)
	env := obj.NewEnviroment()

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

		if engine == "--eng=vm" {
			var err error
			constants, err = runVM(line, out, constants, globals, symbolTable, cwd)
			_ = err
			continue
		}

		if engine == "--eng=eval" {
			_ = runEval(line, env, out)
			continue
		}

	}
}

func runVM(source string, out io.Writer, constants []obj.Object, globals []obj.Object, symbolTable *compiler.SymbolTable, baseDir string) ([]obj.Object, error) {
	l := lx.New(source)
	p := ps.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParseErrors(out, p.Errors())
		return constants, fmt.Errorf("parse error")
	}

	comp := compiler.NewWithState(symbolTable, constants)
	comp.BaseDir = baseDir
	err0 := comp.Compile(program)
	if err0 != nil {
		fmt.Fprintf(out, "Oh shit here we go again! Compilation failed:\n %s\n", err0)
		return constants, err0
	}

	bytecode := comp.Bytecode()

	machine := vm.NewWithGlobalsStore(bytecode, globals)
	err1 := machine.Run()
	if err1 != nil {
		fmt.Fprintf(out, "Oh shit here we go again! Executing bytecode failed:\n %s\n", err1)
		return bytecode.Constant, err1
	}

	LastPopped := machine.LastPoppedStackElem()
	if LastPopped != nil && LastPopped.Type() != obj.NULL_OBJ {
		io.WriteString(out, LastPopped.Inspect())
		io.WriteString(out, "\n")
	}

	errObj, ok := LastPopped.(*obj.Error)
	if ok {
		return bytecode.Constant, fmt.Errorf("%s", errObj.Message)
	}

	return bytecode.Constant, nil
}

func Execute(source string, out io.Writer, engine string, baseDir string) error {
	if engine == "--eng=vm" {
		constants := make([]obj.Object, 0, 256)
		globals := make([]obj.Object, vm.GlobalsSize)
		symbolTable := compiler.NewSymbolTable()
		for i, v := range obj.Builtins {
			symbolTable.DefineBuiltin(i, v.Name)
		}

		_, err := runVM(source, out, constants, globals, symbolTable, baseDir)
		return err
	}
	if engine == "--eng=eval" {
		env := obj.NewEnviroment()
		return runEval(source, env, out)
	}
	return fmt.Errorf("unknown engine %s", engine)
}

func runEval(source string, env *obj.Enviroment, out io.Writer) error {
	l := lx.New(source)
	p := ps.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParseErrors(out, p.Errors())
		return fmt.Errorf("parse error")
	}

	eval := ev.Eval(program, env)
	if eval != nil && eval.Type() != obj.NULL_OBJ {
		io.WriteString(out, eval.Inspect())
		io.WriteString(out, "\n")
	}

	err, ok := eval.(*obj.Error)
	if ok {
		return fmt.Errorf("%s", err.Message)
	}

	return nil
}

func printParseErrors(out io.Writer, error []string) {
	io.WriteString(out, "We ran into some problems while parsing your program.\n")
	io.WriteString(out, "Parse errors:\n")
	for _, msg := range error {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
