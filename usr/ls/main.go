package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	entries, err := os.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range entries {
		fmt.Println(e.Name())
	}
}
