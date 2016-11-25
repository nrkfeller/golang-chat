package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func main() {
	clientport := rand.Intn(200) + 10010

	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	checkError(err)

	fmt.Println("Your local address is ", strconv.Itoa(clientport))

	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+strconv.Itoa(clientport))
	checkError(err)

	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	checkError(err)

	defer Conn.Close()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")
		text, _ := reader.ReadString('\n')
		buf := []byte(text)
		_, err := Conn.Write(buf)
		if err != nil {
			fmt.Println(text, err)
		}
	}
}
