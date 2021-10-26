package main

import (
    "bufio"
    "fmt"
    "os"
)

func main() {
    fmt.Fprintln(os.Stderr, "Err test")
    fmt.Println("Read test")
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        fmt.Println(scanner.Text()) // Println will add back the final '\n'
    }
    if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "reading standard input:", err)
    }
}
