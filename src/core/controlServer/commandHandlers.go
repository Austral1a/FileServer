package controlServer

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"
)

var ErrCommandUSERWrongSyntax = errors.New(`'USER' command has wrong syntax`)

func DoCommandUSER(conn net.Conn, command string) error {
	userName := strings.TrimSpace(strings.Split(command, CommandUSER)[1])
	if userName == "" {
		return ErrCommandUSERWrongSyntax
	}

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

func DoCommandQUIT(conn net.Conn) error {
	_, err := conn.Write([]byte("221 Bye!\n"))
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

func DoCommandCWD(conn net.Conn, cs *ControlServer, newWorkingDir string) error {
	cs.changeWorkingDir(newWorkingDir)

	_, err := conn.Write([]byte("250 Working dir has been changed\n"))
	if err != nil {
		return err
	}

	return nil
}

func DoCommandEPSV(conn net.Conn) error {
	_, err := conn.Write([]byte("229 Entering Extended Passive Mode (|||20|)\n"))
	if err != nil {
		return err
	}

	return nil
}

// TODO: need a enum/union to A or I types

func DoCommandTYPE(conn net.Conn, cs *ControlServer, newDataTransferType string) error {
	err := cs.changeDataTransferType(conn.RemoteAddr(), newDataTransferType)
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

func DoCommandLIST(conn net.Conn, cs *ControlServer, flags string) error {
	conn, err := net.Dial("tcp", cs.dataServerAddr)
	if err != nil {
		// TODO: need to send proper code to FTP client
		fmt.Println("dial to data server error: ", err)
	}

	conn.Write([]byte(CommandLIST + " " + flags))

	resp := bytes.Buffer{}
	resp.WriteString("150 Files status ok; about to open data connection.\n")

	for _, entry := range dirEntry {
		resp.WriteString(entry.Name() + "\n")
	}

	_, err = conn.Write(resp.Bytes())
	if err != nil {
		return err
	}

	return nil
}