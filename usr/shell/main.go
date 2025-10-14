package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"syscall"
	"unsafe"
)

func exitShell() {
	fmt.Println("")
	os.Exit(0)
}

func readLine() []byte {
	var inputBuf []byte = make([]byte, 1)
	var input_line []byte = make([]byte, 0, 20)

	for {
		_, err := os.Stdin.Read(inputBuf)

		if err == io.EOF {
			exitShell()
		}

		input := inputBuf[0]

		// Transform input
		switch input {
		case '\r':
			input = '\n'
		case 0x7f:
			input = '\b'
		case 0x1b: // \e
			input_line = append(input_line, '^')
			fmt.Print("^")
			input = '['
		case 0x4:
			exitShell()
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
	os.Setenv("PWD", "/usr")
	os.Setenv("NAME", "Let'sGo OS!")
	os.Setenv("SHELL", "/usr/shell")
	os.Setenv("HOSTTYPE", "x86")
	for {
		fmt.Print("> ")
		line := readLine()
		args := bytes.Fields(line)
		if len(args) == 0 {
			continue
		}

		cmd := args[0]

		if bytes.EqualFold(cmd, []byte("exit")) {
			exitShell()
		}

		binPath := cmd
		if cmd[0] != '/' {
			binPath = append([]byte("/usr/"), cmd...)
		}

		argv := make([]string, len(args))
		for i := range len(args) {
			argv[i] = string(args[i])
		}

		argv0p, err := syscall.BytePtrFromString(string(binPath))
		if err != nil {
			fmt.Printf("Failed to parse cmd")
			continue
		}
		argvp, err := syscall.SlicePtrFromStrings(argv)
		if err != nil {
			fmt.Printf("Failed to parse argv")
			continue
		}
		envvp, err := syscall.SlicePtrFromStrings(os.Environ())
		if err != nil {
			fmt.Printf("Failed to parse envp")
			continue
		}

		r, _, _ := syscall.RawSyscall(syscall.SYS_EXECVE,
			uintptr(unsafe.Pointer(argv0p)),
			uintptr(unsafe.Pointer(&argvp[0])),
			uintptr(unsafe.Pointer(&envvp[0])))

		pid := int(r)
		if pid <= 0 {
			fmt.Printf("Command not found: %s\n", cmd)
		} else {
			syscall.Wait4(pid, nil, 0, nil)
		}
	}
}
