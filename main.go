package main

import (
	"fmt"
	"go/importer"
)

func main() {
	imp := importer.Default()
	hello, err := imp.Import("github.com/kragniz/proxy/pkg")
	if err != nil {
		fmt.Println("you done fucked up:", err)
	}
	pkgScope := hello.Scope()
	fmt.Println(pkgScope)
}
