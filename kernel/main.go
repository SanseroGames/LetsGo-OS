package kernel

import (
	"io"
	"path"
	"unsafe"

	"github.com/sanserogames/letsgo-os/kernel/log"
)

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

var domains [len(progs)]*domain
var threads [len(progs)]*thread

var debugWriters = []io.Writer{&serialDevice}
var errorWriters = []io.Writer{&serialDevice, TextModeErrorWriter{}}
var logWriters = []io.Writer{&serialDevice, TextModeWriter{}}

const ENABLE_DEBUG = false

//go:linkname kmain main.main
func kmain(info *MultibootInfo, stackstart uintptr, stackend uintptr) {
	if stackstart <= stackend {
		kernelPanic("No stack")
	}
	TextModeInit()
	InitSerialDevice()
	log.SetDefaultDebugLogWriters(debugWriters[:])
	log.SetDefaultErrorLogWriters(errorWriters[:])
	log.SetDefaultLogWriters(logWriters[:])

	TextModeFlushScreen()
	s := "Hi and welcome to Let's-Go OS"
	s2 := "An Operating System written in the GO programming language"
	s3 := "It can't do much, but I hope you enjoy your stay"

	log.KPrintLn(s)
	log.KPrintLn(s2)
	log.KPrintLn(s3)
	log.KPrintLn("")
	log.KDebugLn("Starting initialization...")

	InitSegments()

	InitInterrupts()

	InitSyscall()

	InitPIC()

	InitPit()
	InitKeyboard()
	InitATA()
	InitSerialDeviceInterrupt()

	InitMultiboot(info)
	//printMemMaps()

	InitPaging()

	InitUserMode(stackstart, stackend)

	textModePrintLnCol("Initilaization complete", 0x2)
	log.KDebugLn("Initialization complete")
	//HdReadSector()

	var err int
	for i := 0; i < len(progs); i++ {
		newDomainMem := AllocPage()
		Memclr(newDomainMem, PAGE_SIZE)
		newDomain := (*domain)(unsafe.Pointer(newDomainMem))
		newThreadMem := AllocPage()
		Memclr(newThreadMem, PAGE_SIZE)
		newThread := (*thread)(unsafe.Pointer(newThreadMem))
		err = StartProgram(progs[i], newDomain, newThread)
		if err != 0 {
			kernelPanic("Could not start program")
		}
		newDomain.MemorySpace.MapPage(newThreadMem, newThreadMem, PAGE_RW|PAGE_PERM_KERNEL)
		AddDomain(newDomain)
	}

	log.KPrintLn("domain size: ", uint(unsafe.Sizeof(domains[0])),
		" thread_size: ", uint(unsafe.Sizeof(threads[0])),
		" total: ", uint(unsafe.Sizeof(domains[0])+unsafe.Sizeof(threads[0])))
	log.KPrintLn("stack start: ", uintptr(scheduleThread.kernelStack.hi),
		" stack end: ", uintptr(scheduleThread.kernelStack.lo))
	log.KPrintLn("info: ", uint(unsafe.Sizeof(currentThread.info)),
		" regs: ", uint(unsafe.Sizeof(currentThread.regs)))
	if currentThread == nil {
		kernelPanic("I expect AddDomain to set currentThread variable")
	}
	kernelThreadInit()
	kernelPanic("Could not jump to user space :/")

}

func printFuncName(pc uintptr) {
	f := runtimeFindFunc(pc)
	if !f.valid() {
		log.KPrintLn("func: ", pc)
		return
	}
	s := f._Func().Name()
	file, line := f._Func().FileLine(pc)
	_, filename := path.Split(file)
	log.KPrintLn(s, " (", filename, ":", line, ")")
}

func printTid(t *thread) {
	log.KPrintLn("Pid: ", t.domain.pid, ", Tid: ", t.tid)
}

func cstring(ptr uintptr) string {
	var n int
	for p := ptr; *(*byte)(unsafe.Pointer(p)) != 0; p++ {
		n++
	}
	return unsafe.String((*byte)(unsafe.Pointer(ptr)), n)
}
