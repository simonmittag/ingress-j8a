package main

import (
	"flag"
	"os"
	"testing"
)

func TestMainFuncWithHelp(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append([]string{"-h"}, "-h")
	main()
}

func TestMainFuncWithVersion(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append([]string{"-v"}, "-v")
	main()
}
