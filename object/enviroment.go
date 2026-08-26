package object


type Enviroment struct {
	store map[string]Object
	outer *Enviroment
	Depth int
}

func NewEnclosedEnviroment(outer *Enviroment) *Enviroment {
	env := NewEnviroment()
	// Depth is retained for potential future use but recursion now uses global callDepth in evaluator.
	if outer != nil {
		env.outer = outer
		env.Depth = outer.Depth
	}
	return env
}

func NewEnviroment() *Enviroment {
	s := make(map[string]Object)
	return &Enviroment{store: s, outer: nil, Depth: 0}
}

func (e *Enviroment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}
	return obj, ok
}

func (e *Enviroment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

func (e *Enviroment) Assign(name string, val Object) bool {
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		return true
	}
	if e.outer != nil {
		return e.outer.Assign(name, val)
	}
	return false
}

