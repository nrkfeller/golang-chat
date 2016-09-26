package main

import (
	"io"
	"net"
)

func main() {
	l, _ := net.Listen("tcp", "localhost:4000")

	for {
		connection, _ := l.Accept()
		io.Copy(connection, connection) // This is like self chat
		// notice the connection does not get closed
	}
}
