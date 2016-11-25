package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

//Server is one of two servers
type Server struct {
	conn     *net.UDPConn //
	messages chan string  //接受到的消息
	clients  map[int]Client
}

// Client are users
type Client struct {
	userID   int
	userName string
	userAddr *net.UDPAddr
}

// Message passed by users
type Message struct {
	status   int
	userID   int
	userName string
	content  string
}

func (s *Server) handleMessage() {
	var buf [512]byte

	n, addr, err := s.conn.ReadFromUDP(buf[0:])
	if err != nil {
		return
	}

	//分析消息
	msg := string(buf[0:n])
	//fmt.Println(msg)
	m := s.analyzeMessage(msg)
	switch m.status {
	//进入聊天室消息
	case 1:
		var c Client
		c.userAddr = addr
		c.userID = m.userID
		c.userName = m.userName
		s.clients[c.userID] = c //添加用户
		s.messages <- fmt.Sprintln("REGISTERED with server", s.conn.LocalAddr())
	//用户发送消息
	case 2:
		s.messages <- msg
	//client发来的退出消息
	case 3:
		delete(s.clients, m.userID)
		s.messages <- msg
	case 4:
		sendstr := "REGISTER-DENIED server is full"
		s.conn.WriteToUDP([]byte(sendstr), addr)
	default:
		fmt.Println("Unrecognized", msg)
	}

	//fmt.Println(n,addr,string(buf[0:n]))

}

//这里还要判断一下数组的长度，
func (s *Server) analyzeMessage(msg string) (m Message) {

	s2 := strings.Split(msg, "##")

	switch s2[0] {
	case "1":
		// MAKE SURE MAX 5
		fmt.Println("How many clients", len(s.clients))
		if len(s.clients) == 5 {
			m.status = 4
			m.userID, _ = strconv.Atoi(s2[1])
			m.userName = s2[2]
			return
		}
		m.status, _ = strconv.Atoi(s2[0])
		m.userID, _ = strconv.Atoi(s2[1])
		m.userName = s2[2]
		//fmt.Println(m)
		return
	case "2":
		m.status, _ = strconv.Atoi(s2[0])
		m.userID, _ = strconv.Atoi(s2[1])
		m.content = s2[2]
		return
	case "3":
		m.status, _ = strconv.Atoi(s2[0])
		m.userID, _ = strconv.Atoi(s2[1])
		return
	default:
		fmt.Println("Unrecognized", msg)
		return
	}
}
func (s *Server) sendMessage() {
	for {
		msg := <-s.messages
		daytime := time.Now().String()
		sendstr := msg + daytime
		fmt.Println(00, sendstr)
		for _, c := range s.clients {
			//fmt.Println(c)
			n, err := s.conn.WriteToUDP([]byte(sendstr), c.userAddr)
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

func main() {
	if len(os.Args) != 2 {
		os.Exit(0)
	}
	udpAddr, err := net.ResolveUDPAddr("udp4", ":"+os.Args[1])
	checkError(err)

	var s1 Server
	s1.messages = make(chan string, 20)
	s1.clients = make(map[int]Client, 0)

	s1.conn, err = net.ListenUDP("udp", udpAddr)
	checkError(err)

	go s1.sendMessage()

	for {
		s1.handleMessage()
	}
}
