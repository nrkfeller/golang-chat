package main

import (
	"fmt"
	"net"
)

func main() {
	listenner, _ := net.Listen("tcp", "localhost:4000")

	for {
		connection, _ := listenner.Accept() // connection is a net.Conn which is an io.Writer so Fprintln can write to it
		fmt.Fprintln(connection, "hi man")
		connection.Close()
	}
}
