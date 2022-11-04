package main

import (
    "fmt"
    //"os/exec"
    //"os"
    "syscall"
)

func main() {
    fmt.Println("Hello go world")
    //cmd := exec.Command("/usr/helloc")
    //cmd.Start()
    //_, err := os.StartProcess("/usr/helloc", []string{}, &os.ProcAttr{
	//	Dir:   "",
	//	Files: []*os.File{},
	//	Env:   []string{},
	//	Sys:   nil,
	//})
    var err error
    err = syscall.Exec("/usr/helloc", []string{}, []string{})
	if err != nil {
        fmt.Println(err)
	}
    err = syscall.Exec("/usr/hellocxx", []string{}, []string{})
	if err != nil {
        fmt.Println(err)
	}
    err = syscall.Exec("/usr/hellorust", []string{}, []string{})
	if err != nil {
        fmt.Println(err)
	}
    err = syscall.Exec("/usr/statx", []string{}, []string{})
	if err != nil {
        fmt.Println(err)
	}
    defer fmt.Println("I can use go features")
    //go fmt.Println("1")
    //go fmt.Println("2")
    //go fmt.Println("3")
    //go fmt.Println("4")
    fmt.Println("Bye go world")
    //test()

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
