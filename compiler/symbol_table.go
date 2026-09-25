package compiler

type SymbolScope string

const (
	LocalScope    SymbolScope = "LOCAL"
	GlobalScope   SymbolScope = "GLOBAL"
	BuiltinScope  SymbolScope = "BUILTIN"
	FreeScope     SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTION"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer *SymbolTable

	store          map[string]Symbol
	numDefinitions int

	FreeSymbols []Symbol

	isFunction bool

	scopedGlobals map[string]bool
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol, 32)
	freeSymbols := make([]Symbol, 0, 4)
	return &SymbolTable{store: s, FreeSymbols: freeSymbols, isFunction: true}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	s.isFunction = true
	return s
}

func NewBlockSymbolTable(outer *SymbolTable) *SymbolTable {
	s := make(map[string]Symbol, 8)
	return &SymbolTable{
		Outer:          outer,
		store:          s,
		FreeSymbols:    make([]Symbol, 0, 4),
		numDefinitions: outer.numDefinitions,
		isFunction:     false,
	}
}

func (s *SymbolTable) Define(name string) Symbol {
	if s.Outer == nil {
		symbol := Symbol{Name: name, Index: s.numDefinitions, Scope: GlobalScope}
		s.store[name] = symbol
		s.numDefinitions++
		return symbol
	}
	if !s.isFunction {
		hasFunc := false
		root := s
		for t := s; t != nil; t = t.Outer {
			if t.isFunction && t.Outer != nil {
				hasFunc = true
				break
			}
			if t.Outer == nil {
				root = t
			}
		}
		if !hasFunc {
			symbol := Symbol{Name: name, Index: root.numDefinitions, Scope: GlobalScope}
			s.store[name] = symbol
			root.numDefinitions++
			if root.scopedGlobals == nil {
				root.scopedGlobals = make(map[string]bool)
			}
			root.scopedGlobals[name] = true
			if s.numDefinitions < root.numDefinitions {
				s.numDefinitions = root.numDefinitions
			}
			return symbol
		}
	}
	symbol := Symbol{Name: name, Index: s.numDefinitions, Scope: LocalScope}
	s.store[name] = symbol
	s.numDefinitions++

	return symbol
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	if !ok && s.Outer != nil {
		obj, ok = s.Outer.Resolve(name)
		if !ok {
			return obj, ok
		}

		if obj.Scope == BuiltinScope {
			return obj, ok
		}

		if obj.Scope == GlobalScope {
			if !s.isFunction || !s.isScopedGlobal(obj.Name) {
				return obj, ok
			}
		} else if !s.isFunction {
			return obj, ok
		}

		free := s.defineFree(obj)
		return free, true
	}
	return obj, ok
}

func (s *SymbolTable) isScopedGlobal(name string) bool {
	root := s
	for t := s; t != nil; t = t.Outer {
		if t.Outer == nil {
			root = t
		}
	}
	if root.scopedGlobals == nil {
		return false
	}
	return root.scopedGlobals[name]
}

func (s *SymbolTable) DefineIfNotExists(name string) Symbol {
	if symbol, ok := s.store[name]; ok {
		// Builtins must be shadowable (e.g. stdlib `pub push` must
		// override builtin `push`). Otherwise top-level `:=` and
		// imports of colliding names silently keep the builtin in VM.
		if symbol.Scope != FunctionScope && symbol.Scope != BuiltinScope {
			return symbol
		}
	}
	return s.Define(name)
}

func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) DefineFunctionName(name string) Symbol {
	symbol := Symbol{Name: name, Index: 0, Scope: FunctionScope}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) defineFree(original Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, original)

	symbol := Symbol{Name: original.Name, Index: len(s.FreeSymbols) - 1}
	symbol.Scope = FreeScope

	s.store[original.Name] = symbol
	return symbol
}
