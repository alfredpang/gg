package main

import (
	"fmt"
	"os"
)

func (cmd *ggcmd) cmdListcore() {
	for _, pkg := range cmd.corePkgs {
		fmt.Printf(pkg)
		fmt.Printf("\n")
	}
	os.Exit(0)
}
