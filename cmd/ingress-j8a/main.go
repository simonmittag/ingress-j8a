package main

import (
	"flag"
	"fmt"
	"github.com/simonmittag/ingress-j8a/server"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Mode uint8

const (
	Server Mode = 1 << iota
	Version
	Usage
)

func main() {
	mode := Server
	go waitForSignal()

	//for Bootstrap
	defer recovery()

	v := flag.Bool("v", false, "print the server version")
	h := flag.Bool("h", false, "print usage instructions")
	flag.Usage = printUsage
	flag.Parse()
	if *v {
		mode = Version
	}
	if *h {
		mode = Usage
	}

	switch mode {
	case Server:
		server.
			NewServer().
			Bootstrap().
			Daemon()
	case Version:
		printVersion()
	case Usage:
		printUsage()
	}

}

func printUsage() {
	printVersion()
	flag.PrintDefaults()
}

func printVersion() {
	fmt.Printf("ingress-j8a[%s]\n", server.Version)
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func recovery() {
	if r := recover(); r != nil {
		pid := os.Getpid()
		fmt.Printf("pid %v, exiting...\n", pid)
		os.Exit(-1)
	}
}

func waitForSignal() {
	defer recovery()
	sig := interruptChannel()
	for {
		select {
		case <-sig:
			panic("os signal")
		default:
			time.Sleep(time.Second * 1)
		}
	}
}

func interruptChannel() chan os.Signal {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	return sigs
}
