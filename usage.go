package main

// all the usage
func (cmd *ggcmd) initCommands() {
	cmd.commands = map[string]*action{

		"": {`gg - golang go get vendor manager
        
Manage vendor directory:

 vinit    Initialize vendor directory.
 vadd     Add package to vendor.
 vlist    List packages being vendored.
 vrebuild Rebuild from config file.
 vupdate  Update packages.

Import rewriting:

 usev     Rewrite to use import vendored packages.
 unusev   Rewrite to undo vendored imports.

Other commands:

 rdep     List dependencies of a go-getable package.
 ldep     List dependencies of local directory/package.
 listcore List known core packages.

Special files:

 _ggv.json Vendor configuration file.
 .gg       Specifies vendor root to use.

Use "gg help <command>" for usage of a specific command.
`, nil},

		// ---------------------------------------------------
		"help": {`Run "gg help"`, cmd.cmdHelp},

		// ---------------------------------------------------
		"vinit": {`gg vinit [options]

Initialize vendor directory.

    Create vendoring description file at current directory. The vendor package
    root will be derived from the current directory and GOPATH.

Options:

 -v --vendor VENDOR_ROOT Create vendor description file at specified
                         package directory under GOPATH.
`, cmd.cmdVinit},

		// ---------------------------------------------------
		"vadd": {`gg vadd [options] <gg-package> [<gg-packages ...]
        
Add package to vendor.

    For a go-gettable package(s), add to a vendor package directory. Will try
    current directory, "internal" in order to find vendor directory. Else
    specify with -v. If you need specific options for this package (lock,
    rewrite, notes), please vadd only one package.

Options:

 -v --vendor VENDOR_ROOT Vendor package root
 --vcs VCS               git, hg
 --vcs-source URL        Source of the package. https://github.com/a/b
 --revision REVISION     Revision, or latest if not specified.
 --lock=false            Lock on revision when adding done.
 --rewrite=true          Will bring in the package(s), but skip import rewrite.
 --shallow=false         Only get the gg-package without going recursively.
                         Default is recursive.
 --save-repo=false       Keep copy of .git or .hg in vendor directories.
 --dep-tests=true        Check for dependencies of tests as well.
 --notes NOTES           Add notes for package.
 --test=false            Dry run test.
`, cmd.cmdVadd},
		// ---------------------------------------------------
		"voption": {`gg voption [options] [<gg-package> ...]

Future, allow editing options via command line action.
Print or set options for vendored package.

    For vendored packages, you may change the update options. For instance, you
    may lock its revision or disable import rewriting. If you update the
    revision, consider using using vupdate as well to update that package.

Options:
 -v --vendor VENDOR_ROOT Vendor package root
 -a --all                If options specified, apply on all packages.
 --vcs VCS               git, hg, manual
 --vcs-source URL        Source of the package. https://github.com/a/b
 --revision REVISION     Revision, or latest if not specified.
 --lock=false            Lock on revision when adding done.
 --rewrite=true          Will bring in the package(s), but skip import rewrite.
 --notes NOTES           Add notes for package.
`, cmd.cmdVoption},
		// ---------------------------------------------------
		"vrm": {`gg vrm [options] <gg-package> [<gg-package> ...]

Future, allow deleting by command line.
Remove previously vendored package.

    Unlike vadd, this will not use recursive dependency to remove dependent
    packages.

Options:
 -v --vendor VENDOR_ROOT Vendor package root.
`, nil},
		// ---------------------------------------------------
		"vlist": {`gg vlist [options]

List managed packages.

Options:
 -v --vendor VENDOR_ROOT Vendor package root.
`, cmd.cmdVlist},
		// ---------------------------------------------------
		"vstrip": {`gg vstrip [options]

Future, allow command line helper to strip vendor directory.
Strip vendor directory.

    Remove everything that is rebuildable.

Options:
 -v --vendor VENDOR_ROOT Vendor package root.
`, nil},
		// ---------------------------------------------------
		"vrebuild": {`gg vrebuild [options]

Rebuild vendor directory.

    Given the configuration file _ggv.json in the vendor directory, strip and
    rebuild as described.

Options:
 -v --vendor VENDOR_ROOT Vendor package root.
`, cmd.cmdVrebuild},
		// ---------------------------------------------------
		"vupdate": {`gg vupdate [options] [<gg-package> ...]

Update previously added vendored packages to latest revision.

    By default it will recheck dependencies of packages.

Options:
 -v --vendor VENDOR_ROOT Vendor package root.
 --shallow=false      Skip rechecking of dependencies.
 --save-repo=false    Keep copy of .git or .hg in vendor directories.
 --dep-tests=true     Check for dependencies of tests as well.
 --revision REVISION  Update specified package to revision. (shallow)
 --test               See what would actually get updated without modifying
                      your vendor directory.
`, cmd.cmdVupdate},
		// ---------------------------------------------------
		"usev": {`gg usev [options] [<gg-package> ...]

Use previously vendored package.

    In the current directory, convert imports to vendored imports that
    currently exists. Import rewrites will not recurse into subdirectories with
    vendor configuration file or "internal" directories.

    Vendored directory is specified as --vendor, or as a ".gg" file in the
    directory. If neither is specified, it will look for the vendor directory
    under "internal".

Options:
 -v --vendor VENDOR_ROOT Vendor package root.
 -d --dir    DIR         Start at specified directory.
 `, cmd.cmdUsev},
		// ---------------------------------------------------
		"unusev": {`gg unusev [options] [<gg-package> ...]

Unvendor imports.

    This must be run with respect to a vendor package root. Vendored directory
    is specified as --vendor, or as a ".gg" file in the directory. If neither
    is specified, it will look for the vendor directory under "internal".

Options:
 -v --vendor VENDOR_ROOT Vendor package root.
 -d --dir    DIR         Start at specified directory.
`, cmd.cmdUnusev},
		// ---------------------------------------------------
		"irewrite": {`gg irewrite [options] <previous-import> <new-import>

Raw import rewrite.

    Starting at the current directory, perform import rewrites as described. It
    will not recurse into subdirectories with the vendor configuration file or
    "internal" directories. It will skip core packages during this import
    rewrite.

Options:
 -d --dir DIR Start import rewrite recursively down.
 --no-recurse Perform default or specified directory only.
 --force      Recursively do all .go files.

`, nil},
		// ---------------------------------------------------
		"rdep": {`gg rdep [options] <gg-package>

Check dependencies of a package in a public repo.

    Given a canonical package name, get dependencies as it currently exists on
    the internet. This is done by creating a clean temporary directory and
    running go get.

Options:

 --dep-tests=true  Check for dependencies of tests as well.
`, cmd.cmdRdep},
		// ---------------------------------------------------
		"ldep": {`gg ldep [options] [<local-package>]

Check dependencies of a package residing on local disk.

    For a local directory/package on your disk, determine the dependencies.
    Skip core packages and subpackages already found under the package being
    queried.

Options:

 --dep-tests=true  Check for dependencies of tests as well.

`, cmd.cmdLdep},
		// ---------------------------------------------------
		"pkgmeta": {`gg pkgmeta

Get meta for package.
`, cmd.cmdPkgmeta},
	}
}
