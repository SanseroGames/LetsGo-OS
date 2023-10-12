package main

import (
	"io"
	"path"
	"reflect"
	"unsafe"
)

var progs = [...]string{
	"/usr/hacker",
}

var domains [len(progs)]*domain
var threads [len(progs)]*thread

func main()

const ENABLE_DEBUG = true

//go:linkname kmain main.main
func kmain(info *MultibootInfo, stackstart uintptr, stackend uintptr) {
	if stackstart <= stackend {
		kernelPanic("No stack")
	}
	text_mode_init()
	InitSerialDevice()

	text_mode_flush_screen()
	s := "Hi and welcome to Let's-Go OS"
	s2 := "An Operating System written in the GO programming language"
	s3 := "It can't do much, but I hope you enjoy your stay"

	kprintln(s)
	kprintln(s2)
	kprintln(s3)
	kprintln("")
	kdebugln("Starting initialization...")

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
	kdebugln("Initialization complete")
	//HdReadSector()

	var err int
	for i := 0; i < len(progs); i++ {
		newDomainMem := allocPage()
		Memclr(newDomainMem, PAGE_SIZE)
		newDomain := (*domain)(unsafe.Pointer(newDomainMem))
		newThreadMem := allocPage()
		Memclr(newThreadMem, PAGE_SIZE)
		newThread := (*thread)(unsafe.Pointer(newThreadMem))
		err = StartProgram(progs[i], newDomain, newThread)
		if err != 0 {
			kernelPanic("Could not start program")
		}
		newDomain.MemorySpace.mapPage(newThreadMem, newThreadMem, PAGE_RW|PAGE_PERM_KERNEL)
		AddDomain(newDomain)
	}

	kprintln("domain size: ", uint(unsafe.Sizeof(domains[0])),
		" thread_size: ", uint(unsafe.Sizeof(threads[0])),
		" total: ", uint(unsafe.Sizeof(domains[0])+unsafe.Sizeof(threads[0])))
	kprintln("stack start: ", uintptr(scheduleThread.kernelStack.hi),
		" stack end: ", uintptr(scheduleThread.kernelStack.lo))
	kprintln("info: ", uint(unsafe.Sizeof(currentThread.info)),
		" regs: ", uint(unsafe.Sizeof(currentThread.regs)))
	if currentThread == nil {
		kernelPanic("I expect AddDomain to set currentThread variable")
	}
	kernelThreadInit()
	kernelPanic("Could not jump to user space :/")

}

func gpfPanic() {
	kerrorln("\nReceived General Protection fault. Disabling Interrupts and halting")
	kprintln("Errorcode: ", uintptr(currentThread.info.ExceptionCode))
	panicHelper(currentThread)
}

func printFuncName(pc uintptr) {
	f := findfuncTest(pc)
	if !f.valid() {
		kprintln("func: ", pc)
		return
	}
	s := f._Func().Name()
	file, line := f._Func().FileLine(pc)
	_, filename := path.Split(file)
	kprintln(s, " (", filename, ":", line, ")")
}

func panicHelper(thread *thread) {
	kprintln("Domain ID: ", thread.domain.pid, ", Thread ID: ", thread.tid)
	if thread.isKernelInterrupt {
		kprint("In kernel function: ")
		printFuncName(thread.kernelInfo.EIP)
	} else {
		kprintln("In user function: ", thread.info.EIP)
	}
	printThreadRegisters(thread, defaultScreenWriter)
	printThreadRegisters(thread, defaultLogWriter)
	DisableInterrupts()
	Hlt()
}

// wrapper for do_kernelPanic that gets the return address
// and pushers it on the stack and then calls do_kernelPanic
func kernelPanic(msg string)

//go:nosplit
func do_kernelPanic(caller uintptr, msg string) {
	kerrorln("\n", msg, " - kernel panic :(")
	kprint("Called from function: ")
	printFuncName(caller - 4) // account for the fact that caller points to the instruction after the call
	if currentThread != nil {
		panicHelper(currentThread)
	} else {
		kerrorln("Cannot print registers. 'currentThread' is nil")
	}
	DisableInterrupts()
	Hlt()
	// does not return
}

func printThreadRegisters(t *thread, w io.Writer) {
	kFprint(w, "User regs:          Kernel regs:\n")
	f := findfuncTest(uintptr(t.kernelInfo.EIP))
	kFprint(w, "EIP: ", t.info.EIP, "      ", "EIP: ", t.kernelInfo.EIP, " ", f._Func().Name(), "\n")
	//rintRegisterLineInfo("EIP: ", t.info.EIP, t.kernelInfo.EIP, f._Func().Name())
	printRegisterLine(w, 20, "ESP: ", t.info.ESP, t.kernelInfo.ESP)
	printRegisterLine(w, 20, "EBP: ", t.regs.EBP, t.kernelRegs.EBP)
	printRegisterLine(w, 20, "EAX: ", t.regs.EAX, t.kernelRegs.EAX)
	printRegisterLine(w, 20, "EBX: ", t.regs.EBX, t.kernelRegs.EBX)
	printRegisterLine(w, 20, "ECX: ", t.regs.ECX, t.kernelRegs.ECX)
	printRegisterLine(w, 20, "EDX: ", t.regs.EDX, t.kernelRegs.EDX)
	printRegisterLine(w, 20, "ESI: ", t.regs.ESI, t.kernelRegs.ESI)
	printRegisterLine(w, 20, "EDI: ", t.regs.EDI, t.kernelRegs.EDI)
	printRegisterLine(w, 20, "EFLAGS: ", t.info.EFLAGS, t.kernelInfo.EFLAGS)
	printRegisterLine(w, 20, "Exception: ", t.info.ExceptionCode, t.kernelInfo.ExceptionCode)
	printRegisterLine(w, 20, "Interrupt: ", t.info.InterruptNumber, t.kernelInfo.InterruptNumber)
	printRegisterLine(w, 20, "Krn ESP: ", t.regs.KernelESP, t.kernelRegs.KernelESP)
}

func printRegisterLine(w io.Writer, tablength int, label string, userReg, kernelReg uint32) {
	firstLength := len(label)
	kFprint(w, label, uintptr(userReg))
	// pad number
	firstLength += 3 // account for the hexadecimal 0x#
	for i, n := firstLength, userReg>>4; i < 20; i, n = i+1, n>>4 {
		if n == 0 {
			kFprint(w, " ")
		}
	}
	kFprint(w, label, uintptr(kernelReg), "\n")
}

func printRegisters(w io.Writer, info *InterruptInfo, regs *RegisterState) {
	kFprint(w, "Interrupt: ", uintptr(info.InterruptNumber), "\n")
	kFprint(w, "Exception: ", uintptr(info.ExceptionCode), "\n")
	kFprint(w, "EIP: ", uintptr(info.EIP), "\n")
	kFprint(w, "CS: ", uintptr(info.CS), "\n")
	kFprint(w, "EFLAGS: ", uintptr(info.EFLAGS), "\n")
	kFprint(w, "ESP: ", uintptr(info.ESP), "\n")
	kFprint(w, "SS: ", uintptr(info.SS), "\n")
	kFprint(w, "-----------\n")
	kFprint(w, "GS: ", uintptr(regs.GS), "\n")
	kFprint(w, "FS: ", uintptr(regs.FS), "\n")
	kFprint(w, "ES: ", uintptr(regs.ES), "\n")
	kFprint(w, "DS: ", uintptr(regs.DS), "\n")
	kFprint(w, "EBP: ", uintptr(regs.EBP), "\n")
	kFprint(w, "EAX: ", uintptr(regs.EAX), "\n")
	kFprint(w, "EBX: ", uintptr(regs.EBX), "\n")
	kFprint(w, "ECX: ", uintptr(regs.ECX), "\n")
	kFprint(w, "EDX: ", uintptr(regs.EDX), "\n")
	kFprint(w, "ESI: ", uintptr(regs.ESI), "\n")
	kFprint(w, "EDI: ", uintptr(regs.EDI), "\n")
	kFprint(w, "KernelESP", uintptr(regs.KernelESP), "\n")
}

func printTid(w io.Writer, t *thread) {
	kFprint(w, "Pid: ", t.domain.pid, ", Tid: ", t.tid, "\n")
}

func debug_print_flags(flags uint8) {
	res := flags
	for i := 0; i < 8; i++ {
		if res&uint8(1) == 1 {
			text_mode_print_char(0x30 + uint8(i))
		}
		res = res >> 1
	}

	text_mode_print_char(0x0a)

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
	var s string
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&s)) // case 1
	hdr.Data = uintptr(unsafe.Pointer(ptr))            // case 6 (this case)
	hdr.Len = int(n)
	return s
}
