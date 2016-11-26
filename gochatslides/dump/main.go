package main

import (
	"fmt"
	"net"
)

func main() {

	udpAddr1, _ := net.ResolveUDPAddr("udp4", ":1203")
	conn, _ := net.ListenUDP("udp", udpAddr1)

	fmt.Println(conn.RemoteAddr())

}
