package main

import (
	"github.com/Austral1a/FileServer/src/ftpServer"
)

func main() {
	ftpServer := new(ftpserver.FTPServer).NewFTPServer()

	if ftpServer.Cs.IsStarted {
		ftpServer.HandleCommands()
	}
}
