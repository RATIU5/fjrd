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

	var cfg toml.FjrdConfig
	err = toml.ParseConfig(content, &cfg)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	if err = cfg.Execute(); err != nil {
		fmt.Printf("%v", err)
		return
	}
}
