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

func main() {
	go waitForSignal()

	//for Bootstrap
	defer recovery()

	flag.Usage = func() {
		fmt.Printf(`ingress-j8a[%s]"`, server.Version)
		fmt.Print("\n")
		flag.PrintDefaults()
	}

	server.Bootstrap()
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
