package dataServer

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/Austral1a/FileServer/src/types"
	"github.com/google/uuid"
	"log"
	"net"
	"os"
)

type DsFTPClient struct {
	Conn     net.Conn
	ConnType types.ConnectionType
}

type ClientAddr struct {
	// why is not net.IP; it is not comparable as net.IP - []byte
	Ip   string
	Port int
}

type DataServer struct {
	// when ftp client wants to connect via active conn type FTP client address adds to this chan
	PendingActiveConnClientsQueue chan ClientAddr

	Pds *PassiveDataServer
	Ads []*ActiveDataServer
}

func (ds *DataServer) NewDataServer() *DataServer {
	return &DataServer{}
}

func (ds *DataServer) Start() {
	ds.startPassiveServer()
	ds.startActiveServer()
}

func (ds *DataServer) startActiveServer() {
	go func() {
		for {
			select {
			case client := <-ds.PendingActiveConnClientsQueue:
				ads := new(ActiveDataServer).NewServer()

				ds.Ads = append(ds.Ads, ads)

				err := ads.DialToFTPClient(client.Port)
				if err != nil {
					fmt.Println("@DS_builder conn to FTP client error: ", err)
				}
			}
		}
	}()
}

func (ds *DataServer) startPassiveServer() {
	pds := new(PassiveDataServer).NewServer()

	ds.Pds = pds

	err := pds.Start()
	if err != nil {
		fmt.Println("passive data server start error:", err)
	}

}

func (ds *DataServer) SendDataToFTPClient(client *DsFTPClient, info []byte) error {
	_, err := client.Conn.Write(info)
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

func (ds *DataServer) GetFilesAndDirsListByLISTFormat() (bytes.Buffer, error) {
	dir, err := os.ReadDir("storage")
	if err != nil {
		return bytes.Buffer{}, err
	}

	files := bytes.Buffer{}

	for _, entry := range dir {
		info, err := entry.Info()
		if err != nil {
			return bytes.Buffer{}, err
		}

		files.WriteString(fmt.Sprintf("-%s %d %s %s\n", info.Mode().Perm(), info.Size(), info.ModTime().Format("Jan 01 2006"), info.Name()))
	}

	return files, nil
}

func (ds *DataServer) GetFilesAndDirsByMLSDFormat() (bytes.Buffer, error) {
	dir, err := os.ReadDir("storage")
	if err != nil {
		return bytes.Buffer{}, err
	}

	files := bytes.Buffer{}

	for _, entry := range dir {

		// format and return only files for now
		info, err := entry.Info()
		if err != nil {
			return bytes.Buffer{}, err
		}

		files.WriteString(fmt.Sprintf("Type=%s;Unique=%s;Size=%d;Modify=%d;Perm=%s; %s\n", "file", uuid.New().String(), info.Size(), info.ModTime().Unix(), "r", info.Name()))
	}

	return files, nil
}
