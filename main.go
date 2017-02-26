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
}

type Func struct {
	ParamTypes []string
}

func handleFunc(f *types.Func) Item {
	fmt.Println("Func:", f.Name())

	sig := f.Type().(*types.Signature)
	params := []string{}
	for i := 0; i < sig.Params().Len(); i++ {
		t := strings.Split(sig.Params().At(i).String(), " ")
		params = append(params, t[len(t)-1])
	}
	item := Item{ObjectType: "Func", Func: Func{ParamTypes: params}}
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

func handleTypeName(t *types.TypeName) {
	fmt.Println("TypeName:", t.Name(), t.Type())
	switch t.Type().Underlying().(type) {
	case *types.Struct:
		fmt.Println("Struct!", t)
		s := t.Type().Underlying().(*types.Struct)
		for i := 0; i < s.NumFields(); i++ {
			v := s.Field(i)
			handleVar(v)
		}
	default:
		fmt.Println("Warning: TypeName", t.Type().Underlying(), "is not implemented")
	}
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
				handleTypeName(obj.(*types.TypeName))
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
