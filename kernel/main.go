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
var errorWriters = []io.Writer{&serialDevice, text_mode_error_writer{}}
var logWriters = []io.Writer{&serialDevice, text_mode_writer{}}

const ENABLE_DEBUG = false

//go:linkname kmain main.main
func kmain(info *MultibootInfo, stackstart uintptr, stackend uintptr) {
	if stackstart <= stackend {
		kernelPanic("No stack")
	}
	text_mode_init()
	InitSerialDevice()
	log.SetDefaultDebugLogWriters(debugWriters[:])
	log.SetDefaultErrorLogWriters(errorWriters[:])
	log.SetDefaultLogWriters(logWriters[:])

	text_mode_flush_screen()
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

	SetInterruptHandler(0xd, gpfPanic, KCS_SELECTOR, PRIV_USER)

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

	text_mode_println_col("Initilaization complete", 0x2)
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

func gpfPanic() {
	log.KErrorLn("\nReceived General Protection fault. Disabling Interrupts and halting")
	log.KPrintLn("Errorcode: ", uintptr(currentThread.info.ExceptionCode))
	panicHelper(currentThread)
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

func panicHelper(thread *thread) {
	log.KPrintLn("Domain ID: ", thread.domain.pid, ", Thread ID: ", thread.tid)
	log.KPrintLn("Program name: ", thread.domain.programName)
	if thread.isKernelInterrupt {
		log.KPrint("In kernel function: ")
		printFuncName(thread.kernelInfo.EIP)
	} else {
		log.KPrintLn("In user function: ", thread.info.EIP)
	}
	printThreadRegisters(thread)
	DisableInterrupts()
	Hlt()
}

// wrapper for do_kernelPanic that gets the return address
// and pushers it on the stack and then calls do_kernelPanic
func kernelPanic(msg string)

//go:nosplit
func do_kernelPanic(caller uintptr, msg string) {
	log.KErrorLn("\n", msg, " - kernel panic :(")
	log.KPrint("Called from function: ")
	printFuncName(caller - 4) // account for the fact that caller points to the instruction after the call
	if currentThread != nil {
		panicHelper(currentThread)
	} else {
		log.KErrorLn("Cannot print registers. 'currentThread' is nil")
	}
	DisableInterrupts()
	Hlt()
	// does not return
}

func printThreadRegisters(t *thread) {
	log.KPrint("User regs:          Kernel regs:\n")
	f := runtimeFindFunc(uintptr(t.kernelInfo.EIP))
	log.KPrint("EIP: ", t.info.EIP, "      ", "EIP: ", t.kernelInfo.EIP, " ", f._Func().Name(), "\n")
	//rintRegisterLineInfo("EIP: ", t.info.EIP, t.kernelInfo.EIP, f._Func().Name())
	printRegisterLine(20, "ESP: ", t.info.ESP, t.kernelInfo.ESP)
	printRegisterLine(20, "EBP: ", t.regs.EBP, t.kernelRegs.EBP)
	printRegisterLine(20, "EAX: ", t.regs.EAX, t.kernelRegs.EAX)
	printRegisterLine(20, "EBX: ", t.regs.EBX, t.kernelRegs.EBX)
	printRegisterLine(20, "ECX: ", t.regs.ECX, t.kernelRegs.ECX)
	printRegisterLine(20, "EDX: ", t.regs.EDX, t.kernelRegs.EDX)
	printRegisterLine(20, "ESI: ", t.regs.ESI, t.kernelRegs.ESI)
	printRegisterLine(20, "EDI: ", t.regs.EDI, t.kernelRegs.EDI)
	printRegisterLine(20, "EFLAGS: ", t.info.EFLAGS, t.kernelInfo.EFLAGS)
	printRegisterLine(20, "Exception: ", t.info.ExceptionCode, t.kernelInfo.ExceptionCode)
	printRegisterLine(20, "Interrupt: ", t.info.InterruptNumber, t.kernelInfo.InterruptNumber)
	printRegisterLine(20, "Krn ESP: ", t.regs.KernelESP, t.kernelRegs.KernelESP)
}

func printRegisterLine(tablength int, label string, userReg, kernelReg uint32) {
	firstLength := len(label)
	log.KPrint(label, uintptr(userReg))
	// pad number
	firstLength += 3 // account for the hexadecimal 0x#
	for i, n := firstLength, userReg>>4; i < 20; i, n = i+1, n>>4 {
		if n == 0 {
			log.KPrint(" ")
		}
	}
	log.KPrint(label, uintptr(kernelReg), "\n")
}

func printRegisters(info *InterruptInfo, regs *RegisterState) {
	log.KPrintLn("Interrupt: ", uintptr(info.InterruptNumber))
	log.KPrintLn("Exception: ", uintptr(info.ExceptionCode))
	log.KPrintLn("EIP: ", uintptr(info.EIP))
	log.KPrintLn("CS: ", uintptr(info.CS))
	log.KPrintLn("EFLAGS: ", uintptr(info.EFLAGS))
	log.KPrintLn("ESP: ", uintptr(info.ESP))
	log.KPrintLn("SS: ", uintptr(info.SS))
	log.KPrintLn("-----------")
	log.KPrintLn("GS: ", uintptr(regs.GS))
	log.KPrintLn("FS: ", uintptr(regs.FS))
	log.KPrintLn("ES: ", uintptr(regs.ES))
	log.KPrintLn("DS: ", uintptr(regs.DS))
	log.KPrintLn("EBP: ", uintptr(regs.EBP))
	log.KPrintLn("EAX: ", uintptr(regs.EAX))
	log.KPrintLn("EBX: ", uintptr(regs.EBX))
	log.KPrintLn("ECX: ", uintptr(regs.ECX))
	log.KPrintLn("EDX: ", uintptr(regs.EDX))
	log.KPrintLn("ESI: ", uintptr(regs.ESI))
	log.KPrintLn("EDI: ", uintptr(regs.EDI))
	log.KPrintLn("KernelESP", uintptr(regs.KernelESP))
}

func printTid(t *thread) {
	log.KPrintLn("Pid: ", t.domain.pid, ", Tid: ", t.tid)
}

func delay(v int) {
	for i := 0; i < 684000; i++ {
		for j := 0; j < v; j++ {
		}
	}
}

func cstring(ptr uintptr) string {
	var n int
	for p := ptr; *(*byte)(unsafe.Pointer(p)) != 0; p++ {
		n++
	}
	return unsafe.String((*byte)(unsafe.Pointer(ptr)), n)
}
