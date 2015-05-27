package main

import (
	"fmt"
	"sort"
)

func (cmd *ggcmd) cmdVlist() {
	var optVendorRoot argOptionStr
	options := argOptions{}
	options.init("vlist")
	options.stringVar(&optVendorRoot, "v", "", "Vendor package root")
	options.stringVar(&optVendorRoot, "vendor", "", "Vendor package root")
	options.parse()

	vendorFilename, currentGgv, err := resolveVendorConfigFilename(optVendorRoot.String, optVendorRoot.IsSet)
	_ = vendorFilename
	if err != nil {
		ggFatal("Unable to get vendor file %s", err)
	}

	var pprint []string
	for p, _ := range currentGgv.Packages {
		pprint = append(pprint, p)
	}

	sort.Strings(pprint)
	for _, p := range pprint {
		info := currentGgv.Packages[p]
		fmt.Printf("%s %s %s %s\n", p, info.Vcs, info.VcsSource, info.Revision)
	}
}
