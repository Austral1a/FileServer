package main

import (
	"github.com/Austral1a/FileServer/src/core/dataServer"
	"github.com/Austral1a/FileServer/src/utils"
	"time"
)

func main() {
	go func() {
		time.Sleep(time.Millisecond * 300)
		utils.SendRealFile("googlechrome.dmg")
		utils.SendRealFile("vlc.dmg")
	}()

	server := dataServer.NewDataServer()
	server.Start()
}
