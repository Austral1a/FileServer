package ftpserver

import (
	"fmt"
	"github.com/Austral1a/FileServer/src/command"
	"github.com/Austral1a/FileServer/src/commandServer"
	"github.com/Austral1a/FileServer/src/dataServer"
	"net"
	"strings"
)

const FileStorageLocalPath = "storage"

type FTPServer struct {
	Ds *dataServer.DataServer
	Cs *commandServer.CommandServer
}

func (ftp *FTPServer) NewFTPServer() *FTPServer {
	cs := ftp.Cs.NewCommandServer()
	cs.Start()

	ds := ftp.Ds.NewDataServer()

	return &FTPServer{Cs: cs, Ds: ds}
}

func (ftp *FTPServer) HandleCommands() {
	for {
		if len(ftp.Cs.Clients) < 1 {
			continue
		}

		fmt.Println(len(ftp.Cs.Clients), " CLIENTE LENTH")

		for _, client := range ftp.Cs.Clients {
			select {
			case cmd := <-client.CommandsQueueCh:
				fmt.Println(client.Conn.RemoteAddr().String(), " clients conn")
				err := ftp.handleCommand(client.Conn, cmd)
				if err != nil {
					fmt.Printf("can't handle command: %s; from client: %s; err: %s\n", strings.TrimSpace(cmd), client.Conn.RemoteAddr().String(), err)
				}
			}
		}
	}
}

/*
to properly implement command there should be file storages type
possible imlp to File Storages is Factory Pattern
 1. Local File Storage (dedicated dir)
 2. AWS S3

Temporary there'll be 1 option
*/
func (ftp *FTPServer) handleCommand(conn net.Conn, msg string) error {
	slicedCommand := strings.Split(msg, " ") // msg example: "USER Anonymous"
	cmd := strings.TrimSpace(slicedCommand[0])

	switch cmd {

	case command.USER:
		userName := strings.TrimSpace(slicedCommand[1])

		return DoCommandUSER(conn, userName)

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

		return DoCommandCWD(conn, ftp, newWorkingDir)

	case command.EPSV:
		return DoCommandEPSV(conn, ftp)

	case command.TYPE:
		newDataTransferType := slicedCommand[1]

		return DoCommandTYPE(conn, ftp, newDataTransferType)

	case command.LIST:
		flags := slicedCommand[1]

		return DoCommandLIST(conn, ftp, flags)

	case command.QUIT:
		return DoCommandQUIT(conn, ftp)

	default:
		return nil
	}
}
