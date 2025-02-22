// This package exists so that there is a main.main symbol
// This will cause the go compiler to emit an ELF file instead of a go archive
// I need an ELF file to build the kernel
package main

import (
	"io"
	"unsafe"

	"github.com/sanserogames/letsgo-os/kernel"
	"github.com/sanserogames/letsgo-os/kernel/log"
	"github.com/sanserogames/letsgo-os/kernel/mm"
	"github.com/sanserogames/letsgo-os/kernel/panic"
)

// This method will be linked in the kernel
func main()



var progs = [...]string{
	// "/usr/cread",
	// "/usr/helloc",
	// "/usr/hellocxx",
	// "/usr/hellogo",
	// "/usr/hellorust",
	// "/usr/readtest",
	// "/usr/rustread",
	"/usr/shell",
	// "/usr/statx",

	// "/usr/syscall-test",
}

var debugWriters = []io.Writer{&kernel.SerialDevice}
var errorWriters = []io.Writer{&kernel.SerialDevice, kernel.TextModeErrorWriter{}}
var logWriters = []io.Writer{&kernel.SerialDevice, kernel.TextModeWriter{}}


//go:linkname kmain main.main
func kmain(info *kernel.MultibootInfo, stackstart uintptr, stackend uintptr) {
	if stackstart <= stackend {
		// This will fail as printing is not yet initialized
		panic.KernelPanic("No stack")
	}
	kernel.TextModeInit()
	kernel.InitSerialDevice()
	log.SetDefaultDebugLogWriters(debugWriters[:])
	log.SetDefaultErrorLogWriters(errorWriters[:])
	log.SetDefaultLogWriters(logWriters[:])

	kernel.TextModeFlushScreen()
	s := "Hi and welcome to Let's-Go OS"
	s2 := "An Operating System written in the GO programming language"
	s3 := "It can't do much, but I hope you enjoy your stay"

	log.KPrintLn(s)
	log.KPrintLn(s2)
	log.KPrintLn(s3)
	log.KPrintLn("")
	log.KDebugLn("Starting initialization...")

	kernel.InitSegments()

	kernel.InitInterrupts()

	kernel.InitSyscall()

	kernel.InitPIC()

	kernel.InitPit()
	kernel.InitKeyboard()
	kernel.InitATA()
	kernel.InitSerialDeviceInterrupt()

	kernel.InitMultiboot(info)
	//printMemMaps()

	kernel.InitPaging()

	kernel.InitUserMode(stackstart, stackend)

	kernel.TextModePrintLnCol("Initilaization complete", 0x2)
	log.KDebugLn("Initialization complete")
	//HdReadSector()

	var err int
	for i := 0; i < len(progs); i++ {
		newDomainMem := mm.AllocPage()
		mm.Memclr(newDomainMem, mm.PAGE_SIZE)
		newDomain := (*kernel.Domain)(unsafe.Pointer(newDomainMem))
		newThreadMem := mm.AllocPage()
		mm.Memclr(newThreadMem, mm.PAGE_SIZE)
		newThread := (*kernel.Thread)(unsafe.Pointer(newThreadMem))
		err = kernel.StartProgram(progs[i], newDomain, newThread)
		if err != 0 {
			panic.KernelPanic("Could not start program")
		}
		newDomain.MemorySpace.MapPage(newThreadMem, newThreadMem, mm.PAGE_RW|mm.PAGE_PERM_KERNEL)
		kernel.AddDomain(newDomain)
	}

	if kernel.CurrentThread == nil {
		panic.KernelPanic("I expect AddDomain to set currentThread variable")
	}
	kernel.KernelThreadInit()
	panic.KernelPanic("Could not jump to user space :/")
}