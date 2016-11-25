package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func checkError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

type client struct {
	name string
	port int
}

func main() {
	/* Lets prepare a address at any address at port 10001*/
	ServerAddr, err := net.ResolveUDPAddr("udp", ":10001")
	checkError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	checkError(err)
	defer ServerConn.Close()

	buf := make([]byte, 128)

	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		//fmt.Println("Received ", string(buf[0:n]), " from ", addr)
		msg := strings.Split(string(buf[0:n]), " ")
		//fmt.Println(name, port)
		switch msg[0] {
		case "REGISTER":
			//newClient := client{msg[1], addr.Port}
			Conn, _ := net.DialUDP("udp", ServerAddr, addr)
			msg := "REGISTERED \n"
			rep := []byte(msg)
			_, err = Conn.Write(rep)
			if err != nil {
				fmt.Println(err)
			}
			defer Conn.Close()
		}

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}
