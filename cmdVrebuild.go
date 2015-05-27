package main

import (
	"path/filepath"
)

func (cmd *ggcmd) cmdVrebuild() {
	var optVendorRoot argOptionStr
	options := argOptions{}
	options.init("vrebuild")
	options.stringVar(&optVendorRoot, "v", "", "Vendor package root")
	options.parse()

	// maybe in future allow rebuilding of specific packages
	// but just do it all for now
	optPackages := options.args()

	if len(optPackages) > 0 {
		ggFatal("We only do full rebuilds for now and do not allow specifying specific packages")
	}

	// see if we can resolve the vendor config file, then load it up
	vendorFilename, currentGgv, err := resolveVendorConfigFilename(optVendorRoot.String, optVendorRoot.IsSet)
	if err != nil {
		ggFatal("Unable to get vendor file %s", err)
	}
	vendorDir := filepath.Dir(vendorFilename)
	vendorRoot := currentGgv.VendorPrefix

	updatedPackages := map[string]*ggvPackage{}
	for pkgName, currentPackageInfo := range currentGgv.Packages {
		if currentPackageInfo.Vcs == "manual" {
			continue
		}

		var newPackageInfo *ggvPackage = &ggvPackage{LastUpdate: getNowStr()}
		newPackageInfo.LastUpdate = currentPackageInfo.LastUpdate
		newPackageInfo.Vcs = currentPackageInfo.Vcs
		newPackageInfo.VcsSource = currentPackageInfo.VcsSource
		newPackageInfo.Revision = currentPackageInfo.Revision
		newPackageInfo.Lock = currentPackageInfo.Lock
		newPackageInfo.RewriteImports = currentPackageInfo.RewriteImports
		newPackageInfo.ShallowUpdate = currentPackageInfo.ShallowUpdate
		newPackageInfo.SaveRepo = currentPackageInfo.SaveRepo
		newPackageInfo.DepTests = currentPackageInfo.DepTests
		newPackageInfo.Notes = currentPackageInfo.Notes

		updatedPackages[pkgName] = newPackageInfo
	}

	err = cmd.downloadUpdate(vendorDir, vendorRoot, updatedPackages, false)
	// _ggv.json stays the same of course
}
