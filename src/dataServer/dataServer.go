package dataServer

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/Austral1a/FileServer/src/types"
	"io"
	"log"
	"net"
	"os"
)

type DataServer struct {
	ln net.Listener
}

func (ds *DataServer) NewDataServer() *DataServer {
	return &DataServer{}
}

func (ds *DataServer) Start() error {
	ln, err := net.Listen("tcp", ":20")
	if err != nil {
		return err
	}

	ds.ln = ln

	go ds.acceptConnection(ln)

	return nil
}

func (ds *DataServer) Close() error {
	err := ds.ln.Close()
	if err != nil {
		return err
	}

	return nil
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
	buf := bytes.Buffer{}
	var b []byte
	defer conn.Close()

	for {
		conn.Read(b)

		fmt.Println("HANDLE CONN ", b)

		f, err := ds.deserializeFile(conn)
		if err != nil {
			fmt.Println("deserialization error: ", err)
			break
		}

		n, err := io.CopyN(&buf, conn, int64(len(f.Bytes)))
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

func (ds *DataServer) SendDataToFTPClient(conn net.Conn, info []byte) error {
	ftpClientConn, err := net.Dial("tcp", conn.RemoteAddr().String())
	if err != nil {
		return err
	}

	_, err = ftpClientConn.Write(info)
	if err != nil {
		return err
	}

	return nil
}

func (ds *DataServer) deserializeFile(conn net.Conn) (*types.File, error) {
	f := &types.File{}

	decoder := gob.NewDecoder(conn)
	err := decoder.Decode(f)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (ds *DataServer) saveFile(f *types.File, where string) error {
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

func (ds *DataServer) GetFilesAndDirs() (bytes.Buffer, error) {
	dir, err := os.ReadDir("storage")
	if err != nil {
		return bytes.Buffer{}, err
	}

	filesList := bytes.Buffer{}

	for _, entry := range dir {
		filesList.WriteString(entry.Name() + "\n")
	}

	return filesList, nil
}
