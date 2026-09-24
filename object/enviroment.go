package object


type Enviroment struct {
	store   map[string]Object
	publics map[string]bool
	outer   *Enviroment
	Depth   int
}

func NewEnclosedEnviroment(outer *Enviroment) *Enviroment {
	env := NewEnviroment()
	if outer != nil {
		env.outer = outer
		env.Depth = outer.Depth
	}
	return env
}

func NewEnviroment() *Enviroment {
	s := make(map[string]Object)
	p := make(map[string]bool)
	return &Enviroment{store: s, publics: p, outer: nil, Depth: 0}
}

func (e *Enviroment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}
	return obj, ok
}

func (e *Enviroment) GetAll() map[string]Object {
	result := make(map[string]Object, len(e.store))

	for k, v := range e.store {
		result[k] = v
	}

	return result
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

func (e *Enviroment) SetPublic(name string) {
	if e.publics == nil {
		e.publics = make(map[string]bool)
	}
	e.publics[name] = true
}

func (e *Enviroment) IsPublic(name string) bool {
	if e.publics != nil {
		if _, ok := e.publics[name]; ok {
			return true
		}
	}
	if e.outer != nil {
		return e.outer.IsPublic(name)
	}
	return false
}

func (e *Enviroment) GetPublic(name string) (Object, bool) {
	obj, ok := e.Get(name)
	if !ok {
		return nil, false
	}
	if !e.IsPublic(name) {
		return nil, false
	}
	return obj, true
}

func (e *Enviroment) GetAllPublic() map[string]Object {
	result := make(map[string]Object, len(e.store))

	for k, v := range e.store {
		if e.IsPublic(k) {
			result[k] = v
		}
	}

	return result
}

