package main

import (
	"fmt"
)

func (cmd *ggcmd) cmdLdep() {
	var optDepTests argOptionBool
	options := argOptions{}
	options.init("ldep")
	options.boolVar(&optDepTests, "dep-tests", true, "Also check dependencies of tests")
	options.parse()
	optPackages := options.args()

	// was a package specified
	// if not just do it in current directory
	var goListArg string = "."

	if len(optPackages) == 0 {
		goListArg = "."
	} else if len(optPackages) == 1 {
		goListArg = optPackages[0]
	} else {
		ggFatal("Please specify exactly one package to check.")
	}

	deps := cmd.ldepHelper(goListArg, optDepTests.Bool)
	for _, pkg := range deps {
		fmt.Printf("%s\n", pkg)
	}
}
