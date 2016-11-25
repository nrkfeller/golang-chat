package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// Client is a struct for user
type Client struct {
	conn            *net.UDPConn
	gkey            bool //用来判断用户退出
	userID          int
	userName        string
	sendMessages    chan string
	receiveMessages chan string
}

//突然加上一个函数，不加就需要去掉for或者多设一个变量，
func (c *Client) funcSendMessage(sid int, msg string) {
	str := fmt.Sprintf("%d##%d##%s##%s", sid, c.userID, c.userName, msg)
	//str := fmt.Sprintf("%s : %s", c.userName, msg)
	_, err := c.conn.Write([]byte(str))
	checkError(err, "func_sendMessage")
}

//send
func (c *Client) sendMessage() {
	for c.gkey {
		msg := <-c.sendMessages
		//str := fmt.Sprintf("(%s) \n %s: %s", nowTime(), c.userName,msg)
		str := fmt.Sprintf("2##%d##%s##%s", c.userID, c.userName, msg)
		_, err := c.conn.Write([]byte(str))
		checkError(err, "sendMessage")
	}

}

//接收
func (c *Client) receiveMessage() {
	var buf [512]byte
	for c.gkey {
		n, err := c.conn.Read(buf[0:])
		checkError(err, "receiveMessage")
		recmsg := string(buf[0:n])
		c.receiveMessages <- recmsg
		if len(recmsg) > 15 {
			if recmsg[0:15] == "REGISTER-DENIED" {
				c.gkey = false
			}
		}
	}

}

//获得输入并处理之，这里有Println
func (c *Client) getMessage() {
	for c.gkey {
		fmt.Println("msg: ")

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")
		msg, err := reader.ReadString('\n')

		checkError(err, "getMessage")
		if msg == "BYE" {
			c.gkey = false
		} else {
			c.sendMessages <- encodeMessage(msg)
		}
	}
}

//打印，这里有Println
func (c *Client) printMessage() {
	//var msg string
	for c.gkey {
		msg := <-c.receiveMessages
		fmt.Println(msg)
	}
}

//转换需要发送的字符串
func encodeMessage(msg string) string {
	return strings.Join(strings.Split(strings.Join(strings.Split(msg, "\\"), "\\\\"), "#"), "\\#")

}
func nowTime() string {
	return time.Now().String()
}
func checkError(err error, funcName string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error:%s-----in func: %s", err.Error(), funcName)
		os.Exit(1)
	}
}

func connectTo(port string) {

	udpAddr, err := net.ResolveUDPAddr("udp4", "localhost:"+port)
	checkError(err, "main")

	var c Client
	c.gkey = true
	c.sendMessages = make(chan string)
	c.receiveMessages = make(chan string)

	c.userID = rand.Intn(1000)
	fmt.Println("Enter Name: ")
	_, err = fmt.Scanln(&c.userName)
	checkError(err, "main")

	c.conn, err = net.DialUDP("udp", nil, udpAddr)
	checkError(err, "main")

	defer c.conn.Close()

	c.funcSendMessage(1, c.userName+" has entered the chatroom")

	// failover to next server
	var buf [128]byte
	n, err := c.conn.Read(buf[0:])
	checkError(err, "receiveMessage")
	recmsg := string(buf[0:n])
	if len(recmsg) > 15 {
		if recmsg[0:15] == "REGISTER-DENIED" {
			return
		}
	}

	go c.printMessage()
	go c.receiveMessage()

	go c.sendMessage()
	c.getMessage()

	c.funcSendMessage(3, c.userName+" has left the chatroom")
	fmt.Println("Left chatroom")

}

func main() {
	// if len(os.Args) != 2 {
	// 	fmt.Fprintf(os.Stderr, "Usage: %s, host:port", os.Args[0])
	// 	os.Exit(1)
	// }
	// service := os.Args[1]
	connectTo("1200")

	connectTo("1201")
}
