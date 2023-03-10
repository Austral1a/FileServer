package ftpserver

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Austral1a/FileServer/src/utils"
	"math"
	"math/rand"
	"net"
	"time"
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

	_, err = conn.Write([]byte(fmt.Sprintf("200 Type set to %s\n", newDataTransferType)))
	if err != nil {
		return err
	}

	return nil
}

func DoCommandLIST(conn net.Conn, s *FTPServer) error {
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

	// TODO: addн ClientAddr type as key instead of net.Addr for CS as well
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
