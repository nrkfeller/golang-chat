package main

import "fmt"

func main() {
	ch := make(chan int, 1)
	done := make(chan bool)
	go fib(ch, done)

	<-done

	fmt.Println(<-ch)
}

func fib(ch chan int, done chan bool) {
	i, j := 0, 1
	ch <- j
	for k := 0; k < 202; k++ {
		j = <-ch
		i, j = j, i+j
		ch <- j
	}
	done <- true
}
