package main

import (
	"fmt"
	"os"
)

// on bad errors, just log.Fatal

type action struct {
	usage  string
	helper func()
}

// command handling
type ggcmd struct {
	commands map[string]*action

	// track seen vendored or internal directories
	astmodSpecialDirs map[string]*string
}

// print out stderr "ERROR: <message>", exit
func ggFatal(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", a...)
	os.Exit(1)
}

func (cmd *ggcmd) init() {
	gglogDisable()
	cmd.initCommands()
}

func (cmd *ggcmd) run() {
	doAction := cmd.getCommand()
	if doAction == nil {
		ggFatal("No known command found.")
	}

	doAction.helper()
	os.Exit(0)
}

func (cmd *ggcmd) getCommand() (doAction *action) {
	if len(os.Args) <= 1 {
		fmt.Printf(cmd.commands[""].usage)
		os.Exit(0)
	}

	if doAction = cmd.commands[os.Args[1]]; doAction == nil {
		ggFatal("Command %s not understood.", os.Args[1])
	}

	return doAction
}
