package dataServer

import (
	"fmt"
	"github.com/Austral1a/FileServer/src/controlServer"
	"os"
)

func DoCommandLIST(ds *DataServer) {
	dirEntry, err := os.ReadDir(controlServer.FileStorageLocalPath)
	if err != nil {
		// TODO: need to send proper code to FTP client
		fmt.Println("dir read error: ", err)
	}

	// say to CS that
	ds.controlServerConn
}

func InformControlServerAboutCommandProcessing() {

}
