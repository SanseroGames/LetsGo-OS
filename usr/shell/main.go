package main

import (
    "fmt"
    "os"
    "bufio"
)

func main() {
    print("Hi from shell\n")
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
	    fmt.Println(scanner.Text())
    }

    if err := scanner.Err(); err != nil {
	    fmt.Println(err)
    }
}
