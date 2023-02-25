package dataServer

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/Austral1a/FileServer/src"
	"github.com/Austral1a/FileServer/src/command"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

type DataServer struct {
	ftpClientAddr net.Addr
	// TODO: Change to net.Addr ?
	controlServerAddr string
}

func NewDataServer() *DataServer {
	return &DataServer{controlServerAddr: ":2121"}
}

func (fs *DataServer) Start() {
	ln, err := net.Listen("tcp", ":20")
	if err != nil {
		fmt.Println(err)
	}

	fs.acceptConnection(ln)
}

func (fs *DataServer) acceptConnection(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept error: ", err)
		}

		go fs.handleConnection(conn)
	}
}

func (fs *DataServer) handleConnection(conn net.Conn) {
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
			fmt.Println(err)
			break
		}

		fmt.Printf("reveiced %d bytes over network\n", n)

		err = fs.saveFile(f, "storage")
		if err != nil {
			fmt.Println(err)
			break
		}

		fmt.Printf("File has been saved to file system\n")
	}
}

func (fs *DataServer) deserializeFile(conn net.Conn) (*src.File, error) {
	f := &src.File{}

	decoder := gob.NewDecoder(conn)
	err := decoder.Decode(f)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (fs *DataServer) saveFile(f *src.File, where string) error {
	newFile, err := os.Create(where + "/" + "_" + f.Name + "." + f.Extension)
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

func (fs *DataServer) handleCommand(msg string, conn net.Conn) error {
	slicedCommand := strings.Split(msg, " ") // command example: "USER Anonymous"
	cmd := strings.TrimSpace(slicedCommand[0])

	switch cmd {
	case command.LIST:

	}
}

// ftp passive mode connection
/*
func (fs *DataServer) newFTPClientAddr(newAddr net.Addr) {
	fs.ftpClientAddr = newAddr
}

func (fs *DataServer) connectToFTPClient() {

}*/
