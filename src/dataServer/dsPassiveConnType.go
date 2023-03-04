package dataServer

import (
	"errors"
	"fmt"
	"github.com/Austral1a/FileServer/src/utils"
	"io"
	"net"
	"time"
)

type PassiveDataServer struct {
	ln      net.Listener
	Clients map[string]*DsFTPClient
}

func (pds *PassiveDataServer) NewServer() *PassiveDataServer {
	return &PassiveDataServer{Clients: make(map[string]*DsFTPClient)}
}

func (pds *PassiveDataServer) Start() error {
	// TODO: Get port from env
	ln, err := net.Listen("tcp", ":20")
	if err != nil {
		return err
	}

	pds.ln = ln

	go pds.acceptConnection(ln)

	fmt.Println("Passive DS started")

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

		go pds.handleConnection(conn)
	}
}

func (pds *PassiveDataServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	err := pds.clientConnected(conn)

	if err != nil && err != io.EOF {
		fmt.Println("clientConnected:", err)
	}

	// Results: need to somehow imlp EPSV mode, mean FTP client connects to FTP Server DS and then when needed DS sends to FTP Client data
	for {
		time.Sleep(time.Millisecond * 100)
	}
}

func (pds *PassiveDataServer) clientConnected(conn net.Conn) error {
	if pds.isClientConnected(conn.RemoteAddr()) {
		return errors.New(fmt.Sprintf("client %d already connected", conn.RemoteAddr()))
	}

	ip, _ := utils.GetIpAndPortFromAddr(conn.RemoteAddr())

	pds.Clients[ip] = &DsFTPClient{Conn: conn, ConnType: "passive"}

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
