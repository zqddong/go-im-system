package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

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
	go user.ListenMessage()
	return user
}

// Online 用户上线 将用户加入到OnlineMap中
func (u *User) Online() {
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()
	u.server.BroadCast(u, "已上线")
}

// Offline 用户下线 将用户加从OnlineMap中删除
func (u *User) Offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()
	u.server.BroadCast(u, "下线")
}

// SendMsg 给当前User对应的客户端发送消息
func (u *User) SendMsg(msg string) {
	u.conn.Write([]byte(msg))
}

// DoMessage 用户消息业务
func (u *User) DoMessage(msg string) {
	if msg == "who" {
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线... \n"
			u.SendMsg(onlineMsg)
		}
		u.server.mapLock.Unlock()

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]

		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.SendMsg("当前用户名已被使用\n")
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()

			u.Name = newName
			u.SendMsg("您已更新用户名" + newName + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式： to|张三|消息内容

		// 获取目标用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			u.SendMsg("消息格式不正确，请使用\"to|张三|你好啊\"格式。\n")
			return
		}

		// 根据用户名得到对应的User
		remoteUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.SendMsg("该用户名不存在\n")
			return
		}

		// 获取消息内容 通过对方User 发送消息
		content := strings.Split(msg, "|")[2]
		if content == "" {
			u.SendMsg("不能发送空消息\n")
			return
		}

		remoteUser.SendMsg(u.Name + "对您说：" + content + "\n")

	} else {
		u.server.BroadCast(u, msg)
	}
}

// ListenMessage 监听当前User channel 的方法，一旦有消息 直接发送给对的客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}
