package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/theawakener0/Zod/compiler"
	eval "github.com/theawakener0/Zod/evaluator"
	lx "github.com/theawakener0/Zod/lexer"
	obj "github.com/theawakener0/Zod/object"
	ps "github.com/theawakener0/Zod/parser"
	"github.com/theawakener0/Zod/vm"
)

var engine = flag.String("engine", "vm", "use 'vm' or 'eval'")

var input = "fibonacci := fn(x) { if (x==0) { return 0; } else { if (x==1) { return 1; } else { fibonacci(x-1) + fibonacci(x-2) } } }; fibonacci(15)"

func main() {
	flag.Parse()

	var duration 	time.Duration
	var result		obj.Object

	l := lx.New(input)
	p := ps.New(l)
	program := p.ParseProgram()

	if *engine == "vm" {
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Printf("compiler error: %s", err)
			return
		}

		machine := vm.New(comp.Bytecode())

		start := time.Now()

		err = machine.Run()
		if err != nil {
			fmt.Printf("vm error: %s", err)
			return
		}

		duration = time.Since(start)
		result = machine.LastPoppedStackElem()
	} else {
		env := obj.NewEnviroment()
		start := time.Now()
		result = eval.Eval(program, env)
		duration = time.Since(start)
	}

	fmt.Printf(
		"engine=%s, result=%s, duration=%s\n",
		*engine,
		result.Inspect(),
		duration,
	)
}

