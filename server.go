package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

func NewServer(ip string, port int) *Server {
	return &Server{Ip: ip, Port: port}
}
func (server *Server) Handler(conn net.Conn) {
	fmt.Println("Hander链接建立成功")
}

func (server Server) Start() {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.Linster err", err)
		return
	}
	defer listen.Close()
	for {
		accept, err := listen.Accept()
		if err != nil {
			fmt.Println("linster accept err:", err)
			continue
		}
		go server.Handler(accept)
	}

}
