package main

import (
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"
)

func main() {
	http.Handle("/", websocket.Handler(handler))
	http.ListenAndServe("localhost:4000", nil)
}

func handler(c *websocket.Conn) {
	var s string
	fmt.Fscan(c, &s)
	fmt.Println("Received", s)
	fmt.Fprint(c, "Wie gehts!")
}
