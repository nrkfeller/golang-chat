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

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func main() {

	var user client
	user.on = true
	user.sendMessages = make(chan string)
	user.receiveMessages = make(chan string)
	udpAddr, err := net.ResolveUDPAddr("udp4", getLocalIP()+":10010") // change default port
	checkError(err, "main")
	user.conn, err = net.DialUDP("udp", nil, udpAddr)
	checkError(err, "main")
	defer user.conn.Close()

	fmt.Print("Enter Username: ")
	_, err = fmt.Scanln(&user.name)
	checkError(err, "main")

	// reader := bufio.NewReader(os.Stdin)
	// fmt.Print("Register with a server: ")
	// text, _ := reader.ReadString('\n')

	user.register()

	go user.printMessage()
	go user.receiveMessage()

	for {
		reader := bufio.NewReader(os.Stdin)
		//fmt.Print("Enter Command: ")
		text, _ := reader.ReadString('\n')
		t := strings.Split(strings.TrimSpace(text), " ")

		switch t[0] {
		case "REGISTER":
			user.register()
		case "PUBLISH":
			user.publish()
		case "INFORMReq":
			user.informreq()
		case "FINDReq":
			user.findreq()
		case "CHAT":
			user.chat()
		case "EXIT":
			user.exit()
		}
	}
}

func (c *client) exit() {
	msg := fmt.Sprintf("EXIT|20|%s|%s", c.name, c.conn.RemoteAddr().String())
	buf := []byte(strings.TrimSpace(msg))
	_, err := c.conn.Write(buf)
	checkError(err, "exit")
}

func (c *client) chat() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Who do you want to chat with? (1 person): ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	msg := fmt.Sprintf("CHAT|16|%s|%s", name, c.name)
	buf := []byte(strings.TrimSpace(msg))
	_, err := c.conn.Write(buf)
	checkError(err, "chat")

	for {
		creader := bufio.NewReader(os.Stdin)
		mchat, _ := creader.ReadString('\n')
		mchat = strings.TrimSpace(mchat)

		if mchat == "BYE" {
			mchat = fmt.Sprintf("BYE|14|%s|%s", c.name, c.conn.RemoteAddr().String())
			buf = []byte(mchat)
			_, err := c.conn.Write(buf)
			checkError(err, "chat")
			break
		}

		mchat = fmt.Sprintf("MCHAT|17|%s|%s", c.name, mchat)

		buf = []byte(mchat)
		_, err := c.conn.Write(buf)
		checkError(err, "chat")
	}
}

// func (c *client) exit() {
// 	msg := fmt.Sprintf("exit|14|%s|%s", c.name, c.conn.RemoteAddr().String())
// 	buf := []byte(strings.TrimSpace(msg))
// 	_, err := c.conn.Write(buf)
// 	checkError(err, "bye")
// 	fmt.Println("You have left the channel")
// }

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

	msg := fmt.Sprintf("PUBLISH|5|%s|%s|%t|%s", c.name, c.conn.RemoteAddr().String(), c.on, names)
	buf := []byte(strings.TrimSpace(msg))
	_, err := c.conn.Write(buf)
	checkError(err, "publish")
}

// CAN DELETE
// func (c *client) newregister() {
// 	reader := bufio.NewReader(os.Stdin)
// 	fmt.Print("Register with a new server: ")
// 	text, _ := reader.ReadString('\n')
//
// 	//text = strings.TrimSpace(text)
//
// 	udpAddr, err := net.ResolveUDPAddr("udp4", text)
// 	c.conn, err = net.DialUDP("udp", nil, udpAddr)
// 	checkError(err, "newregister1")
//
// 	msg := fmt.Sprintf("REGISTER|1|%s|%s", c.name, c.conn.RemoteAddr().String())
// 	buf := []byte(msg)
// 	_, err = c.conn.Write(buf)
// 	checkError(err, "newregister2")
// }

func (c *client) register() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Register with a server: ")
	text, _ := reader.ReadString('\n')

	udpAddr, err := net.ResolveUDPAddr("udp4", text)
	checkError(err, "register1")
	c.conn, err = net.DialUDP("udp", nil, udpAddr)
	checkError(err, "register2")

	msg := fmt.Sprintf("REGISTER|1|%s|%s", c.name, c.conn.RemoteAddr().String())
	buf := []byte(msg)
	_, err = c.conn.Write(buf)
	checkError(err, "register3")
}

func (c *client) receiveMessage() {
	var buf [128]byte
	for {
		n, err := c.conn.Read(buf[0:])

		checkError(err, "receiveMessage")
		if c.on {
			c.receiveMessages <- string(buf[0:n])
		}

		// msg := string(buf[0:n])
		// ss := strings.Split(msg, "|")
		// if len(ss) > 1 {
		// 	if ss[1] == "3" {
		// 		c.conn = nil
		// 		c.register()
		// 	}
		// }
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
		fmt.Fprintf(os.Stderr, "ERROR OCCURED: %s in func: %s ", err.Error(), funcName)
		os.Exit(1)
	}
}
