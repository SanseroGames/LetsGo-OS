package main

import (
    "fmt"
    "syscall"
)

func main() {
    fmt.Println("Starting Bomg")
    var err error
    err = syscall.Exec("/usr/bomb", []string{}, []string{})
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println("Done bombing")
}
