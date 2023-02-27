package ftpserver

import (
	"bytes"
	"fmt"
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
	_, err := conn.Write([]byte(fmt.Sprintf("257 \"%s\" %v", FileStorageLocalPath, "\n")))
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
	err := s.Ds.Start()
	if err != nil {
		return err
	}

	err = s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 229, "Entering Extended Passive Mode (|||20|).")
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

/*
How to imlp communication between ds and cs ?
	ds <-> cs bidirectional communication
	steps:
		1) dial to DS
		2) send to DS command (e.g. LIST); ds.Write("LIST")
		3) DS process commands the same way as CS does, but on "data" level; filesList -> FTP client
		4) DS send status of command either error or not.; status -> CS

	ds and cs its one program
	steps:
		1) e.g. getting LIST command, leverage DS directly from CS to send needed data
*/

func DoCommandLIST(conn net.Conn, s *FTPServer, flags string) error {
	// TODO: refactor, Start and Close data server works for passive mode but not for active
	//err := s.Ds.Start()
	//if err != nil {
	//	return err
	//}
	//
	//defer s.Ds.Close()

	err := s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 125, "Data Server ok; about to open data connection.")
	if err != nil {
		return err
	}

	_, err = s.Ds.GetFilesAndDirs()
	if err != nil {
		return err
	}

	//err = s.Ds.SendDataToFTPClient(conn, list.Bytes())
	//if err != nil {
	//	return err
	//}

	err = s.Cs.SendMsgToFTPClient(conn.RemoteAddr(), 226, "Files sent, OK.")
	if err != nil {
		return err
	}

	return nil
}
