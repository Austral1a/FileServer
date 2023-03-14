package ftpserver

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Austral1a/FileServer/src/commandServer"
	"github.com/Austral1a/FileServer/src/types"
	"github.com/Austral1a/FileServer/src/utils"
	"math"
	"math/rand"
	"net"
	"os"
	"time"
)

func DoCommandUSER(conn net.Conn, s *FTPServer, userName string) error {
	defer func() {
		s.Cs.Clients[conn.RemoteAddr()].UserName = userName
	}()

	// "anonymous" user handler
	if userName == "anonymous" {
		err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 230, "Anonymous login ok.")
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func DoCommandPWD(conn net.Conn, s *FTPServer) error {
	err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 257, "/")
	if err != nil {
		return err
	}

	return nil
}

func DoCommandSYST(conn net.Conn, s *FTPServer) error {
	err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 215, "MACOS")
	if err != nil {
		return err
	}

	return nil
}

func DoCommandOPTS(conn net.Conn) error {
	// todo: awaits imlp of OPTS command
	// need: MODE, MLST, UTF8
	_, err := conn.Write([]byte("451\n"))
	if err != nil {
		return err
	}

	return nil
}

func DoCommandQUIT(conn net.Conn, s *FTPServer) error {
	err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 221, "Bye.")
	if err != nil {
		return err
	}

	err = s.Cs.DisconnectClient(conn.RemoteAddr())
	if err != nil {
		return err
	}

	return nil
}

func DoCommandFEAT(conn net.Conn, s *FTPServer) error {
	// TODO: refactor
	// need check for possible "extended features" list add it or not add it ) and impl
	supportedFeatures := bytes.Buffer{}

	supportedFeatures.Write([]byte("211-Extensions supported: \r\n"))
	// TODO: SIZE Command is not implemented, yet
	supportedFeatures.Write([]byte("feat=SIZE\r\n"))

	supportedFeatures.Write([]byte("211 End \r\n"))

	//err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 211, "Extensions supported:")
	_, err := conn.Write(supportedFeatures.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func DoCommandCWD(conn net.Conn, s *FTPServer, newWorkingDir string) error {
	s.Cs.ChangeWorkingDir(newWorkingDir)

	err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 250, "Working dir has been changed")
	if err != nil {
		return err
	}

	return nil
}

func DoCommandEPSV(conn net.Conn, s *FTPServer) error {
	err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 229, fmt.Sprintf("Entering Extended Passive Mode (|||%d|).", s.Ds.Pds.Port))
	if err != nil {
		return err
	}

	return nil
}

func getP1P2() (int, int, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	getRandomPort := func() int {
		return r.Intn(math.MaxUint16-1024) + 1024
	}

	for {
		// first generate random port, second cast uint16 to uint8 "port / 256"
		p1 := getRandomPort() / 256
		p2 := getRandomPort() / 256

		port := (p1 * 256) + p2

		if port < 1024 {
			return 0, 0, errors.New("port cannot be less than 1024 since values < 1024 are reserved")
		}

		if port > math.MaxUint16 {
			return 0, 0, errors.New(fmt.Sprintf("port cannot be more than %d since it is more than max port number", math.MaxUint16))
		}

		return p1, p2, nil
	}
}

func DoCommandPASV(conn net.Conn, s *FTPServer) error {
	p1 := s.Ds.Pds.Port / 256
	p2 := 0

	// TODO: ip is mocked
	err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 227, fmt.Sprintf("Entering Passive Mode (127,0,0,1,%d,%d).", p1, p2))
	if err != nil {
		return err
	}

	return nil
}

// TODO: need a enum/union to A or I types
func DoCommandTYPE(conn net.Conn, s *FTPServer, newDataTransferType string) error {
	err := s.Cs.ChangeDataTransferType(conn.RemoteAddr(), newDataTransferType)
	if err != nil {
		fmt.Println("change data transfer type error: ", err)
	}

	err = s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 200, "Type set to "+newDataTransferType)
	if err != nil {
		return err
	}

	return nil
}

func DoCommandDELE(conn net.Conn, s *FTPServer, filenameToDelete string) error {
	err := os.Remove("storage" + filenameToDelete)
	if err != nil {
		err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 550, "Requested action not taken. File not found")
		if err != nil {
			fmt.Println("send to client error: ", err)
		}
	}

	err = s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 250, "Requested file action okay, completed.")
	if err != nil {
		fmt.Println("send to client error: ", err)
	}

	return nil
}

func DoCommandSTOR(conn net.Conn, s *FTPServer, filenameToStore string) error {
	// TODO: STOR can re-write existing file
	err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 125, "Data connection already open; transfer starting.")
	if err != nil {
		fmt.Println("send to client error: ", err)
	}

	actualClient := s.getActualClient(conn)

L:
	for {
		select {

		case data := <-actualClient.SentDataCh:
			err := s.saveFile(&types.File{
				Name: filenameToStore,
				Data: data,
			}, "storage")
			if err != nil {
				fmt.Println("save file error: ", err)
			}

		case <-actualClient.AllDataIsSent:
			break L
		}
	}

	err = s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 250, "Requested file action okay, completed.")
	if err != nil {
		fmt.Println("send to client error: ", err)
	}

	return nil
}

func DoCommandSTAT(conn net.Conn, s *FTPServer, pathname string) error {
	// check if pathname is given or not

	// if given then send info about file
	/*
		example:
		Client: STAT /
		Server: 213-Status of '/'
		Server: drwxrwxr-x    3 user     group        4096 Jan 01 00:00 directory1
		Server: -rw-rw-r--    1 user     group        1024 Jan 01 00:00 file1
		Server: -rw-rw-r--    1 user     group        2048 Jan 01 00:00 file2
		Server: 213 End of status.
	*/
	if len(pathname) > 0 {
		buf := bytes.Buffer{}

		buf.WriteString(fmt.Sprintf("213-Status of '%s':\r\n", pathname))

		list, err := s.Ds.GetFilesAndDirsListByLISTFormat(pathname)
		if err != nil {
			return err
		}

		buf.WriteString(list.String())

		buf.WriteString("213 End of status.\r\n")

		_, err = conn.Write(buf.Bytes())
		if err != nil {
			return err
		}
	}

	// if not given then send info about ftp server
	/*
		example:
		Client: STAT
		Server: 211-FTP server status:
		Server:     Connected to ::1
		Server:     Logged in as username
		Server:     TYPE: ASCII
		Server:     No session bandwidth limit
		Server:     Session timeout in seconds is 300
		Server:     Control connection is plain text
		Server:     Data connections will be plain text
		Server:     At session startup, client count was 1
		Server:     vsFTPd 3.0.3
		Server: 211 End of status
	*/
	if len(pathname) == 0 {
		buf := bytes.Buffer{}

		buf.WriteString("211-FTP Server status:\r\n")

		infoMap := s.GetServerInfo(conn.RemoteAddr())

		for k, v := range infoMap {
			buf.WriteString("\t" + k + " " + v + "\r\n")
		}

		buf.WriteString("211 End of status\r\n")

		_, err := conn.Write(buf.Bytes())
		if err != nil {
			return err
		}

		_, err = conn.Write(buf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func DoCommandLIST(conn net.Conn, s *FTPServer) error {
	err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 125, "Data connection already open; transfer starting.")
	if err != nil {
		return err
	}

	list, err := s.Ds.GetFilesAndDirsListByLISTFormat("")
	if err != nil {
		return err
	}

	fmt.Println(list.String())

	ip, _ := utils.GetIpAndPortFromAddr(conn.RemoteAddr())

	connType := s.defineConnTypeByClient(conn)
	switch connType {

	case "passive":
		err = s.Ds.SendDataToFTPClient(s.Ds.Pds.Clients[ip], list.Bytes())
		if err != nil {
			return err
		}

	case "active":
		for _, adsClient := range s.Ds.Ads {
			adsClientIp, _ := utils.GetIpAndPortFromAddr(adsClient.Client.Conn.RemoteAddr())

			if ip == adsClientIp {
				err = s.Ds.SendDataToFTPClient(adsClient.Client, list.Bytes())
				if err != nil {
					return err
				}
			}
		}

	default:
		err = s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 451, "Requested action aborted. Local error in processing.")
		if err != nil {
			return err
		}

		return nil
	}

	// TODO: probably when there conn type branches  "s.Cs.SendMsgToFTPClient" should be in active and passive to handle different response codes
	err = s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 250, "Requested file action okay, completed.")
	if err != nil {
		return err
	}

	return nil
}

func DoCommandMLSD(conn net.Conn, s *FTPServer) error {
	err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 125, "Data connection already open; transfer starting.")
	if err != nil {
		return err
	}

	files, err := s.Ds.GetFilesAndDirsByMLSDFormat()
	if err != nil {
		return err
	}

	ip, _ := utils.GetIpAndPortFromAddr(conn.RemoteAddr())

	// TODO: add ClientAddr type as key instead of net.Addr for CS as well
	connType := s.defineConnTypeByClient(conn)
	switch connType {
	case "passive":
		err = s.Ds.SendDataToFTPClient(s.Ds.Pds.Clients[ip], files.Bytes())
		if err != nil {
			return err
		}

	case "active":
		for _, adsClient := range s.Ds.Ads {
			adsClientIp, _ := utils.GetIpAndPortFromAddr(adsClient.Client.Conn.RemoteAddr())

			if ip == adsClientIp {
				err = s.Ds.SendDataToFTPClient(adsClient.Client, files.Bytes())
				if err != nil {
					return err
				}
			}
		}
	}

	err = s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 250, "Requested file action okay, completed.")
	if err != nil {
		return err
	}

	return nil
}

func DoCommandRNFR(conn net.Conn, s *FTPServer, filenameToBeRenamed string) error {
	defer func() {
		s.Cs.Clients[conn.RemoteAddr()].RenameFileProcedure = &commandServer.FileRenameProcedurePayload{
			OldFilename: filenameToBeRenamed,
			NewFilename: "",
		}
	}()

	err := utils.IsFileExists("storage" + filenameToBeRenamed)
	if err != nil {
		err = s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 450, "Requested file action not taken. File unavailable.")
		if err != nil {
			fmt.Println("file error: ", err)
		}
	}

	err = s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 350, "Requested file action pending further information.")
	if err != nil {
		fmt.Println("send to msg to client error: ", err)
	}

	return nil
}

func DoCommandRNTO(conn net.Conn, s *FTPServer, newFilename string) error {
	procedure := s.Cs.Clients[conn.RemoteAddr()].RenameFileProcedure

	procedure.NewFilename = newFilename

	defer func() {
		procedure.OldFilename = ""
		procedure.NewFilename = ""
	}()

	err := os.Rename("storage"+procedure.OldFilename, "storage"+procedure.NewFilename)
	if err != nil {
		err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 553, "Requested action not taken. Rename error.")
		if err != nil {
			fmt.Println("send to msg to client error: ", err)
		}
	}

	err = s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 250, "Requested file action okay, completed.")
	if err != nil {
		fmt.Println("send to msg to client error: ", err)
	}

	return nil
}
