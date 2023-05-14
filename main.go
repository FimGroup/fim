package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"esbconcept/sample"
)

//go:embed version
var ver string

func main() {
	fmt.Println("version:", strings.TrimSpace(string(ver)))

	if err := sample.StartForum(); err != nil {
		panic(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT)
	_ = <-c
	fmt.Println("service exit!")
}
