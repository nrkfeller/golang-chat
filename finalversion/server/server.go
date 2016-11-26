package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type server struct {
	conn     *net.UDPConn //
	messages chan string  //接受到的消息
	clients  map[string]client
	port     string
}

type client struct {
	ID        int
	name      string
	on        bool
	userAddr  *net.UDPAddr
	whiteList map[string]int
}

type message struct {
	command string
	rqnum   string
	body    []string
}

const (
	port1 string = ":1200"
	port2 string = ":1201"
)

func main() {

	udpAddr1, err := net.ResolveUDPAddr("udp4", port1)
	checkError(err)
	var s1 server
	s1.port = port1
	s1.messages = make(chan string, 20)
	s1.clients = make(map[string]client)
	s1.conn, err = net.ListenUDP("udp", udpAddr1)
	checkError(err)

	udpAddr2, err := net.ResolveUDPAddr("udp4", port2)
	checkError(err)
	var s2 server
	s1.port = port2
	s2.messages = make(chan string, 20)
	s2.clients = make(map[string]client)
	s2.conn, err = net.ListenUDP("udp", udpAddr2)
	checkError(err)

	go s1.sendMessage()
	go s2.sendMessage()

	go func() {
		for {
			s2.handleMessage()
		}
	}()
	go func() {
		for {
			s1.handleMessage()
		}
	}()
	time.Sleep(time.Second * 20000)
}

func (s *server) handleMessage() {
	var buf [512]byte

	n, addr, err := s.conn.ReadFromUDP(buf[0:])
	checkError(err)

	fmt.Println(string(buf[0:n]))
	msg := s.analyzeMessage(string(buf[0:n]))

	switch msg.rqnum {
	case "1": // REGISTER
		s.register(msg, addr)
	case "5": // PUBLISH
		s.publish(msg)
	case "7": // INFORMResp
		s.informreq(msg)
	case "10": // FINDResp
		s.findreq(string(buf[0:n]))
	case "14":
		s.bye(string(buf[0:n]))
	}
}

func (s *server) bye(m string) {
	ss := strings.Split(m, "|")
	retmsg := fmt.Sprintf("BYE|13|%s|left the channel", ss[2])
	s.messages <- retmsg
	delete(s.clients, ss[2])
}

func (s *server) findreq(m string) {
	ss := strings.Split(m, "|")
	infofrom, p := s.clients[ss[3]]
	if p == false {
		p := ""
		if s.port == port1 {
			p = port2
		} else {
			p = port1
		}
		retmsg := fmt.Sprintf("REFER|13|%s|%s", s.conn.RemoteAddr().String(), p)
		s.messages <- retmsg
		return
	}
	to := s.clients[ss[2]] // - should only send to this guy
	retmsg := ""
	if infofrom.on {
		retmsg = fmt.Sprintf("FINDResp|11|%s|%s|%s|-1", infofrom.name, infofrom.userAddr, s.conn.RemoteAddr().String())
	} else {
		retmsg = fmt.Sprintf("FINDResp|11|%s|%s|%s|-2", infofrom.name, infofrom.userAddr, s.conn.RemoteAddr().String())
	}
	if _, ok := infofrom.whiteList[to.name]; !ok {
		retmsg = fmt.Sprintf("FINDDenied|12|%s|denied", infofrom.name)
	}
	s.messages <- retmsg
}

func (s *server) informreq(msg message) {
	c := s.clients[msg.body[2]]
	retmsg := fmt.Sprintf("INFORMResp|9|%s|%s|%t|", c.name, c.userAddr, c.on)
	for k := range c.whiteList {
		retmsg += k + "-"
	}
	s.messages <- retmsg
}

func (s *server) publish(msg message) {
	if len(msg.body[3]) != 0 {
		retmsg := fmt.Sprintf("PUBLISHED|6|%s|%s|%s|", msg.body[0], msg.body[1], msg.body[2])

		for _, v := range strings.Split(msg.body[3], " ") {
			retmsg += v + ""
			s.clients[msg.body[0]].whiteList[v] = 0
		}
		s.messages <- retmsg
	} else {
		rm := "UNPUBLISHED|7 - Invalid input"
		fmt.Println(rm)
		s.messages <- rm
	}
}

func (s *server) register(msg message, addr *net.UDPAddr) {
	if _, ok := s.clients[msg.body[0]]; !ok && len(s.clients) < 5 {
		var c client
		c.userAddr = addr
		c.name = msg.body[0]
		c.whiteList = make(map[string]int)
		s.clients[msg.body[0]] = c
		s.messages <- fmt.Sprintf("REGISTERED|2|%s with server", c.name)
	} else {
		// register with other server
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
		for _, c := range s.clients {
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
