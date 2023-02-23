package main

import (
	"github.com/Austral1a/FileServer/src/core"
	"github.com/Austral1a/FileServer/src/utils"
	"time"
)

func main() {
	go func() {
		time.Sleep(time.Millisecond * 300)
		utils.SendRealFile("googlechrome.dmg")
		utils.SendRealFile("vlc.dmg")
		utils.SendRealFile("vlc.dmg")
	}()
	server := &core.FileServer{}
	server.Start()
}
