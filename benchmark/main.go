package main

import (
	"fmt"
	"time"

	"github.com/theawakener0/Zod/compiler"
	lx "github.com/theawakener0/Zod/lexer"
	obj "github.com/theawakener0/Zod/object"
	ps "github.com/theawakener0/Zod/parser"
	"github.com/theawakener0/Zod/vm"
)

var input = "fibonacci := fn(x) { if (x==0) { return 0; } else { if (x==1) { return 1; } else { fibonacci(x-1) + fibonacci(x-2) } } }; fibonacci(15)"

func main() {
	var duration time.Duration
	var result obj.Object

	l := lx.New(input)
	p := ps.New(l)
	program := p.ParseProgram()

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

	fmt.Printf(
		"engine=vm, result=%s, duration=%s\n",
		result.Inspect(),
		duration,
	)
}
