package main

import "fmt"

const (
	BuiltinFuncNamePrint = "print"
)

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		BuiltinFuncNamePrint,
		&Builtin{Fn: func(args ...Object) Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return nil
		},
		},
	},
}

func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}

	return nil
}
