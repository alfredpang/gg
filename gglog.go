package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

var (
	gglog *log.Logger
)

func gglogEnable(out io.Writer) {
	if out == nil {
		gglog = log.New(os.Stderr, "DEBUG: ", log.LstdFlags|log.Lshortfile)
	} else {
		gglog = log.New(out, "DEBUG: ", log.LstdFlags|log.Lshortfile)
	}
}

func gglogDisable() {
	gglog = log.New(ioutil.Discard, "", 0)
}
