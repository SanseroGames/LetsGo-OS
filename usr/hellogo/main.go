package main

import (
    "fmt"
)

func main() {
    defer fmt.Println("I can use go features")
    go fmt.Println("1")
    go fmt.Println("2")
    go fmt.Println("3")
    go fmt.Println("4")
    fmt.Println("Hello go world")
    test()
}

func sum(s []int, c chan int) {
	sum := 0
	for _, v := range s {
		sum += v
	}
	c <- sum // send sum to c
}

func test() {
	s := []int{7, 2, 8, -9, 4, 0}

	c := make(chan int)
	go sum(s[:len(s)/2], c)
	go sum(s[len(s)/2:], c)
	x, y := <-c, <-c // receive from c

	fmt.Println(x, y, x+y)
}
