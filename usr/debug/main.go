package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func main() {
	if len(os.Args) <= 1 {
		return
	}

	args := os.Args[1:]
	cmd := args[0]

	binPath := cmd
	if cmd[0] != '/' {
		binPath = "/usr/" + cmd
	}

	argv0p, err := syscall.BytePtrFromString(string(binPath))
	if err != nil {
		fmt.Printf("Failed to parse cmd")
		return
	}
	argvp, err := syscall.SlicePtrFromStrings(args)
	if err != nil {
		fmt.Printf("Failed to parse argv")
		return
	}
	envvp, err := syscall.SlicePtrFromStrings(os.Environ())
	if err != nil {
		fmt.Printf("Failed to parse envp")
		return
	}

	r, _, _ := syscall.RawSyscall6(syscall.SYS_EXECVE,
		uintptr(unsafe.Pointer(argv0p)),
		uintptr(unsafe.Pointer(&argvp[0])),
		uintptr(unsafe.Pointer(&envvp[0])),
		uintptr(1),
		0, 0)

	pid := int(r)
	if pid <= 0 {
		fmt.Printf("Command not found: %s\n", cmd)
	} else {
		syscall.Wait4(pid, nil, 0, nil)
	}
}
