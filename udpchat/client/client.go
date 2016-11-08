package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func checkError(err error, funcName string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error:%s-----in func:%s", err.Error(), funcName)
		os.Exit(1)
	}
}

type Client struct {
	conn            *net.UDPConn
	gkey            bool
	userID          int
	userName        string
	sendMessages    chan string
	receiveMessages chan string
}

func (c *Client) func_sendMessage(sid int, msg string) {
	str := fmt.Sprintf("###%d##%d##%s##%s###", sid, c.userID, c.userName, msg)
	_, err := c.conn.Write([]byte(str))
	checkError(err, "func send message")
}

func (c *Client) sendMessage() {
	for c.gkey {
		msg := <-c.sendMessages
		str := fmt.Sprintf("###2##%d##%s##%s###", c.userID, c.userName, msg)
		_, err := c.conn.Write([]byte(str))
		checkError(err, "send message")
	}
}

func (c *Client) receiveMessage() {
	var buf [512]byte
	for c.gkey {
		n, err := c.conn.Read(buf[0:])
		checkError(err, "receive message")
		c.receiveMessages <- string(buf[0:n])
	}
}

func (c *Client) getMessages() {
	var msg string
	for c.gkey {
		fmt.Println("msg: ")
		_, err := fmt.Scanln(&msg)
		checkError(err, "getMessage")
		if msg == ":q" {
			c.gkey = false
		} else {
			c.sendMessages <- encodeMessage(msg)
		}
	}
}

func (c *Client) printMessage() {
	for c.gkey {
		msg := c.receiveMessages
		fmt.Println(msg)
	}
}

func encodeMessage(msg string) string {
	return strings.Join(strings.Split(strings.Join(strings.Split(msg, "\\"), "\\\\"), "#"), "\\#")
}

func nowTime() string {
	return time.Now().String()
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s, host:port", os.Args[0])
		os.Exit(1)
	}
	service := os.Args[1]
	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	checkError(err, "main")

	var c Client
	c.gkey = true
	c.sendMessages = make(chan string)
	c.receiveMessages = make(chan string)

	fmt.Println("Input id: ")
	_, err = fmt.Scanln(&c.userID)
	checkError(err, "main")
	_, err = fmt.Scanln(&c.userName)
	checkError(err, "main")

	c.conn, err = net.DialUDP("udp", nil, udpAddr)
	checkError(err, "main")

	defer c.conn.Close()

	c.func_sendMessage(1, c.userName+" has entered the chatroom")

	go c.printMessages()
	go c.receiveMessage()

	go c.sendMessage()
	c.getMessage()

	c.func_sendMessages(3, c.userName+" has left the chatroom")
	fmt.Println("Left chatroom")

	os.Exit(0)
}
