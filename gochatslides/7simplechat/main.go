package main

import (
	"fmt"
	"io"
	"net"
)

var partner = make(chan io.ReadWriteCloser)

func main() {
	listenner, _ := net.Listen("tcp", "localhost:4000")

	for {
		conn, _ := listenner.Accept()
		go match(conn)
		go match(conn)
	}
}

func match(c io.ReadWriteCloser) {
	fmt.Fprint(c, "Waiting for a partner...")
	select {
	case partner <- c:
	case p := <-partner:
		chat(p, c)
	}
}

func chat(a, b io.ReadWriteCloser) {
	fmt.Fprintln(a, "Found one! Say hi.")
	fmt.Fprintln(b, "Found one! Say hi.")
	go io.Copy(a, b)
	io.Copy(b, a)
}
