package main

//
// flag, but has bool to determine if option was specified or not
//

import (
	"flag"
	"os"
)

type argOptionStr struct {
	String string
	IsSet  bool // default is not specified
}

type argOptionBool struct {
	Bool  bool
	IsSet bool // default is not specified
}

type argOptions struct {
	FlagSet     *flag.FlagSet
	DebugOption argOptionBool
	IsSetMap    map[string]*bool
}

// helpers
func (options *argOptions) stringVar(option *argOptionStr, name string, value string, usage string) {
	options.FlagSet.StringVar(&option.String, name, value, usage)
	options.IsSetMap[name] = &option.IsSet
}

func (options *argOptions) boolVar(option *argOptionBool, name string, value bool, usage string) {
	options.FlagSet.BoolVar(&option.Bool, name, value, usage)
	options.IsSetMap[name] = &option.IsSet
}

func (options *argOptions) init(flagSetName string) {
	options.FlagSet = flag.NewFlagSet(flagSetName, flag.PanicOnError)
	options.IsSetMap = map[string]*bool{}
}

func (options *argOptions) parse() {
	// extra for debug
	options.boolVar(&options.DebugOption, "debug", false, "show debug messages")
	options.FlagSet.Parse(os.Args[2:])
	options.FlagSet.Visit(func(flag *flag.Flag) {
		*options.IsSetMap[flag.Name] = true
	})

	// turn on debug here if needed
	if options.DebugOption.Bool {
		gglogEnable(nil)
	}
}

func (options *argOptions) args() []string {
	return options.FlagSet.Args()
}
