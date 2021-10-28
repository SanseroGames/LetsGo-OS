package main

import (
	"unsafe"
    "reflect"
)

var progs = [...]string {
    "/usr/hellogo",
    "/usr/helloc",
    "/usr/hellorust",
    "/usr/hellogo",
    "/usr/hellorust",
    "/usr/readtest",
    "/usr/rustread",
    "/usr/rustread",
    "/usr/rustread",
}

var domains [len(progs)]domain
var threads [len(progs)]thread

func main()

//go:linkname kmain main.main
func kmain(info *MultibootInfo, stackstart uintptr) {
    text_mode_init()

    text_mode_flush_screen()
    s :=  "Hi and welcome to Let's-Go OS"
    s2 := "An Operating System written in the GO programming language"
    s3 := "It can't do much, but I hope you enjoy your stay"

    text_mode_println(s)
    text_mode_println(s2)
    text_mode_println(s3)
    text_mode_println("")

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

    InitUserMode(stackstart)

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

    if currentThread == nil {
        kernelPanic("I expect AddDomain to set currentThread variable")
    }
    switchPageDir(currentThread.domain.MemorySpace.PageDirectory)
    JumpUserMode(currentThread.regs, currentThread.info)
    kernelPanic("Could not jump to user space :/")

}

func gpfPanic(){
    text_mode_print_char(0xa)
    text_mode_print_errorln("Received General Protection fault. Disabling Interrupts and halting")
    text_mode_print("Domain ID: ")
    text_mode_print_hex(uint8(currentThread.domain.pid))
    text_mode_println("")
    text_mode_print("Thread ID: ")
    text_mode_print_hex(uint8(currentThread.tid))
    text_mode_println("")
    text_mode_print("Errorcode: ")
    text_mode_print_hex32(currentThread.info.ExceptionCode)
    text_mode_println("")
    panicHelper(&currentThread.info, &currentThread.regs)
}

func panicHelper(info *InterruptInfo, regs *RegisterState){
    printRegisters(info, regs)
    DisableInterrupts()
    Hlt()
}

func kernelPanic(msg string) {
    text_mode_print_errorln(msg)
    text_mode_print_errorln("kernel panic :(")
    if currentThread != nil {
        printTid()
    }
    DisableInterrupts()
    Hlt()
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
