package main

import (
	"encoding/json"
	"errors"
	_ "fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type ggvPackage struct {
	LastUpdate     string // date-time of last update or touch
	Vcs            string // git, hg, manual for now
	VcsSource      string
	Revision       string
	Lock           bool // do not update on update
	RewriteImports bool // on update do import rewrites, or not
	ShallowUpdate  bool // do not recuse on go get dependencies
	SaveRepo       bool // keep copy of .git or .hg
	DepTests       bool // check dependencies of tests (when not shallow)
	Notes          string
}

// handling the vendor package file
type ggvJson struct {
	Version      string
	VendorPrefix string
	Packages     map[string]*ggvPackage // key is canonical pkg name
}

func (ggv *ggvJson) saveGvv(vendorFilename string) error {
	ggv.Version = "0.1" // force version
	b, err := json.MarshalIndent(ggv, "", "    ")
	if err != nil {
		ggFatal("Unable to marshal _ggv.json file. %s", err)
	}

	err = ioutil.WriteFile(vendorFilename, b, os.ModePerm)
	if err != nil {
		ggFatal("Unable to write %s. %s", vendorFilename, err)
	}
	return nil
}

func readGvvFromFile(filename string) (*ggvJson, error) {
	var ggv ggvJson
	file, err := os.Open(filename)
	if err != nil {
		ggFatal("%s", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		ggFatal("%s", err)
	}

	buffer := make([]byte, stat.Size())
	_, err = file.Read(buffer)
	if err != nil {
		ggFatal("%s", err)
	}

	err = json.Unmarshal(buffer, &ggv)
	if err != nil {
		ggFatal("%s", err)
	}

	return &ggv, nil
}

func resolveVendorConfigFilename(optVendor string, userSpecified bool) (string, *ggvJson, error) {

	gopath, err := getCurrentGopath()

	if userSpecified {
		if err != nil {
			ggFatal("Unable to get GOPATH %s", err)
		}

		fn := filepath.Join(gopath, "src", optVendor, "_ggv.json")
		statRet, err := os.Stat(fn)
		if err == nil && !statRet.IsDir() {
			ggv, err := readGvvFromFile(fn)
			return fn, ggv, err
		}

		return "", nil, errors.New("Unable to find _ggv.json at " + fn)
	}

	// try
	// current OR current/internal
	currentDir, err := os.Getwd()
	if err != nil {
		ggFatal("Unable to get current directory %s", err)
	}

	// is there a .gg file in current directory?
	currentDirGg := filepath.Join(currentDir, ".gg")
	statRet, err := os.Stat(currentDirGg)
	if err == nil && !statRet.IsDir() {
		// read the file and try to use that as the
		content, err := ioutil.ReadFile(currentDirGg)
		if err == nil {
			lines := strings.Split(string(content), "\n")
			if len(lines) >= 1 {
				// this
				fn := filepath.Join(gopath, "src", lines[0], "_ggv.json")
				statRet, err := os.Stat(fn)
				if err == nil && !statRet.IsDir() {
					ggv, err := readGvvFromFile(fn)
					return fn, ggv, err
				} else {
					// pretty bad, .gg specified vendor dir is no good
					ggFatal(".gg specifies vendor %s, but unable to read _ggv.json", lines[0])
				}
			}
		}
	}

	fn := currentDir + string(os.PathSeparator) + "_ggv.json"
	statRet, err = os.Stat(fn)
	if err == nil && !statRet.IsDir() {
		// found
		ggv, err := readGvvFromFile(fn)
		return fn, ggv, err
	}

	// current directory, or "internal"
	fn = currentDir + string(os.PathSeparator) + "internal" + string(os.PathSeparator) + "_ggv.json"
	statRet, err = os.Stat(fn)
	if err == nil && !statRet.IsDir() {
		// found
		ggv, err := readGvvFromFile(fn)
		return fn, ggv, err
	}

	return "", nil, errors.New("Unable to find _ggv.json at the default locations.")
}
