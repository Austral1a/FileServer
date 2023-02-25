package dataServer

import (
	"fmt"
	"github.com/Austral1a/FileServer/src/core/controlServer"
	"net"
	"os"
)

func DoCommandLIST(ftpClientConn net.Conn) {
	dirEntry, err := os.ReadDir(controlServer.FileStorageLocalPath)
	if err != nil {
		// TODO: need to send proper code to FTP client
		fmt.Println("dir read error: ", err)
	}

	ftpClientConn.
}
