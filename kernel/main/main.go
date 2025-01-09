package main

import (
	_ "github.com/sanserogames/letsgo-os/kernel"
)

// This file exists so that there is a main.main symbol
// This will cause the go compiler to emit an ELF file instead of a go archive
// I need an ELF file to build the kernel
// This method will be linked in the kernel
func main()
