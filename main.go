package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/Austral1a/FileServer/src"
	"github.com/Austral1a/FileServer/src/utils"
	"io"
	"log"
	"net"
	"os"
	"time"
)

type FileServer struct{}

func (fs *FileServer) start() {
	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		fmt.Println(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept error: ", err)
		}

		go fs.handleConnection(conn)
	}
}

func (fs *FileServer) handleConnection(conn net.Conn) {
	buf := new(bytes.Buffer)
	defer conn.Close()

	for {
		f, err := fs.deserializeFile(conn)
		if err != nil {
			fmt.Println("deserialization error: ", err)
			break
		}

		n, err := io.CopyN(buf, conn, int64(len(f.Bytes)))
		if err != nil && err != io.EOF {
			fmt.Println("IN COPY")
			fmt.Println(err)
			break
		}

		fmt.Printf("reveiced %d bytes over network\n", n)

		err = fs.saveFile(f)
		if err != nil {
			fmt.Println(err)
			break
		}

		fmt.Printf("File has been saved to file system\n")
	}
}

func (fs *FileServer) deserializeFile(conn net.Conn) (*src.File, error) {
	f := &src.File{}

	decoder := gob.NewDecoder(conn)
	err := decoder.Decode(f)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (fs *FileServer) saveFile(f *src.File) error {
	newFile, err := os.Create("_" + f.Name + "." + f.Extension)
	defer func() {
		err := newFile.Close()
		if err != nil {
			log.Println("Failed to close file: ", err)
		}
	}()
	if err != nil {
		return err
	}

	_, err = newFile.Write(f.Bytes)
	if err != nil {
		return err
	}

	return nil
}

func main() {

	go func() {
		time.Sleep(time.Millisecond * 300)
		utils.SendRealFile("googlechrome.dmg")
		utils.SendRealFile("vlc.dmg")
		utils.SendRealFile("vlc.dmg")
	}()
	server := &FileServer{}
	server.start()
}
