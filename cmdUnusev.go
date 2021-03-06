package main

import (
	"os"
	"path/filepath"
)

func (cmd *ggcmd) cmdUnusev() {
	var optVendorRoot argOptionStr
	var optDir argOptionStr

	options := argOptions{}
	options.init("usev")
	options.stringVar(&optVendorRoot, "v", "", "Vendor package root")
	options.stringVar(&optVendorRoot, "vendor", "", "Vendor package root")

	options.parse()
	optPackages := options.args()

	vendorFilename, currentGgv, err := resolveVendorConfigFilename(optVendorRoot.String, optVendorRoot.IsSet)
	if err != nil {
		ggFatal("Unable to get vendor file %s", err)
	}
	vendorDir := filepath.Dir(vendorFilename)
	_ = vendorDir
	vendorRoot := currentGgv.VendorPrefix

	// we might not want to vend everything, but usually we do
	vendAvail := map[string]*ggvPackage{}
	if len(optPackages) == 0 {
		vendAvail = nil // when removing this means unconditional prefix removal
	} else {
		for _, p := range optPackages {
			vendAvail[p] = &ggvPackage{}
		}
	}

	var targetDir string
	if optDir.IsSet {
		targetDir = optDir.String
	} else {
		targetDir, err = os.Getwd()
	}

	gglog.Printf("vendAvail %v vendorRoot %s targetDir %s", vendAvail, vendorRoot, targetDir)
	err = cmd.astmodVendorWithPrefix(vendAvail, vendorRoot, targetDir, true)

	if err != nil {
		ggFatal("Error while doing import rewrites", err)
	}
}
