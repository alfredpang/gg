package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (cmd *ggcmd) cmdVadd() {
	var optVendorRoot argOptionStr
	var optVcs argOptionStr
	var optVcsSource argOptionStr
	var optRevision argOptionStr
	var optLock argOptionBool
	var optRewrite argOptionBool
	var optShallow argOptionBool
	var optSaveRepo argOptionBool
	var optDepTests argOptionBool
	var optNotes argOptionStr
	var optTest argOptionBool
	var optPackages []string

	options := argOptions{}
	options.init("vadd")
	options.stringVar(&optVendorRoot, "v", "", "Vendor package root")
	options.stringVar(&optVendorRoot, "vendor", "", "Vendor package root")
	options.stringVar(&optVcs, "vcs", "", "git, hg, manual")
	options.stringVar(&optVcsSource, "vcs-source", "", "e.g. https://github.com/aaa/bb")
	options.stringVar(&optRevision, "revision", "", "source control revision hash")
	options.boolVar(&optLock, "lock", false, "Lock on revision")
	options.boolVar(&optRewrite, "rewrite", true, "Rewrite imports on vendored package")
	options.boolVar(&optShallow, "shallow", false, "Get only pkg or recurse dependencies")
	options.boolVar(&optSaveRepo, "save-repo", false, "Keep copy of .hg or .git")
	options.boolVar(&optDepTests, "dep-tests", false, "Also check dependencies of tests")
	options.stringVar(&optNotes, "notes", "", "Additional notes")
	options.boolVar(&optTest, "test", false, "Just test to see what will change.")
	options.parse()
	optPackages = options.args()

	if len(optPackages) <= 0 {
		ggFatal("Please specify at least one canonical package to add.")
	}

	if len(optPackages) > 1 {
		if optVcs.IsSet || optVcsSource.IsSet || optRevision.IsSet {
			ggFatal("When specifying more than one package, --vcs, --vcs-source, --revision may not be specified.")
		}
	}

	vendorFilename, currentGgv, err := resolveVendorConfigFilename(optVendorRoot.String, optVendorRoot.IsSet)
	if err != nil {
		ggFatal("Unable to get vendor file %s", err)
	}
	vendorDir := filepath.Dir(vendorFilename)
	vendorRoot := currentGgv.VendorPrefix

	// check if we have this one already
	for _, p := range optPackages {
		var existingPackage *ggvPackage = nil

		// does any of the existing packages handle this?
		for pkgName, pkgInfo := range currentGgv.Packages {
			if strings.HasPrefix(p, pkgName) {
				existingPackage = pkgInfo
				break
			}
		}

		if existingPackage != nil {
			// don't add if specified package already exists
			ggFatal("Exiting. Package %s already exists.", p)
		}
	}

	todoPackages := map[string][]string{}

	// if we are doing single package and specifing vcs, then do it
	if optVcs.IsSet && optVcsSource.IsSet && len(optPackages) == 1 {
		todoPackages[optPackages[0]] = []string{optPackages[0], optVcs.String, optVcsSource.String}
	} else {
		todoPackages = cmd.getMinimalPackagesList(optPackages, optShallow.Bool, optDepTests.Bool, nil)
	}

	gglog.Printf("Affected packages:\n")
	for pkgName, goGetInfo := range todoPackages {
		gglog.Printf("%s %v\n", pkgName, goGetInfo)
	}

	updatedPackages := map[string]*ggvPackage{}
	for pkgName, goGetInfo := range todoPackages {
		var currentPackageInfo *ggvPackage = currentGgv.Packages[pkgName]
		var newPackageInfo *ggvPackage = &ggvPackage{LastUpdate: getNowStr()}

		if currentPackageInfo == nil {
			// new package
			newPackageInfo.Vcs = optVcs.String
			newPackageInfo.VcsSource = optVcsSource.String
			newPackageInfo.Revision = optRevision.String
			newPackageInfo.Lock = optLock.Bool
			newPackageInfo.RewriteImports = optRewrite.Bool
			newPackageInfo.ShallowUpdate = optShallow.Bool
			newPackageInfo.SaveRepo = optSaveRepo.Bool
			newPackageInfo.DepTests = optDepTests.Bool
			newPackageInfo.Notes = optNotes.String

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
