package main

import (
	"unsafe"
    "reflect"
)

const DOMAINS = 0x1

var domains [DOMAINS]domain
var threads [DOMAINS]thread

func goEntry()

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
    text_mode_print_char(0x0a)

    InitSegments()

    InitInterrupts()

    SetInterruptHandler(0xd, gpfPanic, KCS_SELECTOR, PRIV_USER)

    InitSyscall()

    InitPIC()
    EnableInterrupts()

    InitKeyboard()
    InitATA()

    InitMultiboot(info)
    //printMemMaps()

    InitPaging()

    InitUserMode(stackstart)

    InitShell()

    text_mode_println_col("Initilaization complete", 0x2)
    //HdReadSector()

    for i:=0; i < DOMAINS; i++ {
        var err int
        err = StartProgram("/usr/hellorust", &domains[i], &threads[i])
        if err != 0 {
            kernelPanic("Could not start program")
        }
    }
    for i := range domains {
        EnqueueDomain(&domains[i])
    }

    currentDomain = &domains[0x0]
    switchPageDir(currentDomain.MemorySpace.PageDirectory)
    JumpUserMode(currentDomain.CurThread.regs, currentDomain.CurThread.info)
    kernelPanic("Could not jump to user space :/")
}

func gpfPanic(){
    text_mode_print_char(0xa)
    text_mode_print_errorln("Received General Protection fault. Disabling Interrupts and halting")
    text_mode_print("Domain ID: ")
    text_mode_print_hex(uint8(currentDomain.pid))
    text_mode_println("")
    text_mode_print("Thread ID: ")
    text_mode_print_hex(uint8(currentDomain.CurThread.tid))
    text_mode_println("")
    text_mode_print("Errorcode: ")
    text_mode_print_hex32(currentInfo.ExceptionCode)
    text_mode_println("")
    panicHelper(currentInfo, currentRegs)
}

func panicHelper(info *InterruptInfo, regs *RegisterState){
    printRegisters(info, regs)
    DisableInterrupts()
    Hlt()
}

func kernelPanic(msg string) {
    text_mode_print_errorln(msg)
    text_mode_print_errorln("kernel panic :(")
    printTid()
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
    text_mode_print_hex(uint8(currentDomain.pid))
    text_mode_print(" tid: ")
    text_mode_print_hex(uint8(currentDomain.CurThread.tid))
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
