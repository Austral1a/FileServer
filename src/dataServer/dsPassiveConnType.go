package dataServer

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Austral1a/FileServer/src/utils"
	"io"
	"log"
	"net"
	"time"
)

type PassiveDataServer struct {
	Port    uint16
	ln      net.Listener
	Clients map[string]*DsFTPClient
}

func (pds *PassiveDataServer) NewServer() *PassiveDataServer {
	return &PassiveDataServer{Clients: make(map[string]*DsFTPClient), Port: 1024}
}

func (pds *PassiveDataServer) Start() error {
	// TODO: Get port from env
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", pds.Port))
	if err != nil {
		return err
	}

	pds.ln = ln

	go pds.acceptConnection(ln)

	log.Println("Passive DS started")

	return nil
}

func (pds *PassiveDataServer) Close() error {
	err := pds.Close()
	if err != nil {
		return err
	}

	return nil
}

func (pds *PassiveDataServer) acceptConnection(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept error: ", err)
		}

		err = pds.clientConnected(conn)
		if err != nil && err != io.EOF {
			fmt.Println("clientConnected:", err)
		}

		go pds.handleConnection(conn)
	}
}

func (pds *PassiveDataServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	ip, _ := utils.GetIpAndPortFromAddr(conn.RemoteAddr())

	client := pds.Clients[ip]

	for {
		buf := bytes.Buffer{}

		n, err := io.Copy(&buf, conn)
		if err != nil {
			fmt.Println("read from conn error:", err)
			continue
		}

		if n > 0 {
			client.SentDataCh <- buf.Bytes()

			client.AllDataIsSent <- struct{}{}
		}

		time.Sleep(time.Millisecond * 100)
	}
}

func (pds *PassiveDataServer) clientConnected(conn net.Conn) error {
	if pds.isClientConnected(conn.RemoteAddr()) {
		return errors.New(fmt.Sprintf("client %d already connected", conn.RemoteAddr()))
	}

	ip, _ := utils.GetIpAndPortFromAddr(conn.RemoteAddr())

	pds.Clients[ip] = &DsFTPClient{Conn: conn, ConnType: "passive", SentDataCh: make(chan []byte, 99), AllDataIsSent: make(chan struct{}, 1)}

	return nil
}

// TODO: imlp client disconnection func
func (pds *PassiveDataServer) isClientConnected(addr net.Addr) bool {
	ip, _ := utils.GetIpAndPortFromAddr(addr)

	_, ok := pds.Clients[ip]

	return ok
}

/*
	func (pds *PassiveDataServer) handleConnection(conn net.Conn) {
		buf := bytes.Buffer{}

		defer conn.Close()

		for {
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
*/
