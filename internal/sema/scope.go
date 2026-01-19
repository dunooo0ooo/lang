package sema

import "github.com/dunooo0ooo/lang/internal/token"

type SymbolKind int

const (
	SymVar SymbolKind = iota
	SymFn
)

type Symbol struct {
	Kind SymbolKind
	Name string
	Pos  token.Position

	// var
	Ty Type

	// fn
	Params []Type
	Ret    Type
}

type Scope struct {
	Parent *Scope
	Syms   map[string]Symbol
}

func NewScope(parent *Scope) *Scope {
	return &Scope{Parent: parent, Syms: make(map[string]Symbol)}
}

func (s *Scope) Declare(sym Symbol) bool {
	if _, ok := s.Syms[sym.Name]; ok {
		return false
	}
	s.Syms[sym.Name] = sym
	return true
}

func (s *Scope) Lookup(name string) (Symbol, bool) {
	for sc := s; sc != nil; sc = sc.Parent {
		if sym, ok := sc.Syms[name]; ok {
			return sym, true
		}
	}
	return Symbol{}, false
}
