package main

import (
	"encoding/json"
	"fmt"
	"go/types"
	"os"
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
	Recv     string   `json:",omitempty"`
}

type Struct struct {
	Methods map[string]Item `json:",omitempty"`
	Fields  map[string]Item `json:",omitempty"`
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

	var recv string
	if r := sig.Recv(); r != nil {
		recv = r.Type().String()
	}

	item := Item{
		ObjectType: "Func",
		Func: Func{
			ArgTypes: args,
			ResTypes: res,
			Variadic: sig.Variadic(),
			Recv:     recv,
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
	s := t.Type().Underlying().(*types.Struct)

	fields := make(map[string]Item)
	for i := 0; i < s.NumFields(); i++ {
		v := s.Field(i)
		fields[v.Name()] = handleVar(v)
	}

	mset := types.NewMethodSet(t.Type())

	methods := make(map[string]Item)
	for i := 0; i < mset.Len(); i++ {
		f := mset.At(i).Obj().(*types.Func)
		methods[f.Name()] = handleFunc(f)
	}
	return Item{
		ObjectType: "Struct",
		Type:       t.Type().String(),
		Struct: Struct{
			Fields:  fields,
			Methods: methods,
		},
	}
}

func handleInterface(t *types.TypeName) Item {
	i := t.Type().Underlying().(*types.Interface)
	fmt.Println(i)

	return Item{ObjectType: "Interface"}
}

func handleTypeName(t *types.TypeName) Item {
	fmt.Println("TypeName:", t.Name(), t.Type())

	var item Item

	switch t.Type().Underlying().(type) {
	case *types.Struct:
		item = handleStruct(t)
	case *types.Interface:
		item = handleInterface(t)
	default:
		fmt.Println("Warning: TypeName", t.Type().Underlying(), "is not implemented")
	}

	return item
}

func getPkgInfo(name string) map[string]Item {
	var conf loader.Config
	conf.Import(name)
	//conf.Import("github.com/barrucadu/logdb")
	prog, err := conf.Load()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(prog)
	pkg := prog.Package(name).Pkg

	scope := pkg.Scope()
	fmt.Println(scope)

	items := make(map[string]Item)

	for _, elem := range scope.Names() {
		obj := scope.Lookup(elem)

		if len(os.Args) > 1 {
			if obj.Name() != os.Args[1] {
				continue
			}
		}

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

	return items
}

func main() {
	inf := getPkgInfo("github.com/kragniz/testpkg2")
	b, err := json.MarshalIndent(inf, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))
}
