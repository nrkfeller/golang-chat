package main

import (
	"fmt"
	"time"
)

func main() {
	go say("sup", 3)
	go say("hi", 2)
	go say("allo", 1)

	time.Sleep(time.Second * 4)
}

func say(msg string, secs int) {
	time.Sleep(time.Duration(secs) * time.Second)
	fmt.Println(msg)
}
