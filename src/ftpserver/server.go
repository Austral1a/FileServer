package ftpserver

import (
	"github.com/Austral1a/FileServer/src/command"
	"github.com/Austral1a/FileServer/src/controlServer"
	"github.com/Austral1a/FileServer/src/dataServer"
	"net"
	"strings"
)

const FileStorageLocalPath = "storage"

type FTPClient struct {
	addr net.Addr
	// A (ASCII) or I (image/binary)
	dataTransferType string
}

type FTPServer struct {
	clients       map[net.Addr]*FTPClient
	ftpClientAddr net.Addr
	// current working directory of Server
	workingDir string
	ds         *dataServer.DataServer
	cs         *controlServer.ControlServer
}

func (ftp *FTPServer) NewServer() *FTPServer {
	cs := controlServer.NewControlServer()
	ds := dataServer.NewDataServer()

	cs.Start()
	ds.Start()

	return &FTPServer{clients: make(map[net.Addr]*FTPClient), workingDir: FileStorageLocalPath, cs: cs, ds: ds}
}

/*
to properly implement command there should be file storages type
possible imlp to File Storages is Factory Pattern
 1. Local File Storage (dedicated dir)
 2. AWS S3

Temporary there'll be 1 option
*/
func (s *FTPServer) handleCommand(msg string, conn net.Conn) error {
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

		return DoCommandCWD(conn, s, newWorkingDir)

	case command.EPSV:
		return DoCommandEPSV(conn)

	case command.TYPE:
		newDataTransferType := slicedCommand[1]

		return DoCommandTYPE(conn, s, newDataTransferType)

	case command.QUIT:
		return DoCommandQUIT(conn)

	default:
		return nil
	}
}
