package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
		//Name:       "",
		//conn:       nil,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil
	}

	client.conn = conn

	return client
}

// DealResponse 处理Server回应的消息直接显示到终端
func (c *Client) DealResponse() {
	// 一旦c.coon 有数据 直接copy到stdout标准输出 永久阻塞监听
	io.Copy(os.Stdout, c.conn)
}

func (c *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("4.退出")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		c.flag = flag
		return true
	} else {
		fmt.Println("请输入合法的数字")
		return false
	}

}

func (c *Client) Run() {
	for c.flag != 0 {
		for c.menu() != true {
			//
		}

		// 根据不同的模式处理业务
		switch c.flag {
		case 1:
			//fmt.Println("公聊模式选择...")
			c.PublicChat()
			break
		case 2:
			//fmt.Println("私聊模式选择...")
			c.PrivateChat()
			break
		case 3:
			//fmt.Println("更新用户名选择...")
			c.UpdateName()
			break
		}
	}
}

func (c *Client) PublicChat() {
	var chatMsg string

	fmt.Println("请输入聊天内容，exit退出")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {

		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := c.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println("请输入聊天内容，exit退出")
		fmt.Scanln(&chatMsg)

	}
}

// SelectUsers 查询在线用户
func (c *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}
}

func (c *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	c.SelectUsers()
	fmt.Println("请输入聊天对象【zhangsan】exit退出")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println("请输入消息内容，exit退出")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
			_, err := c.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				return
			}

			chatMsg = ""
			fmt.Println("请输入聊天内容，exit退出")
			fmt.Scanln(&chatMsg)
		}

		c.SelectUsers()
		fmt.Println("请输入聊天对象【zhangsan】exit退出")
		fmt.Scanln(&remoteName)
	}
}

func (c *Client) UpdateName() bool {
	fmt.Println("请输入用户名")
	fmt.Scanln(&c.Name)

	sendMsg := "rename|" + c.Name + "\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址默认127.0.0.1")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口")
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("++++++++++ 链接服务器失败")
		return
	}

	go client.DealResponse()

	fmt.Println("++++++++++ 链接服务器成功")

	client.Run()
}
