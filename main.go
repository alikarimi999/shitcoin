package main

import (
	"os"

	"github.com/alikarimi999/shitcoin/cli"
)

func main() {

	defer os.Exit(0)
	cmd := new(cli.Commandline)

	cmd.Run()
}
