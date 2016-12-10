package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type server struct {
	conn     *net.UDPConn
	messages chan string
	clients  map[string]*client
	port     string
}

type client struct {
	ID          int
	name        string
	active      bool
	userAddr    *net.UDPAddr
	whiteList   map[string]int
	chatpartner *client
}

type message struct {
	command string
	rqnum   string
	body    []string
}

// ./server -port=:1202

var port string

func init() {
	flag.StringVar(&port, "port", ":1200", "help message for flagname")
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

	flag.Parse()

	//fmt.Println(port)

	udpAddr1, err := net.ResolveUDPAddr("udp4", getLocalIP()+port)

	fmt.Println(udpAddr1)
	checkError(err)
	var s1 server
	s1.port = port
	s1.messages = make(chan string, 20)
	s1.clients = make(map[string]*client)
	s1.conn, err = net.ListenUDP("udp", udpAddr1)
	checkError(err)

	// udpAddr2, err := net.ResolveUDPAddr("udp4", port2)
	// checkError(err)
	// var s2 server
	// s1.port = port2
	// s2.messages = make(chan string, 20)
	// s2.clients = make(map[string]*client)
	// s2.conn, err = net.ListenUDP("udp", udpAddr2)
	// checkError(err)

	go s1.sendMessage()
	// go s2.sendMessage()

	go func() {
		for {
			s1.handleMessage()
		}
	}()
	// go func() {
	// 	for {
	// 		s1.handleMessage()
	// 	}
	// }()
	time.Sleep(time.Second * 20000)
}

func (s *server) handleMessage() {
	var buf [128]byte

	n, addr, err := s.conn.ReadFromUDP(buf[0:])
	checkError(err)

	fmt.Println(string(buf[0:n]))
	msg := s.analyzeMessage(string(buf[0:n]))

	switch msg.rqnum {
	case "1": // REGISTER
		s.register(msg, addr)
	case "5": // PUBLISH
		s.publish(string(buf[0:n]))
	case "7": // INFORMResp
		s.informreq(string(buf[0:n]))
	case "10": // FINDResp
		s.findreq(string(buf[0:n]))
	case "20": // BYE
		s.exit(string(buf[0:n]))
	case "14":
		s.bye(string(buf[0:n]))
	case "16": // CHAT
		s.chat(string(buf[0:n]))
	case "17":
		s.mchat(string(buf[0:n]))
	}
}

func (s *server) exit(m string) {
	ss := strings.Split(m, "|")
	retmsg := fmt.Sprintf("EXIT|21|%s|left the channel", ss[2])
	s.messages <- retmsg
	delete(s.clients, ss[2])
}

func (s *server) bye(m string) {
	ss := strings.Split(m, "|")

	sen := strings.TrimSpace(ss[2])

	sender := s.clients[sen]
	partner := sender.chatpartner

	if sender.chatpartner == nil {
		return
	}

	fmt.Println("1")

	msg := "You are no longer chating with "

	n, err := s.conn.WriteToUDP([]byte(msg+partner.chatpartner.name), sender.chatpartner.userAddr)
	fmt.Println(n, err)
	n, err = s.conn.WriteToUDP([]byte(msg+sender.chatpartner.name), partner.chatpartner.userAddr)
	fmt.Println(n, err)

	sender.chatpartner = nil
	partner.chatpartner = nil
}

func (s *server) mchat(m string) {
	ss := strings.Split(m, "|")

	sen := strings.TrimSpace(ss[2])
	sender := s.clients[sen]

	msg := sender.name + ":" + ss[3]
	msg = strings.TrimSpace(msg)

	// fmt.Println("Sending to ", sender.chatpartner.name)
	// fmt.Println("Sender     ", sender.name)

	if sender.chatpartner == nil {
		n, err := s.conn.WriteToUDP([]byte("type:BYE"), sender.userAddr)
		fmt.Println(n, err)
	} else {
		n, err := s.conn.WriteToUDP([]byte(msg), sender.chatpartner.userAddr)
		fmt.Println(n, err)
	}

}

func (s *server) chat(m string) {
	ss := strings.Split(m, "|")

	tar := strings.TrimSpace(ss[2])
	sen := strings.TrimSpace(ss[3])

	targetclient, exists1 := s.clients[sen]
	senderclient, exists2 := s.clients[tar]

	if !exists1 || !exists2 {
		n, err := s.conn.WriteToUDP([]byte("This user does not exist, type:BYE"), targetclient.userAddr)
		fmt.Println(n, err)
		return
	}

	// if senderclient.chatpartner != targetclient {
	// 	n, err := s.conn.WriteToUDP([]byte("This user is already chatting with someone, type:BYE"), targetclient.userAddr)
	// 	fmt.Println(n, err)
	// 	return
	// }

	targetclient.chatpartner = senderclient
	senderclient.chatpartner = targetclient

	m1 := targetclient.name + " wants to chat with you!"
	n, err := s.conn.WriteToUDP([]byte(m1), senderclient.userAddr)
	fmt.Println(n, err)

	// m2 := "You are now chating with " + targetclient.name
	// n, err = s.conn.WriteToUDP([]byte(m2), senderclient.userAddr)
	// fmt.Println(n, err)
}

// func (s *server) bye(m string) {
// 	ss := strings.Split(m, "|")
// 	retmsg := fmt.Sprintf("BYE|15|%s|left the channel", ss[2])
// 	s.messages <- retmsg
// 	delete(s.clients, ss[2])
// }

func (s *server) findreq(m string) {
	ss := strings.Split(m, "|")
	infofrom, p := s.clients[ss[3]]
	if p == false {
		p := "nextserver"
		// if s.port == port1 {
		// 	p = port2
		// } else {
		// 	p = port1
		// }
		retmsg := fmt.Sprintf("REFER|13|%s", p)
		s.messages <- retmsg
		return
	}
	to := s.clients[ss[2]] // - should only send to this guy
	retmsg := ""
	fmt.Println(infofrom.userAddr.String())
	if infofrom.active {
		retmsg = fmt.Sprintf("FINDResp|11|%s|%s|%s|-1", infofrom.name, infofrom.userAddr.String(), s.port)
	} else {
		retmsg = fmt.Sprintf("FINDResp|11|%s|%s|%s|-2", infofrom.name, infofrom.userAddr.String(), s.port)
	}
	fmt.Println("Whitelist of ", infofrom.name, infofrom.whiteList[to.name])
	if _, ok := infofrom.whiteList[to.name]; !ok {
		retmsg = fmt.Sprintf("FINDDenied|12|%s|denied", infofrom.name)
	}
	s.messages <- retmsg
}

func (s *server) informreq(m string) {
	ss := strings.Split(m, "|")
	c := s.clients[ss[2]]
	retmsg := fmt.Sprintf("INFORMResp|9|%s|%s|%t|", c.name, c.userAddr, c.active)
	for k := range c.whiteList {
		retmsg += k + "-"
	}
	s.messages <- retmsg
}

func (s *server) publish(msg string) {
	ss := strings.Split(msg, "|")
	c := s.clients[ss[2]]
	if ss[4] == "true" {
		c.active = true
	}
	if ss[4] == "false" {
		c.active = false
	}

	retmsg := fmt.Sprintf("PUBLISHED|6|%s|%s|%t|", c.name, ss[3], c.active)
	if len(ss[5]) != 0 {
		names := strings.Split(ss[5], " ")
		for _, name := range names {
			retmsg += name + " "
			c.whiteList[name] = 0
		}
		fmt.Println(c.whiteList)
	}

	if len(ss) > 6 {
		rm := "UNPUBLISHED|7 - Invalid input"
		fmt.Println(rm)
		s.messages <- rm
		return
	}

	// client := s.clients[ss[2]]
	//
	// n, err := s.conn.WriteToUDP([]byte("people on channel are notified"), client.userAddr)
	// fmt.Println(n, err)

	s.messages <- retmsg
}

func (s *server) register(msg message, addr *net.UDPAddr) {
	if _, ok := s.clients[msg.body[0]]; !ok && len(s.clients) < 5 {
		var c client
		c.userAddr = addr
		c.name = msg.body[0]
		c.whiteList = make(map[string]int)
		s.clients[msg.body[0]] = &c
		s.messages <- fmt.Sprintf("REGISTERED|2|%s with server", c.name)
	} else {
		other := ""
		if s.port == port {
			other = "nextserver"
		} else {
			other = "previousserver"
		}
		msg := "REGISTER-DENIED|3|This server is at capacity, register with server " + other
		n, err := s.conn.WriteToUDP([]byte(msg), addr)
		fmt.Println(n, err)
	}
}

func (s *server) analyzeMessage(msg string) (m message) {

	mbreakdown := strings.Split(msg, "|")

	if len(mbreakdown) == 2 {
		m.command = mbreakdown[0]
		m.rqnum = mbreakdown[1]
	} else {
		m.command = mbreakdown[0]
		m.rqnum = mbreakdown[1]
		m.body = mbreakdown[2:]
	}
	return
}

func (s *server) sendMessage() {
	for {
		msg := <-s.messages
		// daytime := time.Now().String()
		// sendstr := msg + daytime
		fmt.Println("Outgoing : ", msg)
		f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if _, err = f.WriteString(msg + "\n"); err != nil {
			panic(err)
		}
		for _, c := range s.clients {
			fmt.Println(c.name)
			n, err := s.conn.WriteToUDP([]byte(msg), c.userAddr)
			fmt.Println(n, err)
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error:%s", err.Error())
		os.Exit(1)
	}
}
