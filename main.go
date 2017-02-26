package main

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/loader"
)

func handleFunc(f *types.Func) {
	fmt.Println("Func:", f.Name())
}

func handleVar(v *types.Var) {
	fmt.Println("Var:", v.Name(), v.Type())
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
		fmt.Println("dunno")
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

		for _, elem := range scope.Names() {
			obj := scope.Lookup(elem)
			switch obj.(type) {

			/*Object = *Func  // function, concrete method, or abstract method
			  | *Var          // variable, parameter, result, or struct field
			  | *Const        // constant
			  | *TypeName     // type name
			  | *Label        // statement label
			  | *PkgName      // package name, e.g. json after import "encoding/json"
			  | *Builtin      // predeclared function such as append or len
			  | *Nil          // predeclared nil */

			case *types.Func:
				handleFunc(obj.(*types.Func))
			case *types.Var:
				handleVar(obj.(*types.Var))
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
	}
}
