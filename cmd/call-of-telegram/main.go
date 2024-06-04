package main

import (
	"github.com/volvinbur1/call-of-telegram/internal/cli"
)

func main() {
	rootCmd := cli.NewCmd()
	if err := cli.Execute(rootCmd); err != nil {
		panic(err)
	}
}
