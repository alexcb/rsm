package main

import (
	"fmt"

	"github.com/alexcb/rsm/server"
)

func main() {
	x := server.Foo()
	fmt.Printf("hello %d\n", x)
}
