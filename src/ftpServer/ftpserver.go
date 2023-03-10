package ftpserver

import (
	"fmt"
	"github.com/Austral1a/FileServer/src/command"
	"github.com/Austral1a/FileServer/src/commandServer"
	"github.com/Austral1a/FileServer/src/dataServer"
	"github.com/Austral1a/FileServer/src/types"
	"github.com/Austral1a/FileServer/src/utils"
	"net"
	"strconv"
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
	ds.Start()

	return &FTPServer{Cs: cs, Ds: ds}
}

func (ftp *FTPServer) HandleCommands() {
	for {
		if len(ftp.Cs.Clients) < 1 {
			continue
		}

		for _, client := range ftp.Cs.Clients {
			select {
			case cmd := <-client.CommandsQueueCh:
				fmt.Println(cmd, " - CMD")
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
	cmd, args := ftp.sliceUpCommand(msg)

	switch cmd {

	case command.USER:
		return DoCommandUSER(conn, ftp, args)

	case command.PWD:
		return DoCommandPWD(conn, ftp)

	case command.SYST:
		return DoCommandSYST(conn, ftp)

	case command.OPTS:
		return DoCommandOPTS(conn)

	case command.FEAT:
		return DoCommandFEAT(conn, ftp)

	case command.CWD:
		return DoCommandCWD(conn, ftp, args)

	case command.EPSV:
		return DoCommandEPSV(conn, ftp)

	case command.PASV:
		return DoCommandPASV(conn, ftp)

	case command.EPRT:
		return DoCommandEPSV(conn, ftp)

	case command.TYPE:
		return DoCommandTYPE(conn, ftp, args)

	case command.STAT:
		return DoCommandSTAT(conn, ftp, args)

	case command.LIST:
		return DoCommandLIST(conn, ftp)

	case command.MLSD:
		return DoCommandMLSD(conn, ftp)

	case command.QUIT:
		return DoCommandQUIT(conn, ftp)

	default:
		return nil
	}
}

// defineConnTypeByCommand expects commands: LIST, USER, PORT and so on...; without args;
func (ftp *FTPServer) defineConnTypeByCommand(cmd string) types.ConnectionType {
	for _, actCommand := range command.CommandsEnablingActiveConnType {
		if actCommand == cmd {
			return "active"
		}
	}

	for _, pasvCommand := range command.CommandsEnablingPassiveConnType {
		if pasvCommand == cmd {
			return "passive"
		}
	}

	return ""
}

// SliceUpCommand slices up a command; example of command: "LIST -a" where LIST is cmdItself, -a is args
func (ftp *FTPServer) sliceUpCommand(command string) (cmdItself string, args string) {
	slicedCommand := strings.Split(command, " ") // msg example: "USER Anonymous"

	cmdItself = strings.TrimSpace(slicedCommand[0])

	if len(slicedCommand) > 1 {
		args = strings.TrimSpace(slicedCommand[1])
	}

	return
}

func (ftp *FTPServer) defineConnTypeByClient(conn net.Conn) types.ConnectionType {
	ip, _ := utils.GetIpAndPortFromAddr(conn.RemoteAddr())

	// check if this client ip is in passive ds
	_, ok := ftp.Ds.Pds.Clients[ip]
	if ok {
		return "passive"
	}

	// check if this client ip is in active ds
	for _, adsClient := range ftp.Ds.Ads {
		// just check if addr are the same if yes then its active conn type
		if conn.RemoteAddr() == adsClient.Client.Conn.RemoteAddr() {
			return "active"
		}
	}

	return ""
}

// TODO: remade map into struct
func (ftp *FTPServer) GetServerInfo(clientAddr net.Addr) map[string]string {
	infoMap := make(map[string]string)

	client := ftp.Cs.Clients[clientAddr]

	infoMap["Connected to"] = ":21"
	infoMap["Logged in as"] = client.UserName
	if client.DataTransferType == "A" {
		infoMap["TYPE:"] = "ASCII"
	} else if client.DataTransferType == "I" {
		infoMap["TYPE:"] = "image/binary"
	}
	infoMap["Session timeout in seconds is"] = "0"
	infoMap["Control connection"] = "is plain text"

	if client.DataTransferType == "A" {
		infoMap["Data connections will be"] = "plain text"
	} else if client.DataTransferType == "I" {
		infoMap["Data connections will be"] = "binary"
	}

	infoMap["At session startup, client count was"] = strconv.Itoa(len(ftp.Cs.Clients))
	infoMap["ftpserver"] = "0.0.1"

	return infoMap
}
