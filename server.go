package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

func NewServer(ip string, port int) *Server {
	return &Server{Ip: ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

// 监听广播 有消息就广播给所有的用户
func (this *Server) LinstenMessager() {
	for {
		msg := <-this.Message

		//将msg发送给全部的在线User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// 广播方法
func (server *Server) BroadCast(user *User, msg string) {
	sendmsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	server.Message <- sendmsg

}

func (server *Server) Handler(conn net.Conn) {
	user := NewUser(conn, server)
	user.Online()
	isLive := make(chan bool)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			//提取用户的消息(去除'\n')
			msg := string(buf[:n-1])

			//用户针对msg进行消息处理
			user.DoMessage(msg)

			//用户的任意消息，代表当前用户是一个活跃的
			isLive <- true
		}
	}()
	for {
		select {
		case <-isLive:
		case <-time.After(time.Second * 300):
			user.SendMessage("you are time out")
			close(user.C)
			conn.Close()
			return
		}
	}

}

func (server Server) Start() {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	//启动监听
	go server.LinstenMessager()

	for {
		accept, err := listen.Accept()
		if err != nil {
			fmt.Println("linster accept err:", err)
			continue
		}
		go server.Handler(accept)
	}
	defer listen.Close()
}
