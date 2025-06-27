package main

import (
	"fmt"
	"github.com/RATIU5/fjrd/toml"
	"os"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("Usage: fjrd <path>")
		return
	}

	path := args[0]
	content, err := toml.ResolveTomlResource(path)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	fmt.Printf("%v", content)
}
