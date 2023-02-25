package main

import (
	"github.com/Austral1a/FileServer/src/cli"
	"log"
)

func main() {
	err := cli.RunCLI()
	if err != nil {
		log.Fatalln("cli error: ", err)
	}
}
