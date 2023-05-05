package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.LinstenMessage()
	return user
}

// 用户上线业务
func (user *User) Online() {
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()
	user.server.BroadCast(user, "已上线")
}

// 用户下线业务
func (user *User) Offline() {
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()
	user.server.BroadCast(user, "已下线")
}

// 发送消息
func (user *User) SendMessage(msg string) {
	user.conn.Write([]byte(msg))
}

// 用户处理消息业务
func (user *User) DoMessage(msg string) {
	if msg == "who" {
		user.server.mapLock.Lock()
		user.SendMessage("user have")
		for _, user := range user.server.OnlineMap {
			onlemsg := "[" + user.Addr + "]" + user.Name + "is online\n"
			user.SendMessage(onlemsg)
		}
	} else if len(msg) > 7 && msg[:7] == "rename@" {
		rename := msg[7:]
		_, ok := user.server.OnlineMap[user.Name]
		if ok {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, rename)
			user.server.OnlineMap[rename] = user
			user.server.mapLock.Unlock()
			user.Name = rename
			user.SendMessage("have new name is" + user.Name + "\n")

		} else {
			user.SendMessage("don't have this name \n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		toname := strings.Split(msg, "|")[1]
		if toname == "" {
			user.SendMessage("please send to to|name|message\n")
			return
		}
		remoteUser, ok := user.server.OnlineMap[toname]
		if !ok {
			user.SendMessage("don't have this name\n")
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			user.SendMessage("don't have message\n")
			return
		}
		remoteUser.SendMessage(user.Name + "say to you:" + content + "\n")
	} else {
		//广播消息
		user.server.BroadCast(user, msg)
	}

}

func (u *User) LinstenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}

}
