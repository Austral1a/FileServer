package main

import (
	"github.com/Austral1a/FileServer/src/controlServer"
)

func main() {
	cs := controlServer.NewControlServer()
	cs.Start()
}
