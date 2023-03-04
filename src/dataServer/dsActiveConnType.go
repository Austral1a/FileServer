package dataServer

import (
	"fmt"
	"net"
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
