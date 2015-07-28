package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// other useful things

// go list -e -json output
type goListJson struct {
	Dir         string
	ImportPath  string
	Name        string
	Doc         string
	Target      string
	Root        string
	Gofiles     []string
	Imports     []string
	Deps        []string
	TestGoFiles []string
	TestImports []string
}

func getNowStr() string {
	const layout = "2006-01-02T15:04:05"
	t := time.Now()
	return t.Format(layout)
}

// github.com/jprobinson/go-imap/imap
// actually lives in the github.com/jprobinson/go-imap repo
// only works on github.com and bitbucket.org for now
//
// For other things, you will need to specify the dependencies individually
//
func trimPackageToRepo(p string) string {
	if !strings.HasPrefix(p, "github.com") && !strings.HasPrefix(p, "bitbucket.org") {
		return p
	}

	pparts := strings.Split(p, "/")
	if len(pparts) <= 3 {
		return p
	}

	return strings.Join(pparts[:3], "/")
}

func getCurrentGopath() (string, error) {
	currentenv := os.Environ()
	for _, envval := range currentenv {
		if strings.HasPrefix(envval, "GOPATH=") {
			return string(envval[7:]), nil
		}
	}
	return "", errors.New("GOPATH not found")
}

func getEnvWithNewGopath(newGopath string) []string {
	currentenv := os.Environ()
	// same env, but change GOPATH
	subenv := make([]string, len(currentenv))
	for i, envval := range currentenv {
		if strings.HasPrefix(envval, "GOPATH=") {
			subenv[i] = fmt.Sprintf("GOPATH=%s", newGopath)
		} else {
			subenv[i] = envval
		}
	}

	return subenv
}

// package/repo dependencies
func (cmd *ggcmd) rdepHelper(rpkg string, includeTestDeps bool) []string {
	tempdir, err := ioutil.TempDir("", "gg")
	if err != nil {
		ggFatal("Unable to create temp directory %s", err)
	}
	defer os.RemoveAll(tempdir)
	os.Mkdir(tempdir+"/src", os.ModePerm)

	subcmd := exec.Command("go", "get", rpkg)
	subcmd.Env = getEnvWithNewGopath(tempdir)
	_, err = subcmd.Output()
	if err != nil {
		ggFatal("Unable to get package specified err=%s", err)
	}

	subcmd = exec.Command("go", "list", "-e", "-json", rpkg)
	subcmd.Env = getEnvWithNewGopath(tempdir)
	goListJsonRaw, err := subcmd.Output()
	if err != nil {
		ggFatal("Unable call go list on package err=%s", err)
	}

	var pkgGoList goListJson
	err = json.Unmarshal(goListJsonRaw, &pkgGoList)
	if err != nil {
		ggFatal("Unable to parse json of go list err=%s", err)
	}

	hasSeen := map[string]string{}
	deps := make([]string, 0, len(pkgGoList.Deps)+len(pkgGoList.TestImports))
	for _, pkg := range pkgGoList.Deps {
		if !cmd.isCorePackage(pkg) && !strings.HasPrefix(pkg, pkgGoList.ImportPath) {
			if hasSeen[pkg] != "" {
				continue
			}
			deps = append(deps, pkg)
			hasSeen[pkg] = pkg
		}
	}

	if !includeTestDeps {
		sort.Strings(deps)
		return deps
	}

	for _, pkg := range pkgGoList.TestImports {
		if !cmd.isCorePackage(pkg) && !strings.HasPrefix(pkg, pkgGoList.ImportPath) {
			if hasSeen[pkg] != "" {
				continue
			}
			deps = append(deps, pkg)
			hasSeen[pkg] = pkg
		}
	}
	sort.Strings(deps)
	return deps
}

func (cmd *ggcmd) ldepHelper(goListArg string, includeTestDeps bool) []string {
	subcmd := exec.Command("go", "list", "-e", "-json", goListArg)
	goListJsonRaw, err := subcmd.Output()
	if err != nil {
		ggFatal("Unable call go list on package err=%s", err)
	}

	var pkgGoList goListJson
	err = json.Unmarshal(goListJsonRaw, &pkgGoList)
	if err != nil {
		ggFatal("Unable to parse json of go list err=%s", err)
	}

	hasSeen := map[string]string{}
	deps := make([]string, 0, len(pkgGoList.Deps)+len(pkgGoList.TestImports))
	for _, pkg := range pkgGoList.Deps {
		if !cmd.isCorePackage(pkg) && !strings.HasPrefix(pkg, pkgGoList.ImportPath) {
			deps = append(deps, pkg)
			hasSeen[pkg] = pkg
		}
	}

	if !includeTestDeps {
		sort.Strings(deps)
		return deps
	}

	for _, pkg := range pkgGoList.TestImports {
		if !cmd.isCorePackage(pkg) && !strings.HasPrefix(pkg, pkgGoList.ImportPath) {
			if hasSeen[pkg] != "" {
				continue
			}
			deps = append(deps, pkg)
			hasSeen[pkg] = pkg
		}
	}

	sort.Strings(deps)
	return deps
}

// simple naive way to see if it is an internal package by checking
// first part of package path and seeing if it is something like github.com or
// bitbucket.org; if it has a domain name, it's probably not an internal pkg
func (cmd ggcmd) isCorePackage(name string) bool {
	if !strings.Contains(name, "/") {
		return true
	}

	pkgParts := strings.Split(name, "/")
	return !strings.Contains(pkgParts[0], ".")
}
