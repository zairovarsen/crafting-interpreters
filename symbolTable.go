package main

const (
	GLOBAL_SCOPE  = "global"
	LOCAL_SCOPE   = "local"
	BUILTIN_SCOPE = "builtin"
	UPVALUE_SCOPE = "upvalue"
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

	upvalues []Symbol
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
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
		obj, ok = s.Outer.Resolve(name)
		if !ok {
			return obj, ok
		}

		if obj.Scope == GLOBAL_SCOPE || obj.Scope == BUILTIN_SCOPE {
			return obj, ok
		}

		upvalue := s.DefineUpvalue(obj)
		return upvalue, true
	}
	return obj, ok
}

func (s *SymbolTable) DefineUpvalue(original Symbol) Symbol {
	s.upvalues = append(s.upvalues, original)
	symbol := Symbol{Name: original.Name, Index: len(s.upvalues) - 1, Scope: UPVALUE_SCOPE}
	s.store[original.Name] = symbol
	return symbol
}
