package main

import (
	"fmt"
	"strings"
)

func main() {

	s := "hello potato tomate "
	ss := strings.Split(s, " ")
	fmt.Println(ss, len(ss)-1)
	ss = ss[:len(ss)-1]
	fmt.Println(ss, len(ss)-1)

}
