package main

import (
	"fmt"
	"os"
)

func (cmd *ggcmd) cmdPkgmeta() {
	p := os.Args[2]
	pkg, vcs, vcsSource, err := getPkgMeta(p)
	fmt.Printf("getPkgMeta(%s) = %s, %s, %s, %v\n", p, pkg, vcs, vcsSource, err)
}
