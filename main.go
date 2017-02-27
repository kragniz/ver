package main

import (
	"encoding/json"
	"fmt"
	"go/types"
	"strings"

	"golang.org/x/tools/go/loader"
)

type Item struct {
	ObjectType string
	Type       string
	Func       Func
	Struct     Struct
}

type Func struct {
	ArgTypes []string `json:",omitempty"`
	ResTypes []string `json:",omitempty"`
	Variadic bool     `json:",omitempty"`
}

type Struct struct {
	Methods []Item `json:",omitempty"`
	Fields  []Item `json:",omitempty"`
}

func typeTupleToSlice(types *types.Tuple) []string {
	l := []string{}
	for i := 0; i < types.Len(); i++ {
		t := strings.Split(types.At(i).String(), " ")
		l = append(l, t[len(t)-1])
	}
	return l
}

func handleFunc(f *types.Func) Item {
	fmt.Println("Func:", f.Name())

	sig := f.Type().(*types.Signature)
	args := typeTupleToSlice(sig.Params())
	res := typeTupleToSlice(sig.Results())

	item := Item{
		ObjectType: "Func",
		Func: Func{
			ArgTypes: args,
			ResTypes: res,
			Variadic: sig.Variadic(),
		},
	}
	return item
}

func handleVar(v *types.Var) Item {
	fmt.Println("Var:", v.Name(), v.Type())
	item := Item{ObjectType: "Var", Type: v.Type().String()}
	return item
}

func handleConst(c *types.Const) {
	fmt.Println("Const:", c.Name())
}

func handleStruct(t *types.TypeName) Item {
	fields := []Item{}

	s := t.Type().Underlying().(*types.Struct)
	for i := 0; i < s.NumFields(); i++ {
		v := s.Field(i)
		fields = append(fields, handleVar(v))
	}
	return Item{
		ObjectType: "Struct",
		Type:       t.Type().String(),
		Struct: Struct{
			Fields: fields,
		},
	}
}

func handleTypeName(t *types.TypeName) Item {
	fmt.Println("TypeName:", t.Name(), t.Type())

	var item Item

	switch t.Type().Underlying().(type) {
	case *types.Struct:
		item = handleStruct(t)
	default:
		fmt.Println("Warning: TypeName", t.Type().Underlying(), "is not implemented")
	}

	return item
}

func main() {
	var conf loader.Config
	conf.Import("github.com/kragniz/testpkg")
	prog, err := conf.Load()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(prog)

	for pkg, _ := range prog.AllPackages {
		scope := pkg.Scope()
		fmt.Println(scope)

		items := make(map[string]Item)

		for _, elem := range scope.Names() {
			obj := scope.Lookup(elem)
			switch obj.(type) {
			case *types.Func:
				items[obj.Name()] = handleFunc(obj.(*types.Func))
			case *types.Var:
				items[obj.Name()] = handleVar(obj.(*types.Var))
			case *types.Const:
				handleConst(obj.(*types.Const))
			case *types.TypeName:
				items[obj.Name()] = handleTypeName(obj.(*types.TypeName))
			case *types.Label, *types.PkgName, *types.Builtin, *types.Nil:
				// unimplemented
				fmt.Println("Warning,", obj.Type(), "is unimplemented")
				continue
			default:
				fmt.Println("not sure what it is", obj)
			}
		}
		b, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(b))
	}
}
