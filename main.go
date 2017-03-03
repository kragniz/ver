package main

import (
	"encoding/json"
	"fmt"
	"go/types"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/tools/go/loader"
)

type RequiredBump int

const (
	Major RequiredBump = iota
	Minor
	Patch
)

type VerFile struct {
	Items map[string]Item
}

type Item struct {
	Kind   string
	Type   string
	Func   Func
	Struct Struct
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

func (i Item) String() string {
	return fmt.Sprintf("{Kind:%s, Type:%s, Func:%s, Struct:%s}",
		i.Kind, i.Type, i.Func, i.Struct)
}

func (f Func) String() string {
	return fmt.Sprintf("{ArgTypes:%s, ResTypes:%s, Variadic:%t, Recv:%s}",
		f.ArgTypes, f.ResTypes, f.Variadic, f.Recv)
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
		Kind: "Func",
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
	item := Item{Kind: "Var", Type: v.Type().String()}
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
		Kind: "Struct",
		Type: t.Type().String(),
		Struct: Struct{
			Fields:  fields,
			Methods: methods,
		},
	}
}

func handleInterface(t *types.TypeName) Item {
	i := t.Type().Underlying().(*types.Interface)
	fmt.Println(i)

	return Item{Kind: "Interface"}
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

func GetPkgInfo(name string) map[string]Item {
	var conf loader.Config
	conf.Import(name)
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

		if obj.Exported() {
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
	}

	return items
}

func (a *Func) diff(b Func) RequiredBump {
	// jesus christ there must be a better way
	if len(a.ArgTypes) == len(b.ArgTypes) {
		for i, v := range a.ArgTypes {
			if v != b.ArgTypes[i] {
				fmt.Println("it's the ArgTypes")
				fmt.Println("it's the ArgTypes:", a.ArgTypes, "vs", b.ArgTypes)
				return Major
			}
		}
	} else {
		return Major
	}

	if len(a.ResTypes) == len(b.ResTypes) {
		for i, v := range a.ResTypes {
			if v != b.ResTypes[i] {
				fmt.Println("it's the ResTypes:", a.ResTypes, "vs", b.ResTypes)
				return Major
			}
		}
	} else {
		return Major
	}

	if a.Recv != b.Recv {
		return Major
	}

	if a.Variadic != b.Variadic {
		return Major
	}

	return Patch
}

func (a *Struct) diff(b Struct) RequiredBump {
	return Patch
}

func diff(a, b map[string]Item) {
	changedThings := 0

	for k, v := range a {
		fmt.Println(k)
		switch v.Kind {
		case "Func":
			change := v.Func.diff(b[k].Func)
			if change == Major {
				fmt.Println("", v, "and\n", b[k], "are different")
				changedThings += 1
			}
		case "Struct":
			change := v.Struct.diff(b[k].Struct)
			if change == Major {
				fmt.Println("", v, "and\n", b[k], "are different")
				changedThings += 1
			}
		default:
			fmt.Println("diffing type", v.Kind, "isn't supported yet")
		}
	}

	fmt.Println("---")

	for k, _ := range b {
		fmt.Println(k)
	}

	fmt.Println("changed things:", changedThings)
}

func main() {
	var pkgName string
	if len(os.Args) >= 2 {
		pkgName = os.Args[1]
	} else {
		fmt.Println("Not enough args")
		os.Exit(1)
	}

	file, err := ioutil.ReadFile("v1.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}

	var verFile VerFile
	json.Unmarshal(file, &verFile)
	fmt.Println(verFile.Items)

	items := GetPkgInfo(pkgName)

	newVerFile := VerFile{Items: items}
	b, err := json.MarshalIndent(newVerFile, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))

	diff(verFile.Items, newVerFile.Items)
}
