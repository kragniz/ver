package main

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/loader"
)

func handleFunc(f *types.Func) {
	fmt.Println("func", f)
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

			case *types.Func:
				handleFunc(obj.(*types.Func))
			case *types.Var:
				fmt.Println("var", obj.Name(), obj.Type())
			default:
				fmt.Println("not sure what it is", obj)
			}

		}
	}
}
