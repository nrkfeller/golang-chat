package main

import (
	"io"
	"net"
)

func main() {
	listenner, _ := net.Listen("tcp", "localhost:4000")

	for {
		conn, _ := listenner.Accept()
		go io.Copy(conn, conn)
	}
}
