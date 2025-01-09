package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"syscall"
	"unsafe"
)

func exit_shell(){
	fmt.Println("")
	os.Exit(0)
}

func read_line() []byte {
	var inputBuf []byte = make([]byte, 1)
	var input_line []byte = make([]byte, 0, 20)

	for {
		_, err := os.Stdin.Read(inputBuf)

		if err == io.EOF {
			exit_shell()
		}

		input := inputBuf[0]

		// Transform input
		switch input {
		case '\r':
			input = '\n'
		case 0x7f:
			input = '\b'
		case 0x1b: // \e
			fmt.Print("^")
			input = '['
		case 0x4:
			exit_shell()
		}

		if input == '\b' {
			fmt.Print("\b ")
		}

		if input != '\b' || len(input_line) > 0 {
			// echo character back to user
			fmt.Printf("%c", input)
		}

		if input == '\n' {
			break
		}

		if input == '\b' {
			if len(input_line) > 0 {
				input_line = input_line[:len(input_line)-1]
			}
		} else {
			input_line = append(input_line, input)
		}
	}
	return input_line
}

func main() {
	for {
		fmt.Print("> ")
		line := read_line()
		if len(line) == 0 {
			continue
		}

		if bytes.EqualFold(line, []byte("exit")) {
			exit_shell()
		}

		if line[0] != '/' {
			line = append([]byte("/usr/"), line...)
		}

		r, _, _ := syscall.RawSyscall(syscall.SYS_EXECVE,
			uintptr(unsafe.Pointer(&line[0])),
			uintptr(unsafe.Pointer(nil)),
			uintptr(unsafe.Pointer(nil)))

		pid := int(r)
		if pid <= 0 {
			fmt.Printf("Command not found: %s\n", line)
		} else {
			syscall.Wait4(pid, nil, 0, nil)
		}
	}
}