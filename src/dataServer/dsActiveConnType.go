package dataServer

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"time"
)

type ActiveDataServer struct {
	Client *DsFTPClient
}

func (ads *ActiveDataServer) NewServer() *ActiveDataServer {
	return &ActiveDataServer{}
}

func (ads *ActiveDataServer) DialToFTPClient(port int) error {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	ads.Client = &DsFTPClient{Conn: conn, ConnType: "active"}

	return nil
}

func (ads *ActiveDataServer) ReadConn(client *DsFTPClient) {

	for {
		buf := bytes.Buffer{}

		n, err := io.Copy(&buf, client.Conn)
		if err != nil {
			fmt.Println("read from conn error:", err)
			continue
		}

		if n > 0 {
			client.SentDataCh <- buf.Bytes()

			client.AllDataIsSent <- struct{}{}
		}

		fmt.Printf("reveiced %d bytes over network\n", n)
		time.Sleep(time.Millisecond * 100)
	}

}
