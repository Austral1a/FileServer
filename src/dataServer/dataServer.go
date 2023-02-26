package dataServer

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/Austral1a/FileServer/src/command"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

type DataServer struct {
}

func NewDataServer() *DataServer {
	return &DataServer{}
}

func (ds *DataServer) Start() {
	ln, err := net.Listen("tcp", ":20")
	if err != nil {
		fmt.Println(err)
	}

	ds.acceptConnection(ln)
}

func (ds *DataServer) acceptConnection(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept error: ", err)
		}

		go ds.handleConnection(conn)
	}
}

func (ds *DataServer) handleConnection(conn net.Conn) {
	buf := new(bytes.Buffer)
	defer conn.Close()

	for {
		f, err := ds.deserializeFile(conn)
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

		err = ds.saveFile(f, "storage")
		if err != nil {
			fmt.Println(err)
			break
		}

		fmt.Printf("File has been saved to file system\n")
	}
}

func (ds *DataServer) deserializeFile(conn net.Conn) (*src.File, error) {
	f := &src.File{}

	decoder := gob.NewDecoder(conn)
	err := decoder.Decode(f)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (ds *DataServer) saveFile(f *src.File, where string) error {
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

// TODO: add from env CS Port
func (ds *DataServer) dialToControlServer() error {
	conn, err := net.Dial("tcp", ":2121")
	if err != nil {
		return err
	}

	ds.controlServerConn = conn

	return nil
}

// TODO: Add command handler
func (ds *DataServer) handleCommand(msg string, conn net.Conn) error {
	slicedCommand := strings.Split(msg, " ") // msg example: "USER Anonymous"
	cmd := strings.TrimSpace(slicedCommand[0])

	switch cmd {
	case command.LIST:

	}
}

// ftpserver passive mode connection
/*
func (fs *DataServer) newFTPClientAddr(newAddr net.Addr) {
	fs.ftpClientAddr = newAddr
}

func (fs *DataServer) connectToFTPClient() {

}*/
