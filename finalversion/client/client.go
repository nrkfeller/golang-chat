package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type client struct {
	name            string
	conn            *net.UDPConn
	on              bool
	sendMessages    chan string
	receiveMessages chan string
}

func main() {

	var user client
	user.on = true
	user.sendMessages = make(chan string)
	user.receiveMessages = make(chan string)
	udpAddr, err := net.ResolveUDPAddr("udp4", "localhost:10010") // change default port
	checkError(err, "main")
	user.conn, err = net.DialUDP("udp", nil, udpAddr)
	checkError(err, "main")
	defer user.conn.Close()

	fmt.Print("Enter Username: ")
	_, err = fmt.Scanln(&user.name)
	checkError(err, "main")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Register with a server: ")
	text, _ := reader.ReadString('\n')

	user.register(text)

	go user.printMessage()
	go user.receiveMessage()

	for {
		reader := bufio.NewReader(os.Stdin)
		//fmt.Print("Enter Command: ")
		text, _ := reader.ReadString('\n')
		t := strings.Split(strings.TrimSpace(text), " ")

		switch t[0] {
		case "REGISTER":
			user.register(text)
		case "PUBLISH":
			user.publish()
		case "INFORMReq":
			user.informreq()
		case "FINDReq":
			user.findreq()
		case "BYE":
			user.bye()
		}
	}
}

func (c *client) bye() {
	msg := fmt.Sprintf("BYE|14|%s|%s", c.name, c.conn.RemoteAddr().String())
	buf := []byte(strings.TrimSpace(msg))
	_, err := c.conn.Write(buf)
	checkError(err, "bye")
}

func (c *client) findreq() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Request Information from which user?: ")
	name, _ := reader.ReadString('\n')
	name += " "

	msg := fmt.Sprintf("FINDReq|10|%s|%s", c.name, name)
	buf := []byte(strings.TrimSpace(msg))
	_, err := c.conn.Write(buf)
	checkError(err, "findreq")
}

func (c *client) informreq() {
	msg := fmt.Sprintf("INFORMReq|7|%s", c.name)
	buf := []byte(strings.TrimSpace(msg))
	_, err := c.conn.Write(buf)
	checkError(err, "informreq")
}

func (c *client) publish() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("People you want to whitelist: ")
	names, _ := reader.ReadString('\n')
	names += " "

	ans := ""
	for !(ans == "y" || ans == "n") {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Do you wish to be active? [y/n]: ")
		ans, _ = reader.ReadString('\n')
		ans = strings.TrimSpace(ans)
	}

	if ans == "y" {
		c.on = true
	} else {
		c.on = false
	}

	msg := fmt.Sprintf("PUBLISH|5|%s|%s|active:%t|%s", c.name, c.conn.RemoteAddr().String(), c.on, names)
	buf := []byte(strings.TrimSpace(msg))
	_, err := c.conn.Write(buf)
	checkError(err, "publish")
}

func (c *client) register(text string) {

	udpAddr, err := net.ResolveUDPAddr("udp4", "localhost:"+text)
	c.conn, err = net.DialUDP("udp", nil, udpAddr)
	checkError(err, "register")

	msg := fmt.Sprintf("REGISTER|1|%s|%s", c.name, c.conn.RemoteAddr().String())
	buf := []byte(msg)
	_, err = c.conn.Write(buf)
	checkError(err, "register")
}

func (c *client) receiveMessage() {
	var buf [512]byte
	for c.on {
		n, err := c.conn.Read(buf[0:])
		// modify for unpublished request
		checkError(err, "receiveMessage")
		c.receiveMessages <- string(buf[0:n])
	}
}

func (c *client) printMessage() {
	//var msg string
	for c.on {
		msg := <-c.receiveMessages
		fmt.Println(msg)
	}
}

func checkError(err error, funcName string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR OCCURED: %s in func: %s", err.Error(), funcName)
		os.Exit(1)
	}
}
