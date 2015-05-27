package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (cmd *ggcmd) cmdVinit() {
	var optVendorRoot argOptionStr
	options := argOptions{}
	options.init("vinit")
	options.stringVar(&optVendorRoot, "v", "", "Vendor package root")
	options.stringVar(&optVendorRoot, "vendor", "", "Vendor package root")
	options.parse()

	var vendorPkg string = ""
	if optVendorRoot.IsSet {
		vendorPkg = optVendorRoot.String
	}

	var vfile string
	var err error
	gopath, err := getCurrentGopath()
	if err != nil {
		ggFatal("%s", err)
	}
	// $GOPATH/src
	gopathsrc := filepath.Join(gopath, "src")

	if vendorPkg == "" {
		// determine pkg from current directory
		vfile, err = os.Getwd()

		// get vendorPkg before modify current path
		// note we could be right at /src... i.e. empty vendorPkg root
		if vfile != gopathsrc {
			if !strings.HasPrefix(vfile, gopathsrc) {
				ggFatal("Unable to determine current package GOPATH=%s CurrentDir=%s", gopath, vfile)
			}
			vendorPkg = vfile[len(gopathsrc)+1:]
			if len(vendorPkg) > 0 {
				firstRune := strings.IndexRune(vendorPkg, 0)
				if firstRune == os.PathSeparator {
					vendorPkg = vendorPkg[1:]
				}
			}
		}

		vfile = filepath.Join(vfile, "_ggv.json")
	} else {
		vfile = filepath.Join(gopathsrc, vendorPkg, "_ggv.json")
	}

	fmt.Printf("Vendor Package: %s\n", vendorPkg)
	fmt.Printf("Vendor File: %s\n", vfile)
	_, err = os.Stat(vfile)

	if err == nil {
		ggFatal("Exiting with error. _ggv.json already exists at %s", vfile)
	}

	ggv := ggvJson{"", vendorPkg, map[string]*ggvPackage{}}
	err = ggv.saveGvv(vfile)
	if err != nil {
		ggFatal("Unable to write %s.", vfile)
	}

}
