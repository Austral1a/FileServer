package commandServer

import (
	"errors"
	"fmt"
	"github.com/Austral1a/FileServer/src/types"
	"io"
	"net"
	"os"
)

const FileStorageLocalPath = "storage"

type FileRenameProcedurePayload struct {
	OldFilename string
	NewFilename string
}

type CsFTPClient struct {
	// DataTransferType can be: "A" (ASCII) or "I" (image/binary)
	DataTransferType string
	Conn             net.Conn
	ConnType         types.ConnectionType
	UserName         string
	/*
		Procedures responsible for resolving commands that are work together or can't work without each other
		e.g.: RNFR works only with proceeding RNTO
		map key is procedure name; e.g. file_rename
		map value is procedure payload
	*/
	//Procedures map[string]*interface{}
	RenameFileProcedure *FileRenameProcedurePayload
	//CommandsQueueCh  chan string
}

type Command struct {
	ClientConn net.Conn
	// Command can be "USER anonymous"
	Command string
}

type CommandServer struct {
	Clients       map[net.Addr]*CsFTPClient
	CommandsQueue chan Command
	// current working directory of Server
	WorkingDir string
	IsStarted  bool
}

func (cs *CommandServer) NewCommandServer() *CommandServer {
	return &CommandServer{Clients: make(map[net.Addr]*CsFTPClient), CommandsQueue: make(chan Command, 8), WorkingDir: FileStorageLocalPath}
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

		// adds a client to clients list
		err = cs.clientConnected(conn)
		if err != nil && err != io.EOF {
			fmt.Println("clientConnected:", err)
			continue
		}

		go cs.handleConnection(conn)
	}
}

func (cs *CommandServer) handleConnection(conn net.Conn) {
	defer conn.Close()

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
			client, _ := cs.Clients[conn.RemoteAddr()]
			fmt.Println(client, "- CLIENT with command", command)
			/*if !ok {
				fmt.Println("client is not found")
				continue
			}*/

			fmt.Println(command + " COMMAND  BEFORE CHAN")
			fmt.Println(conn.RemoteAddr(), " CLIENT ADDR")

			//client.CommandsQueueCh <- command
			cs.CommandsQueue <- Command{
				ClientConn: conn,
				Command:    command,
			}
		}
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

	cs.Clients[conn.RemoteAddr()] = &CsFTPClient{Conn: conn}

	err := cs.SendMsgToFTPClient(conn.RemoteAddr(), 220, "Server ready.")
	if err != nil {
		return err
	}

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

// TODO: WorkingDir should be in a FTPClient struct
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
