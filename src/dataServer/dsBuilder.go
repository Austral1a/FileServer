package dataServer

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/Austral1a/FileServer/src/types"
	"log"
	"mime"
	"net"
	"os"
	"path/filepath"
)

type DsFTPClient struct {
	Conn     net.Conn
	ConnType types.ConnectionType

	// SentDataCh stores received from FTP Client bytes. Bytes are typically files.
	SentDataCh    chan []byte
	AllDataIsSent chan struct{}
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
	newFile, err := os.Create(where + f.Name)
	defer func() {
		err := newFile.Close()
		if err != nil {
			log.Println("Failed to close file: ", err)
		}
	}()
	if err != nil {
		return err
	}

	_, err = newFile.Write(f.Data)
	if err != nil {
		return err
	}

	return nil
}

func (ds *DataServer) GetFilesAndDirsListByLISTFormat(pathname string) (bytes.Buffer, error) {
	dir, err := os.ReadDir("storage" + pathname)
	if err != nil {
		return bytes.Buffer{}, err
	}

	files := bytes.Buffer{}

	for _, entry := range dir {
		info, err := entry.Info()
		if err != nil {
			return bytes.Buffer{}, err
		}

		files.WriteString(fmt.Sprintf("%s 1 %d %s %s\r\n", info.Mode().Perm(), info.Size(), info.ModTime().Format("Jan 01 2006"), info.Name()))
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
		facts := make(map[string]any)

		// format and return only files for now
		info, err := entry.Info()
		if err != nil {
			return bytes.Buffer{}, err
		}

		fmt.Println("!!!FILE PATH", info.Name())
		// need to gather all files and dirs in a list, using new func
		factType := ""

		if !info.IsDir() {
			factType = "file"
		} else {
			factType = "dir"
		}

		facts["Size"] = info.Size()
		facts["Modify"] = info.ModTime()
		// no data about this on linux, so leave empty
		facts["Create"] = ""
		facts["Type"] = factType
		facts["Unique"] = info.ModTime().UnixMilli()
		// TODO: skip this for now
		//facts["Perm"] = info.Mode().Perm()

		// imagine all files are en-US
		facts["Lang"] = "en-US"
		facts["Media-Type"] = mime.TypeByExtension(filepath.Ext(info.Name()))
		// TODO: imagine all files are in UTF-8 encoding, for now..
		facts["CharSet"] = "UTF-8"

		var factsLine string

		for key, val := range facts {
			if val != "" {
				factsLine += fmt.Sprintf("%s=%v;", key, val)
			}
		}

		fmt.Println(factsLine, " FACTS LINE")

		files.WriteString(fmt.Sprintf("%s %s\r\n", factsLine, info.Name()))
	}

	//fmt.Println(files.String())

	return files, nil
}
