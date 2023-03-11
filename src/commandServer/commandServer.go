package commandServer

import (
	"errors"
	"fmt"
	"github.com/Austral1a/FileServer/src/types"
	"io"
	"net"
	"os"
	"time"
)

const FileStorageLocalPath = "storage"

type CsFTPClient struct {
	// DataTransferType can be: "A" (ASCII) or "I" (image/binary)
	DataTransferType string
	Conn             net.Conn
	ConnType         types.ConnectionType
	UserName         string
	CommandsQueueCh  chan string
}

type CommandServer struct {
	Clients map[net.Addr]*CsFTPClient
	// current working directory of Server
	WorkingDir string
	IsStarted  bool
}

func (cs *CommandServer) NewCommandServer() *CommandServer {
	return &CommandServer{Clients: make(map[net.Addr]*CsFTPClient), WorkingDir: FileStorageLocalPath}
}

func (cs *CommandServer) Start() {
	ln, err := net.Listen("tcp", ":21")
	if err != nil {
		fmt.Println("listen error:", err)
		os.Exit(1)
	}

	cs.IsStarted = true

	go cs.acceptLoop(ln)
}

func (cs *CommandServer) acceptLoop(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept error: ", err)
			continue
		}

		go cs.handleConnection(conn)
	}
}

func (cs *CommandServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// invokes only when client firstly connected.
	// Needed to send 220 code to FTP Client to establish the following commands
	err := cs.clientConnected(conn)
	if err != nil && err != io.EOF {
		fmt.Println("clientConnected:", err)
		return
	}

	// TODO: Client id is client's host (not full address ip:port how it was before)
	// TODO: Check if ftp client can connect from any port

	for {
		buf := make([]byte, 256)

		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("read from connection error:", err)
			break
		}

		command := string(buf[:n])

		if command != "" {
			fmt.Println("CLIENT'S COMMAND", command)
			client, ok := cs.Clients[conn.RemoteAddr()]
			if !ok {
				fmt.Println("client is not found")
			}

			client.CommandsQueueCh <- command
		}

		fmt.Println("command: ", command)
		fmt.Println(conn.RemoteAddr())

		time.Sleep(time.Millisecond * 100)
	}
}

// TODO: imlp client disconnection func
func (cs *CommandServer) isClientConnected(addr net.Addr) bool {
	_, ok := cs.Clients[addr]

	fmt.Println("CLIENT exists ", addr, ok)

	return ok
}

func (cs *CommandServer) clientConnected(conn net.Conn) error {
	if cs.isClientConnected(conn.RemoteAddr()) {
		return nil
	}

	_, err := conn.Write([]byte("220 Server ready\r\n"))
	if err != nil {
		return err
	}

	cs.Clients[conn.RemoteAddr()] = &CsFTPClient{Conn: conn, CommandsQueueCh: make(chan string, 1)}

	return nil
}

func (cs *CommandServer) SendMsgToFTPClient(clientAddr net.Addr, code int, msg string) error {
	client, ok := cs.Clients[clientAddr]
	if !ok {
		return errors.New("client is not found")
	}

	_, err := client.Conn.Write([]byte(fmt.Sprintf("%d %s\r\n", code, msg)))
	if err != nil {
		return err
	}

	return nil
}

func (cs *CommandServer) DisconnectClient(clientAddr net.Addr) error {
	client, ok := cs.Clients[clientAddr]
	if !ok {
		return errors.New(fmt.Sprintf("no such client: %s", client.Conn.RemoteAddr().String()))
	}

	err := client.Conn.Close()
	if err != nil {
		return err
	}

	delete(cs.Clients, clientAddr)

	return nil
}

func (cs *CommandServer) ChangeWorkingDir(newWorkDir string) {
	cs.WorkingDir = newWorkDir
}

func (cs *CommandServer) ChangeDataTransferType(clientAddr net.Addr, newTransferType string) error {
	client, ok := cs.Clients[clientAddr]
	if !ok {
		return errors.New(fmt.Sprintf("no such client: %d", client))
	}

	client.DataTransferType = newTransferType

	return nil
}
