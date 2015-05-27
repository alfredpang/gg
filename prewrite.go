//
// Originally from: github.com/dmitris/prewrite
//
// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
//
// Author: Dmitry Savintsev <dsavints@yahoo-inc.com>

// prewrite tool to rewrite import paths and package import comments for vendoring
// by adding or removing a given path prefix. The files are rewritten
// in-place with no backup (expectation is that version control is used), the output is gofmt'ed.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (cmd *ggcmd) astmodVendorWithPrefix(availPkgs map[string]*ggvPackage, prefix string, dir string, remove bool) error {
	cmd.astmodSpecialDirs = map[string]*string{}

	processor := cmd.astmodMakeVisitor(availPkgs, prefix, remove, false)
	_, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		return errors.New("Error - the traversal root " + dir + " does not exist, please double-check")
	}
	err = filepath.Walk(dir, processor)
	if err != nil {
		// add more info
		return err
	}
	return nil
}

// makeVisitor returns a rewriting function with parameters bound with a closure
func (cmd *ggcmd) astmodMakeVisitor(availPkgs map[string]*ggvPackage, prefix string, remove bool, verbose bool) filepath.WalkFunc {
	return func(path string, f os.FileInfo, err error) error {
		// check for previously seen special dirs
		for p, _ := range cmd.astmodSpecialDirs {
			if strings.HasPrefix(path, p) {
				return nil
			}
		}

		// check if this itself is a special dir
		if f.IsDir() {
			_, pfile := filepath.Split(path)

			// this is a dot directory
			// when usev, internal - don't recurse into internal
			if filepath.HasPrefix(pfile, ".") || (availPkgs != nil && pfile == "internal") {
				cmd.astmodSpecialDirs[path] = &path
				return nil
			}

			// has _ggv.json file
			ggvfile := filepath.Join(path, "_ggv.json")
			stat, err := os.Stat(ggvfile)
			if err == nil && !stat.IsDir() {
				cmd.astmodSpecialDirs[path] = &path
				return nil
			}

			return nil
		}

		if f.IsDir() || !strings.HasSuffix(f.Name(), ".go") {
			return nil
		}
		// special cases
		if cmd.astmodSkipFile(path) {
			return nil
		}
		src, err := ioutil.ReadFile(path)
		if err != nil {
			ggFatal("Fatal error reading file %s\n", path)
		}
		buf, err := cmd.astmodRewrite(path, src, availPkgs, prefix, remove)
		if err != nil {
			ggFatal("Fatal error rewriting AST, file %s - error: %s\n", path, err)
		}
		// check if there were any mods done for the file, return if non
		if buf == nil {
			return nil
		}
		err = ioutil.WriteFile(path, buf.Bytes(), f.Mode())
		if err != nil {
			ggFatal("Fatal error - unable to write to file %s: %s\n", path, err)
		}
		if verbose {
			fmt.Println(path)
		}
		return nil
	}
}

func (cmd *ggcmd) astmodSkipFile(fname string) bool {
	// known special cases
	skip := [...]string{
		"golang.org/x/tools/go/loader/testdata/badpkgdecl.go",
	}
	for _, s := range skip {
		if strings.HasSuffix(fname, s) {
			return true
		}
	}
	return false
}

// Rewrite modifies the AST to rewrite import statements and package import comments.
// src should be compatible with go/parser/#ParseFile:
// (The type of the argument for the src parameter must be string, []byte, or io.Reader.)
//
// return of nil, nil (no result, no error) means no changes are needed
func (cmd *ggcmd) astmodRewrite(fname string, src interface{}, availPkgs map[string]*ggvPackage, prefix string, remove bool) (buf *bytes.Buffer, err error) {
	gglog.Printf("fname=%s prefix=%s remove=%v\n", fname, prefix, remove)

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, fname, src, parser.ParseComments)
	if err != nil {
		log.Printf("Error parsing file %s, source: [%s], error: %s", fname, src, err)
		return nil, err
	}
	// normalize the prefix ending with a trailing slash
	if len(prefix) > 0 && prefix[len(prefix)-1] != '/' {
		prefix += "/"
	}

	changed, err := cmd.astmodRewriteImports(f, availPkgs, prefix, remove)
	if err != nil {
		log.Printf("Error rewriting imports in the AST: file %s - %s", fname, err)
		return nil, err
	}

	// when using vendoring no need to rewrite comments
	var changed2 bool = false
	if availPkgs == nil {
		changed2, err = cmd.astmodRewriteImportComments(f, fset, availPkgs, prefix, remove)
	}

	if err != nil {
		log.Printf("Error rewriting import comments in the AST: file %s - %s", fname, err)
		return nil, err
	}
	if !changed && !changed2 {
		return nil, nil
	}
	buf = &bytes.Buffer{}
	err = format.Node(buf, fset, f)
	return buf, err
}

// RewriteImports rewrites imports in the passed AST (in-place).
// It returns bool changed set to true if any changes were made
// and non-nil err on error
func (cmd *ggcmd) astmodRewriteImports(f *ast.File, availPkgs map[string]*ggvPackage, prefix string, remove bool) (changed bool, err error) {
	for _, impNode := range f.Imports {
		imp, err := strconv.Unquote(impNode.Path.Value)
		if err != nil {
			log.Printf("Error unquoting import value %v - %s\n", impNode.Path.Value, err)
			return false, err
		}
		// skip standard library imports and relative references
		if !strings.Contains(imp, ".") || strings.HasPrefix(imp, ".") {
			continue
		}

		if availPkgs != nil {
			if remove {
				if strings.HasPrefix(imp, prefix) {
					canonical := imp[len(prefix):]
					if availPkgs[canonical] != nil {
						impNode.Path.Value = strconv.Quote(imp[len(prefix):])
					}
				}
			} else {
				//gglog.Printf("  -> imp=%s availPkgs[imp]=%v\n", imp, availPkgs[imp])
				if availPkgs[imp] != nil {
					changed = true
					impNode.Path.Value = strconv.Quote(prefix + imp)
				}
			}
		} else {
			if remove {
				if strings.HasPrefix(imp, prefix) {
					changed = true
					impNode.Path.Value = strconv.Quote(imp[len(prefix):])
				}
			} else {
				// if import does not start with the prefix already, add it
				if !strings.HasPrefix(imp, prefix) {
					changed = true
					impNode.Path.Value = strconv.Quote(prefix + imp)
				}
			}
		}
	}
	return
}

// RewriteImportComments rewrites package import comments (https://golang.org/s/go14customimport)
func (cmd *ggcmd) astmodRewriteImportComments(f *ast.File, fset *token.FileSet, availPkg map[string]*ggvPackage, prefix string, remove bool) (changed bool, err error) {
	pkgpos := fset.Position(f.Package)
	// Print the AST.
	// ast.Print(fset, f)
	newcommentgroups := make([]*ast.CommentGroup, 0)
	for _, c := range f.Comments {
		commentpos := fset.Position(c.Pos())
		// keep the comment if we are not on the "package <X>" line
		// or the comment after the package statement does not look like import comment
		if commentpos.Line != pkgpos.Line ||
			!strings.HasPrefix(c.Text(), `import "`) {
			newcommentgroups = append(newcommentgroups, c)
			continue
		}
		parts := strings.Split(strings.Trim(c.Text(), "\n\r\t "), " ")
		oldimp, err := strconv.Unquote(parts[1])
		if err != nil {
			ggFatal("Error unquoting import value [%v] - %s\n", parts[1], err)
		}

		if remove {
			// the prefix is not there = nothing to remove, keep the comment
			if !strings.HasPrefix(oldimp, prefix) {
				newcommentgroups = append(newcommentgroups, c)
				continue
			}
		} else {
			// the prefix is already in the import path, keep the comment
			if strings.HasPrefix(oldimp, prefix) {
				newcommentgroups = append(newcommentgroups, c)
				continue
			}
		}
		newimp := ""
		if remove {
			newimp = oldimp[len(prefix):]
		} else {
			newimp = prefix + oldimp
		}
		changed = true
		c2 := ast.Comment{Slash: c.Pos(), Text: `// import ` + strconv.Quote(newimp)}
		cg := ast.CommentGroup{List: []*ast.Comment{&c2}}
		newcommentgroups = append(newcommentgroups, &cg)
	}
	// change the AST only if there are pending mods
	if changed {
		f.Comments = newcommentgroups
	}
	return changed, nil
}
