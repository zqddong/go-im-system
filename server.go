package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播 channel
	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// ListenMessager 监听Message 广播消息channel的goroutine 一旦有消息发送给全部User
func (s *Server) ListenMessager() {
	for {
		msg := <-s.Message
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.Unlock()
	}
}

// BroadCast 广播消息
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

func (s *Server) Handler(conn net.Conn) {
	//fmt.Println("链接建立成功")
	user := NewUser(conn, s)
	user.Online()

	// 监听用户是否活跃
	isLive := make(chan bool)

	// 接收客户端发送的消息
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
			// 提取用户消息 去除 \n
			msg := string(buf[:n-1])

			// 将得到的消息广播
			user.DoMessage(msg)

			// 用户任意消息，更新为活跃
			isLive <- true
		}
	}()

	// 当前handler 阻塞
	for {
		select {
		case <-isLive:
			// 当前用户活跃 重置定时器

		case <-time.After(time.Second * 300):
			// 触发超时
			// 将当前的User强制关闭
			user.SendMsg("你被踢下线了\n")

			// 销毁用户资源
			close(user.C)
			conn.Close()
			// 退出当前handler
			return // runtime.Goexit()
		}
	}
}

func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	// close listen socket
	defer listener.Close()

	// 启动监听 Message
	go s.ListenMessager()

	// accept
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		// do handler
		go s.Handler(conn)
	}

}
