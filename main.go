package main

import (
	"unsafe"
    "reflect"
)

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

    //text_mode_print_hex32(uint32(stackstart))
    //text_mode_print_char(0x0a)

    //for i, n := range (*info).tmp {
    //    if (i > 64) {break}
    //    text_mode_print_hex32(n)
    //    if i % 8 == 7 {
    //        text_mode_print_char(0x0a)
    //    } else if i % 2 == 1 {
    //        text_mode_print("  ")
    //    } else {
    //        text_mode_print(" ")
    //    }
    //}
    //text_mode_print_hex32(uint32(len((*info).tmp)))
    //text_mode_println("")

    //text_mode_print_hex32(uint32(len((*info).elems)))
    //text_mode_println("")

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

    entry := LoadElfFile("/usr/shell.o", &user.MemorySpace)
    //get a page for the initial stack
    stackpages := 16
    userStackAddr := uintptr(0x1fff0000)
    for i:=0; i < stackpages; i++ {
        stack := allocPage()
        Memclr(stack, PAGE_SIZE)
        user.MemorySpace.mapPage(stack, userStackAddr-uintptr(i*PAGE_SIZE), PAGE_RW | PAGE_PERM_USER)

    }
    currentTask = user.MemorySpace
    switchPageDir(user.MemorySpace.PageDirectory)
    JumpUserMode(user.Segments, uintptr(entry), userStackAddr+PAGE_SIZE-16)

    for {
        Hlt()
        //UpdateShell()
    }
    //TODO: Initialize go runtime. Maybe not?
}

func testFunc() {
    text_mode_println("Hi, from user mode")
    //print("asdf")
}


func gpfPanic(info *InterruptInfo, regs *RegisterState){
    text_mode_print_char(0xa)
    text_mode_print_errorln("Received General Protection fault. Disabling Interrupts and halting")
    panicHelper(info, regs)
}

func panicHelper(info *InterruptInfo, regs *RegisterState){
    printRegisters(info, regs)
    DisableInterrupts()
    Hlt()
}

func kernelPanic() {
    text_mode_print_errorln("kernel panic :(")
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
