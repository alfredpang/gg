package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// for each package, get canonical parent package
// note includeTestDeps is ignored if knownPkgs is available
func (cmd *ggcmd) getMinimalPackagesList(pkgs []string, shallow bool, includeTestDeps bool, knownPkgs map[string]*ggvPackage) map[string][]string {

	gglog.Printf("len(pkgs)=%d shallow=%v includeTestDeps=%v len(knownPkgs)=%d\n", len(pkgs), shallow, includeTestDeps, len(knownPkgs))

	todoPackages := map[string][]string{}
	var pkg string
	var vcs string
	var vcsSource string
	var err error

	for _, p := range pkgs {

		var knownPkgInfo *ggvPackage = nil

		// for known packages, skip this meta getting
		if knownPkgs == nil {
			pkg, vcs, vcsSource, err = getPkgMeta(p)
		} else {
			knownPkgInfo = knownPkgs[p]
			if knownPkgInfo != nil {
				pkg, vcs, vcsSource, err = p, knownPkgInfo.Vcs, knownPkgInfo.VcsSource, nil
			}
		}

		if err != nil {
			fmt.Printf("Unable to resolve a package: %s\n", p)
			continue
		}

		if todoPackages[pkg] == nil {
			todoPackages[pkg] = []string{pkg, vcs, vcsSource}
		}

		if shallow {
			continue
		}

		recursePkgs := cmd.rdepHelper(p, includeTestDeps)
		for _, pp := range recursePkgs {
			pkg, vcs, vcsSource, err := getPkgMeta(pp)
			if err == nil {
				if todoPackages[pkg] == nil {
					todoPackages[pkg] = []string{pkg, vcs, vcsSource}
				}
			} else {
				fmt.Printf("Unable to resolve a package: %s\n", pp)
				continue
			}
		}
	}
	return todoPackages
}

// given package -> ggvPackageinfo, download and update

func (cmd *ggcmd) downloadUpdate(vendorDir string, vendorRoot string, updatedPackages map[string]*ggvPackage, dryrun bool) error {
	var err error

	// map pkgname to tempdirs that contain fresh downloads
	downloadedDirs := map[string]struct {
		TempDir string
		DestDir string
	}{}

	// download to temp directories
	for pkgName, newPkgInfo := range updatedPackages {
		gglog.Printf("%s %v\n", pkgName, newPkgInfo)

		// revision is set in newPkgInfo anyways...
		tempDir, destDir, _, err := cmd.downloadPkg(vendorDir, vendorRoot, pkgName, newPkgInfo, true)
		if err != nil {
			ggFatal("%s", err)
		}
		downloadedDirs[pkgName] = struct {
			TempDir string
			DestDir string
		}{tempDir, destDir}
	}

	gglog.Printf("%v\n", downloadedDirs)

	if dryrun {
		// remove all temp dirs
		for _, dirMove := range downloadedDirs {
			err := os.RemoveAll(dirMove.TempDir)
			if err != nil {
				ggFatal("Unable to remove dest directory %s", dirMove.DestDir)
			}
		}
		return nil
	}

	// copy it all over
	for _, dirMove := range downloadedDirs {
		// allow to fail this quietly, esp if the dir does not exist
		err = os.RemoveAll(dirMove.DestDir)
		if err != nil {
			ggFatal("Unable to remove dest directory %s", dirMove.DestDir)
		}

		err = os.MkdirAll(dirMove.DestDir, os.ModePerm)
		if err != nil {
			ggFatal("Unable to make target directory %s", dirMove.DestDir)
		}
		err = os.Rename(dirMove.TempDir, dirMove.DestDir)
		if err != nil {
			ggFatal("Unable to move from %s to %s", dirMove.TempDir, dirMove.DestDir)
		}
	}

	return nil
}

// tempdir, destdir, revision, err
// this will handle import rewrites here if needed
func (cmd *ggcmd) downloadPkg(vendorDir string, vendorRoot string, p string, info *ggvPackage, dryrun bool) (string, string, string, error) {
	targetDir := filepath.Join(vendorDir, p)

	if info.Vcs == "manual" {
		return "", targetDir, "", errors.New("Manual can't be downloaded") // nothing to do, but tryt
	}

	if info.Vcs == "" || info.VcsSource == "" {
		return "", targetDir, "", errors.New("Unknown vcs, or vcs source")
	}

	// if revision is "", then latest
	tempDir, revision, err := cmd.fetchPackage(info.Vcs, info.VcsSource, info.Revision, info.SaveRepo)
	gglog.Printf("%s %s %s %v\n", p, tempDir, revision, err)

	info.Revision = revision

	if info.RewriteImports {
		err = cmd.astmodVendorWithPrefix(nil, vendorRoot, tempDir, false)
		if err != nil {
			// pretty bad
			ggFatal("Unable to do import rewrite for package %s at %s", p, tempDir)
		}
	}

	return tempDir, targetDir, revision, nil
}

// tempdir, revision fetched, error
func (cmd *ggcmd) fetchPackage(vcs string, vcsSource string, revision string, saveRepo bool) (string, string, error) {
	if vcs == "git" {
		return cmd.fetchPackageGit(vcsSource, revision, saveRepo)
	} else if vcs == "hg" {
		return cmd.fetchPackageHg(vcsSource, revision, saveRepo)
	}

	return "", "", errors.New("Unknown vcs specified " + vcs)
}

// tempdir, revision fetched, error
func (cmd *ggcmd) fetchPackageGit(vcsSource string, revision string, saveRepo bool) (string, string, error) {
	gglog.Printf("fetchPackageGit vcsSource=%s revision=%s saveRepo=%v\n", vcsSource, revision, saveRepo)

	tempdir, err := ioutil.TempDir("", "gg")
	if err != nil {
		ggFatal("Unable to create temp directory %s", err)
	}

	subcmd := exec.Command("git", "clone", vcsSource, tempdir)
	err = subcmd.Run()
	if err != nil {
		ggFatal("Unable to git clone %s %s", vcsSource, err)
	}

	if revision != "" {
		// check out specified revision
		subcmd = exec.Command("git", "checkout", revision)
		subcmd.Dir = tempdir
		err = subcmd.Run()
		if err != nil {
			ggFatal("Unable to git checkout %s %s %s", vcsSource, revision, err)
		}
	}

	var revisionRaw []byte = nil
	subcmd = exec.Command("git", "log", "--pretty=format:%H", "-n", "1")
	subcmd.Dir = tempdir
	revisionRaw, err = subcmd.Output()
	if err != nil {
		ggFatal("Unable to git log %s %s", vcsSource, err)
	}

	if !saveRepo {
		repoDir := filepath.Join(tempdir, ".git")
		err := os.RemoveAll(repoDir)
		if err != nil {
			ggFatal("Unable to remove %s", repoDir)
		}
	}

	return tempdir, strings.TrimSpace(string(revisionRaw)), nil
}

// tempdir, revision fetched, error
func (cmd *ggcmd) fetchPackageHg(vcsSource string, revision string, saveRepo bool) (string, string, error) {
	gglog.Printf("fetchPackageHg vcsSource=%s revision=%s saveRepo=%v\n", vcsSource, revision, saveRepo)

	tempdir, err := ioutil.TempDir("", "gg")
	if err != nil {
		ggFatal("Unable to create temp directory %s", err)
	}

	subcmd := exec.Command("hg", "clone", vcsSource, tempdir)
	err = subcmd.Run()
	if err != nil {
		ggFatal("Unable to hg clone %s %s", vcsSource, err)
	}

	if revision != "" {
		subcmd = exec.Command("hg", "update", "-r", revision)
		subcmd.Dir = tempdir
		err = subcmd.Run()
		if err != nil {
			ggFatal("Unable to hg update %s %s", vcsSource, err)
		}
	}

	var revisionRaw []byte = nil
	subcmd = exec.Command("hg", "identify", "--id")
	subcmd.Dir = tempdir
	revisionRaw, err = subcmd.Output()
	if err != nil {
		ggFatal("Unable to hg identify %s %s", vcsSource, err)
	}

	if !saveRepo {
		repoDir := filepath.Join(tempdir, ".hg")
		err := os.RemoveAll(repoDir)
		if err != nil {
			ggFatal("Unable to remove %s", repoDir)
		}
	}
	return tempdir, strings.TrimSpace(string(revisionRaw)), nil
}

// package, vcs, vcssource, error
func getPkgMeta(p string) (string, string, string, error) {
	// curl the package as a url
	// parse and look for meta

	reMeta := regexp.MustCompile("<meta [^>]*>")
	reContent := regexp.MustCompile("content=\"([^\"]*)\"")
	reName := regexp.MustCompile("name=\"([^\"]*)\"")

	for len(p) > 0 {
		purl := "https://" + p + "?go-get=1"
		resp, err := http.Get(purl)

		if err != nil {
			// maybe http instead?
			purl = "http://" + p + "?go-get=1"
			resp, err = http.Get(purl)
		}

		if err != nil {
			// chop p, and continue
			pparts := strings.Split(p, "/")
			p = strings.Join(pparts[:len(pparts)-1], "/")
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue // pretty bad error
		}
		//fmt.Printf("BODY: %s\n", string(body))
		metas := reMeta.FindAllString(string(body), -1)

		// look for the meta e.g.
		// <meta name="go-import" content="gopkg.in/mgo.v2 git https://gopkg.in/mgo.v2">
		for _, meta := range metas {
			//fmt.Printf("META: %s\n", meta)
			mname := reName.FindAllStringSubmatch(meta, -1)
			if mname == nil || mname[0][1] != "go-import" {
				continue
			}

			// found it
			mcontent := reContent.FindAllStringSubmatch(meta, -1)
			if mcontent == nil {
				continue
			}
			mcontentParts := strings.Split(mcontent[0][1], " ")
			if len(mcontentParts) != 3 {
				continue
			}

			return mcontentParts[0], mcontentParts[1], mcontentParts[2], nil
		}

		pparts := strings.Split(p, "/")
		p = strings.Join(pparts[:len(pparts)-1], "/")
	}

	return "", "", "", errors.New("Unable to get package meta info")
}
