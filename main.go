package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"time"
)

type File struct {
	Name      string
	Extension string

	Bytes []byte
}

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

		go fs.readLoop(conn)
	}
}

func (fs *FileServer) readLoop(conn net.Conn) {
	buf := new(bytes.Buffer)
	defer conn.Close()

	c := 0

	for {
		f, err := fs.deserializeFile(conn)
		if err != nil {
			fmt.Println("deserialization error: ", err)
			break
		}

		fmt.Println(c, " - bytes and counter")

		fmt.Println(c, " - counter")
		fmt.Println((*f).Name, " - name")

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
		c++

	}
}

func (fs *FileServer) deserializeFile(conn net.Conn) (*File, error) {
	f := &File{}

	decoder := gob.NewDecoder(conn)
	err := decoder.Decode(f)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (fs *FileServer) saveFile(f *File) error {
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

func getFileNameAndExt(fileName string) (name, ext string, err error) {
	// todo: RE is not safe
	re, err := regexp.Compile(`(?im)^(?P<Name>[^.]*)\.(?P<Ext>.*)$`)
	if err != nil {
		return "", "", nil
	}

	tempMap := map[string]string{}
	subExpNames := re.SubexpNames()

	for i, n := range re.FindAllStringSubmatch(fileName, -1)[0] {
		tempMap[subExpNames[i]] = n
	}

	return tempMap["Name"], tempMap["Ext"], nil
}

func sendRealFile(filename string) {
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		fmt.Println(err)
	}

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}

	f, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	name, ext, err := getFileNameAndExt(file.Name())
	if err != nil {
		fmt.Println(err)
	}

	err = encoder.Encode(File{
		Name:      name,
		Extension: ext,

		Bytes: f,
	})
	if err != nil {
		fmt.Println(err)
	}

	_, err = conn.Write(buf.Bytes())
	if err != nil {
		fmt.Println(err)
	}
}

func sendFile(size int) error {
	file := make([]byte, (1024*1000)*500)
	_, err := io.ReadFull(rand.Reader, file)
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		return err
	}

	binary.Write(conn, binary.LittleEndian, int64(size))
	n, err := io.CopyN(conn, bytes.NewReader(file), int64(size))
	if err != nil {
		return err
	}

	fmt.Printf("Written %d bytes over network\n", n)
	return nil
}

func main() {

	go func() {
		time.Sleep(time.Millisecond * 300)
		sendRealFile("googlechrome.dmg")
		sendRealFile("vlc.dmg")
		sendRealFile("vlc.dmg")
	}()
	server := &FileServer{}
	server.start()
}
