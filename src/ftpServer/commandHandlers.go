package ftpserver

import (
	"bytes"
	"fmt"
	"github.com/Austral1a/FileServer/src/utils"
	"net"
)

func DoCommandUSER(conn net.Conn, userName string) error {
	// "anonymous" user handler
	if userName == "anonymous" {
		n, err := conn.Write([]byte("230 Anonymous login ok\n"))
		if err != nil {
			return err
		}

		fmt.Println("bytes written: ", n)

		return nil
	}

	return nil
}

func DoCommandPWD(conn net.Conn) error {
	_, err := conn.Write([]byte(fmt.Sprintf("257 \"%s\" %v", "", "\n")))
	if err != nil {
		return err
	}

	return nil
}

func DoCommandSYST(conn net.Conn) error {
	_, err := conn.Write([]byte("215 Unix-like, MacOS\n"))
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
	err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 221, "Bye")
	if err != nil {
		return err
	}

	err = s.Cs.DisconnectClient(conn.RemoteAddr())
	if err != nil {
		return err
	}

	return nil
}

func DoCommandFEAT(conn net.Conn) error {
	// TODO: refactor
	// need check for possible "extended features" list add it or not add it ) and impl
	supportedFeatures := bytes.Buffer{}

	supportedFeatures.Write([]byte("211 Extensions supported: \n"))
	// TODO: SIZE Command is not implemented, yet
	supportedFeatures.Write([]byte("SIZE\n"))

	_, err := conn.Write(supportedFeatures.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func DoCommandCWD(conn net.Conn, s *FTPServer, newWorkingDir string) error {
	s.Cs.ChangeWorkingDir(newWorkingDir)

	_, err := conn.Write([]byte("250 Working dir has been changed\n"))
	if err != nil {
		return err
	}

	return nil
}

func DoCommandEPSV(conn net.Conn, s *FTPServer) error {
	err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 229, "Entering Extended Passive Mode (|||20|).")
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

	_, err = conn.Write([]byte(fmt.Sprintf("200 Type set to %s\n", newDataTransferType)))
	if err != nil {
		return err
	}

	return nil
}

func DoCommandLIST(conn net.Conn, s *FTPServer, flags string) error {
	err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 125, "Data connection already open; transfer starting.")
	if err != nil {
		return err
	}

	list, err := s.Ds.GetFilesAndDirsListByLISTFormat()
	if err != nil {
		return err
	}

	fmt.Println(list.String())

	ip, _ := utils.GetIpAndPortFromAddr(conn.RemoteAddr())

	connType := s.defineConnTypeByClient(conn)
	switch connType {
	case "passive":
		fmt.Println(connType, " Conn type ")

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
	}

	err = s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 226, "Requested file action okay, completed.")
	if err != nil {
		return err
	}

	return nil
}

// TODO: Simplify everything!!!. just use switch case when need to understand where should be used Passive or Active Server!!!!
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

	fmt.Println(files.String(), " - MLSD")
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
