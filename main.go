package main

import (
	"unsafe"
    "reflect"
    "path"
)

var progs = [...]string {
    "/usr/hellogo",
}

var domains [len(progs)]domain
var threads [len(progs)]thread

func main()

//go:linkname kmain main.main
func kmain(info *MultibootInfo, stackstart uintptr, stackend uintptr) {
    if stackstart <= stackend {
        kernelPanic("No stack")
    }
    text_mode_init()

    text_mode_flush_screen()
    s :=  "Hi and welcome to Let's-Go OS"
    s2 := "An Operating System written in the GO programming language"
    s3 := "It can't do much, but I hope you enjoy your stay"

    kprintln(s)
    kprintln(s2)
    kprintln(s3)
    kprintln("")

    InitSegments()

    InitInterrupts()

    SetInterruptHandler(0xd, gpfPanic, KCS_SELECTOR, PRIV_USER)

    InitSyscall()

    InitPIC()

    InitPit()
    InitKeyboard()
    InitATA()

    InitMultiboot(info)
    //printMemMaps()

    InitPaging()

    InitUserMode(stackstart, stackend)

    text_mode_println_col("Initilaization complete", 0x2)
    //HdReadSector()

    var err int
    for i :=0; i < len(progs); i++ {
        err = StartProgram(progs[i], &domains[i], &threads[i])
        if err != 0 {
            kernelPanic("Could not start program")
        }
    }

    for i := range domains {
        AddDomain(&domains[i])
    }
    kprintln("domain size: ", uint(unsafe.Sizeof(domains[0])),
                " thread_size: ", uint(unsafe.Sizeof(threads[0])),
                " total: ", uint(unsafe.Sizeof(domains[0]) + unsafe.Sizeof(threads[0])))
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

func gpfPanic(){
    text_mode_print_char(0xa)
    text_mode_print_errorln("Received General Protection fault. Disabling Interrupts and halting")
    text_mode_print("Errorcode: ")
    text_mode_print_hex32(currentThread.info.ExceptionCode)
    text_mode_println("")
    panicHelper(currentThread)
}

func printFuncName(pc uintptr) {
    f := findfuncTest(pc)
    if !f.valid() {
        text_mode_print("func: ")
        text_mode_print_hex32(uint32(pc))
        text_mode_println("\n")
        return
    }
    s := f._Func().Name()
    text_mode_print(s)
    text_mode_print(" (")
    file, line := f._Func().FileLine(pc)
    _, filename := path.Split(file)
    text_mode_print(filename)
    text_mode_print(":")
    text_mode_print_hex32(uint32(line))
    text_mode_println(")")
}

func panicHelper(thread *thread){
    text_mode_print("Domain ID: ")
    text_mode_print_hex(uint8(thread.domain.pid))
    text_mode_println("")
    text_mode_print("Thread ID: ")
    text_mode_print_hex(uint8(thread.tid))
    text_mode_println("")
    if kernelInterrupt {
        text_mode_print("In kernel function: ")
        printFuncName(uintptr(thread.kernelInfo.EIP))
    } else {
        text_mode_print("In user function: ")
        text_mode_print_hex32(thread.info.EIP)
    }
    text_mode_println("")
    printThreadRegisters(thread)
    DisableInterrupts()
    Hlt()
}

// wrapper for do_kernelPanic that gets the return address
// and pushers it on the stack and then jumps to do_kernelPanic
// this messes up the stack but we don't return so it's no issue
func kernelPanic(msg string)
//go:nosplit
func do_kernelPanic(caller uintptr, msg string) {
    text_mode_print_errorln(msg)
    text_mode_print_errorln("kernel panic :(\n")
    text_mode_print("Called from function: ")
    printFuncName(caller-4) // account fo the fact that caller points to the instruction after the call
    if currentThread != nil {
        panicHelper(currentThread)
    }
    DisableInterrupts()
    Hlt()
    // does not return
}

func printThreadRegisters(t *thread) {
    text_mode_println("User regs:       Kernel regs:")
    f := findfuncTest(uintptr(t.kernelInfo.EIP))
    printRegisterLineInfo("EIP: ", t.info.EIP, t.kernelInfo.EIP, f._Func().Name())
    printRegisterLine("ESP: ", t.info.ESP, t.kernelInfo.ESP)
    printRegisterLine("EBP: ", t.regs.EBP, t.kernelRegs.EBP)
    printRegisterLine("EAX: ", t.regs.EAX, t.kernelRegs.EAX)
    printRegisterLine("EBX: ", t.regs.EBX, t.kernelRegs.EBX)
    printRegisterLine("ECX: ", t.regs.ECX, t.kernelRegs.ECX)
    printRegisterLine("EDX: ", t.regs.EDX, t.kernelRegs.EDX)
    printRegisterLine("ESI: ", t.regs.ESI, t.kernelRegs.ESI)
    printRegisterLine("EDI: ", t.regs.EDI, t.kernelRegs.EDI)
}

func printRegisterLine(label string, userReg, kernelReg uint32) {
    printRegisterLineInfo(label, userReg, kernelReg, "")
}


func printRegisterLineInfo(label string, userReg, kernelReg uint32, s string) {
    text_mode_print(label)
    text_mode_print_hex32(userReg)
    text_mode_print("    ")
    text_mode_print(label)
    text_mode_print_hex32(kernelReg)
    if s != "" {
        text_mode_print(" (")
        text_mode_print(s)
        text_mode_print(")")
    }
    text_mode_println("")
}

func printRegisters(info *InterruptInfo, regs *RegisterState){
    text_mode_print("EIP: ")
    text_mode_print_hex32(info.EIP)
    text_mode_print_char(0x0a)
    text_mode_print("EAX: ")
    text_mode_print_hex32(regs.EAX)
    text_mode_print_char(0x0a)
    text_mode_print("EBX: ")
    text_mode_print_hex32(regs.EBX)
    text_mode_print_char(0x0a)
    text_mode_print("ECX: ")
    text_mode_print_hex32(regs.ECX)
    text_mode_print_char(0x0a)
    text_mode_print("EDX: ")
    text_mode_print_hex32(regs.EDX)
    text_mode_print_char(0x0a)
    text_mode_print("ESI: ")
    text_mode_print_hex32(regs.ESI)
    text_mode_print_char(0x0a)
    text_mode_print("EDI: ")
    text_mode_print_hex32(regs.EDI)
    text_mode_print_char(0x0a)
    text_mode_print("EBP: ")
    text_mode_print_hex32(regs.EBP)
}

func printTid() {
    text_mode_print("pid: ")
    text_mode_print_hex(uint8(currentThread.domain.pid))
    text_mode_print(" tid: ")
    text_mode_print_hex(uint8(currentThread.tid))
    text_mode_print(" ")
}

func debug_print_flags(flags uint8){
    res := flags
    for i:=0; i<8; i++ {
        if(res & uint8(1) == 1) {
            text_mode_print_char(0x30+uint8(i))
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
    hdr.Data = uintptr(unsafe.Pointer(ptr)) // case 6 (this case)
    hdr.Len = int(n)
	return s
}
