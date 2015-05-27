package main

import (
	"fmt"
	"os"
)

func (cmd *ggcmd) cmdRdep() {
	var optDepTests argOptionBool
	options := argOptions{}
	options.init("rdep")
	options.boolVar(&optDepTests, "dep-tests", true, "Also check dependencies of tests")
	options.parse()
	optPackages := options.args()

	if len(optPackages) != 1 {
		ggFatal("Please specify exactly one go-gettable package.")
	}

	deps := cmd.rdepHelper(os.Args[2], optDepTests.Bool)
	for _, pkg := range deps {
		fmt.Printf("%s\n", pkg)
	}
}
