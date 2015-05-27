package main

import (
	"fmt"
	"os"
)

func (cmd *ggcmd) cmdHelp() {
	// gg help
	if len(os.Args) == 2 {
		fmt.Printf(cmd.commands[""].usage)
		os.Exit(0)
	}

	// gg help <command>
	action := os.Args[2]
	helpAction := cmd.commands[action]

	if helpAction == nil {
		ggFatal("Command %s not understood for help.", action)
	}

	fmt.Printf(helpAction.usage)
	fmt.Printf("\n")
	os.Exit(0)
}
