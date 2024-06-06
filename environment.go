package main

type Environment struct {
	store     map[string]Object
	enclosing *Environment
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s}
}

func NewEnclosingEnvironment(enclosing *Environment) *Environment {
	env := NewEnvironment()
	env.enclosing = enclosing
	return env
}

func (e *Environment) Define(name string, value Object) Object {
	e.store[name] = value
	return value
}

// 1. Check Current Scope: If the variable exists in the current scope, update it.
// 2. Check Enclosing Scope: If the variable does not exist in the current scope but exists in the enclosing scope, update it in the enclosing scope.
// 3. Define in Current Scope: If the variable does not exist in any enclosing scope, define it in the current scope.
func (e *Environment) Set(name string, value Object) Object {
	if _, exists := e.store[name]; exists {
		e.store[name] = value
		return value
	}
	if e.enclosing != nil {
		if _, exists := e.enclosing.Get(name); exists {
			return e.enclosing.Set(name, value)
		}
	}
	e.store[name] = value
	return value
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.enclosing != nil {
		obj, ok = e.enclosing.Get(name)
	}
	return obj, ok
}
