package main

const (
	GLOBAL_SCOPE  = "global"
	LOCAL_SCOPE   = "local"
	BUILTIN_SCOPE = "builtin"
)

type Symbol struct {
	Name  string
	Index int
	Scope string
}

type SymbolTable struct {
	Outer *SymbolTable

	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{Outer: outer, store: s}
}

func (s *SymbolTable) DefineBuiltin(name string) Symbol {
	symbol := Symbol{Name: name, Index: s.numDefinitions, Scope: BUILTIN_SCOPE}
	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: s.numDefinitions}
	if s.Outer != nil {
		symbol.Scope = LOCAL_SCOPE
	} else {
		symbol.Scope = GLOBAL_SCOPE
	}

	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

func (s *SymbolTable) ResolveInner(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	return obj, ok
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	if !ok && s.Outer != nil {
		obj, ok := s.Outer.Resolve(name)
		if !ok {
			return obj, ok
		}
	}
	return obj, ok
}
