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

func (e *Environment) Set(name string, value Object) Object {
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
