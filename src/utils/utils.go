package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/mbndr/figlet4go"
	"io"
	"net"
	"os"
	"regexp"
)

func GetFileNameAndExt(fileName string) (name, ext string, err error) {
	// todo: RE is not safe
	re, err := regexp.Compile(`(?im)^(?P<Name>[^.]*)\.(?P<Ext>.*)$`)
	if err != nil {
		return "", "", nil
	}

	tempMap := map[string]string{}
	subExpNames := re.SubexpNames()

	for i, n := range re.FindAllStringSubmatch(fileName, -1)[0] {
		tempMap[subExpNames[i]] = n
	}

	return tempMap["Name"], tempMap["Ext"], nil
}

/*
	func SendRealFile(filename string) {
		conn, err := net.Dial("tcp", ":20")
		if err != nil {
			fmt.Println(err)
		}

		file, err := os.Open(filename)
		if err != nil {
			fmt.Println(err)
		}

		f, err := io.ReadAll(file)
		if err != nil {
			fmt.Println(err)
		}

		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)

		name, _, err := GetFileNameAndExt(file.Name())
		if err != nil {
			fmt.Println(err)
		}

		err = encoder.Encode(types.File{
			Name: name,

			Data: bytes.Buffer{}.Write(&f),
		})
		if err != nil {
			fmt.Println(err)
		}

		_, err = conn.Write(buf.Bytes())
		if err != nil {
			fmt.Println(err)
		}
	}
*/
func SendFile(size int) error {
	file := make([]byte, (1024*1000)*500)
	_, err := io.ReadFull(rand.Reader, file)
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		return err
	}

	binary.Write(conn, binary.LittleEndian, int64(size))
	n, err := io.CopyN(conn, bytes.NewReader(file), int64(size))
	if err != nil {
		return err
	}

	fmt.Printf("Written %d bytes over network\n", n)
	return nil
}

func HelloMsgAfterLogin() string {
	ascii := figlet4go.NewAsciiRender()

	options := figlet4go.NewRenderOptions()
	options.FontColor = []figlet4go.Color{
		figlet4go.ColorGreen,
		figlet4go.ColorYellow,
		figlet4go.ColorCyan,
	}
	options.FontName = "larry3d"

	renderStr, _ := ascii.RenderOpts("Hello fella", options)
	return renderStr

}

func GetIpAndPortFromAddr(addr net.Addr) (ip string, port int) {
	switch addr.(type) {

	case *net.TCPAddr:
		ip = addr.(*net.TCPAddr).IP.String()
		port = addr.(*net.TCPAddr).Port
	}

	return ip, port
}

func IsFileExists(filepath string) error {
	if _, err := os.Stat(filepath); err != nil {
		if err != nil {
			return err
		}
	}

	return nil
}
