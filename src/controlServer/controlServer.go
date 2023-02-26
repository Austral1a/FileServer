package controlServer

import (
	"errors"
	"fmt"
	"github.com/Austral1a/FileServer/src/command"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

const FileStorageLocalPath = "storage"

type Client struct {
	addr net.Addr
	// A (ASCII) or I (image/binary)
	dataTransferType string
}

type ControlServer struct {
	DataServerConn net.Conn
	Clients        map[net.Addr]*Client
	// current working directory of Server
	WorkingDir string
}

func NewControlServer() *ControlServer {
	cs := &ControlServer{clients: make(map[net.Addr]*Client), workingDir: FileStorageLocalPath}
	err := cs.dialToDataServer()

	if err != nil {
		fmt.Println("dial to ds error: ", err)
		os.Exit(1)
	}

	return cs
}

func (cs *ControlServer) Start() {

	ln, err := net.Listen("tcp", ":2121")
	if err != nil {
		fmt.Println("listen error:", err)
		os.Exit(1)
	}

	cs.acceptLoop(ln)
}

func (cs *ControlServer) acceptLoop(ln net.Listener) {

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept error: ", err)
			continue
		}

		go cs.handleConnection(conn)
	}
}

func (cs *ControlServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 256)

	// invokes only when client firstly connected.
	// Needed to send 220 code to FTP Client to establish the following commands
	err := cs.clientConnected(conn)
	if err != nil && err != io.EOF {
		fmt.Println("clientConnected:", err)
	}

	for {
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("read from connection error:", err)
			break
		}

		command := string(buf[:n])

		fmt.Println("command: ", command)
		fmt.Println(conn.RemoteAddr())

		err = cs.handleCommand(command, conn)
		if err != nil && err != io.EOF {
			fmt.Println("handle command error:", err)
			break
		}

		time.Sleep(time.Second / 2)
	}
}

// TODO: imlp client disconnection func
func (cs *ControlServer) isClientConnected(addr net.Addr) bool {
	_, ok := cs.clients[addr]

	fmt.Println("CLIENT exists ", addr, ok)

	return ok
}

func (cs *ControlServer) clientConnected(conn net.Conn) error {
	if cs.isClientConnected(conn.RemoteAddr()) {
		return errors.New(fmt.Sprintf("client %d already connected", conn.RemoteAddr()))
	}

	_, err := conn.Write([]byte("220 Server ready\n"))
	if err != nil {
		return err
	}

	cs.clients[conn.RemoteAddr()] = &Client{addr: conn.RemoteAddr()}

	return nil
}

func (cs *ControlServer) changeWorkingDir(newWorkDir string) {
	cs.workingDir = newWorkDir
}

func (cs *ControlServer) changeDataTransferType(clientAddr net.Addr, newTransferType string) error {
	client, ok := cs.clients[clientAddr]
	if !ok {
		return errors.New(fmt.Sprintf("no such client: %d", client))
	}

	client.dataTransferType = newTransferType

	return nil
}

// TODO: add ds port from env
func (cs *ControlServer) dialToDataServer() error {
	conn, err := net.Dial("tcp", ":20")
	if err != nil {
		return err
	}

	cs.dataServerConn = conn

	return nil
}

/*
to properly implement command there should be file storages type
possible imlp to File Storages is Factory Pattern
 1. Local File Storage (dedicated dir)
 2. AWS S3

Temporary there'll be 1 option
*/
func (cs *ControlServer) handleCommand(msg string, conn net.Conn) error {
	slicedCommand := strings.Split(msg, " ") // msg example: "USER Anonymous"
	cmd := strings.TrimSpace(slicedCommand[0])

	switch cmd {

	case command.USER:
		return DoCommandUSER(conn, msg)

	case command.PWD:
		return DoCommandPWD(conn)

	case command.SYST:
		return DoCommandSYST(conn)

	case command.OPTS:
		return DoCommandOPTS(conn)

	case command.FEAT:
		return DoCommandFEAT(conn)

	case command.CWD:
		newWorkingDir := slicedCommand[1]

		return DoCommandCWD(conn, cs, newWorkingDir)

	case command.EPSV:
		return DoCommandEPSV(conn)

	case command.TYPE:
		newDataTransferType := slicedCommand[1]

		return DoCommandTYPE(conn, cs, newDataTransferType)

	case command.QUIT:
		return DoCommandQUIT(conn)

	default:
		return nil
	}
}
