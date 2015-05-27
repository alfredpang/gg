package main

import (
	"fmt"
	"path/filepath"
)

func (cmd *ggcmd) cmdVupdate() {
	var optVendorRoot argOptionStr
	var optShallow argOptionBool
	var optSaveRepo argOptionBool
	var optDepTests argOptionBool
	var optRevision argOptionStr
	var optTest argOptionBool

	options := argOptions{}
	options.init("vadd")
	options.stringVar(&optVendorRoot, "v", "", "Vendor package root")
	options.stringVar(&optVendorRoot, "vendor", "", "Vendor package root")
	options.boolVar(&optShallow, "shallow", false, "Get only pkg or recurse dependencies")
	options.boolVar(&optSaveRepo, "save-repo", false, "Keep copy of .hg or .git")
	options.boolVar(&optDepTests, "dep-tests", false, "Also check dependencies of tests")
	options.stringVar(&optRevision, "revision", "", "source control revision hash")
	options.boolVar(&optTest, "test", false, "Just test to see what will change.")

	options.parse()
	optPackages := options.args()

	if optRevision.IsSet {
		if len(optPackages) == 0 || len(optPackages) > 1 {
			ggFatal("When specifying a revision, you must specify exactly one package.")
		}

		// force shallow when 1 package & revision specified
		optShallow.Bool = true
		optShallow.IsSet = true
	}

	vendorFilename, currentGgv, err := resolveVendorConfigFilename(optVendorRoot.String, optVendorRoot.IsSet)
	if err != nil {
		ggFatal("Unable to get vendor file %s", err)
	}
	vendorDir := filepath.Dir(vendorFilename)
	vendorRoot := currentGgv.VendorPrefix

	// make sure specified package(s) exist
	for _, p := range optPackages {
		if currentGgv.Packages[p] == nil {
			ggFatal("Specified package %s does not exist. vadd it first.", p)
		}
	}

	todoPackages := map[string][]string{}
	if len(optPackages) == 0 {
		for p, _ := range currentGgv.Packages {
			optPackages = append(optPackages, p)
		}
		todoPackages = cmd.getMinimalPackagesList(optPackages, optShallow.Bool, true, currentGgv.Packages)
	} else {
		todoPackages = cmd.getMinimalPackagesList(optPackages, optShallow.Bool, true, currentGgv.Packages)
	}

	gglog.Printf("todoPackages: %v\n", todoPackages)

	updatedPackages := map[string]*ggvPackage{}

	for pkgName, goGetInfo := range todoPackages {
		var currentPackageInfo *ggvPackage = currentGgv.Packages[pkgName]
		var newPackageInfo *ggvPackage = &ggvPackage{LastUpdate: getNowStr()}

		if currentPackageInfo == nil {
			// new package
			newPackageInfo.Vcs = goGetInfo[1]
			newPackageInfo.VcsSource = goGetInfo[2]
			newPackageInfo.Revision = ""
			newPackageInfo.Lock = false
			newPackageInfo.RewriteImports = true
			newPackageInfo.ShallowUpdate = optShallow.Bool
			newPackageInfo.SaveRepo = optSaveRepo.Bool
			newPackageInfo.DepTests = optDepTests.Bool

			newPackageInfo.Notes = ""

			if newPackageInfo.Vcs == "" {
				newPackageInfo.Vcs = goGetInfo[1]
			}

			if newPackageInfo.VcsSource == "" {
				newPackageInfo.VcsSource = goGetInfo[2]
			}
		} else {
			// existing package; may get updated as side effect
			newPackageInfo.LastUpdate = currentPackageInfo.LastUpdate
			newPackageInfo.Vcs = currentPackageInfo.Vcs
			newPackageInfo.VcsSource = currentPackageInfo.VcsSource
			if currentPackageInfo.Lock {
				newPackageInfo.Revision = currentPackageInfo.Revision
			} else {
				newPackageInfo.Revision = ""
			}
			newPackageInfo.Lock = currentPackageInfo.Lock
			newPackageInfo.RewriteImports = currentPackageInfo.RewriteImports
			newPackageInfo.ShallowUpdate = currentPackageInfo.ShallowUpdate
			newPackageInfo.SaveRepo = currentPackageInfo.SaveRepo
			newPackageInfo.DepTests = currentPackageInfo.DepTests
			newPackageInfo.Notes = currentPackageInfo.Notes
		}

		// skip manual packages
		if newPackageInfo.Vcs != "manual" {
			updatedPackages[pkgName] = newPackageInfo
		}
	}

	err = cmd.downloadUpdate(vendorDir, vendorRoot, updatedPackages, optTest.Bool)

	for pkg, pkgInfo := range updatedPackages {
		var oldInfo *ggvPackage = currentGgv.Packages[pkg]
		if oldInfo == nil {
			fmt.Printf("Added %s - %s %s - %s\n", pkg, pkgInfo.Vcs, pkgInfo.VcsSource, pkgInfo.Revision)
		} else {
			if pkgInfo.Revision != oldInfo.Revision {
				fmt.Printf("Updated %s - %s %s - %s to %s\n", pkg, pkgInfo.Vcs, pkgInfo.VcsSource, oldInfo.Revision, pkgInfo.Revision)
			}
		}
	}

	if optTest.Bool {
		fmt.Printf("Dry run. Exiting with no errors.\n")
		return
	}

	// update new ggv
	for pkgName, newPkgInfo := range updatedPackages {
		currentGgv.Packages[pkgName] = newPkgInfo
	}

	err = currentGgv.saveGvv(vendorFilename)
	if err != nil {
		ggFatal("%s", err)
	}
}
