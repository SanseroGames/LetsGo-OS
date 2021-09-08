package main

import (
	"unsafe"
    "reflect"
)

var initDomain domain
var initThread thread

var testDomain domain
var testThread thread

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

    //elfHdr, loadAddr := LoadElfFile("/usr/hellogo", &user.MemorySpace)
    //if elfHdr == nil {
    //    kernelPanic("Could not load elf file")
    //}
    //entry := elfHdr.Entry
    ////get a page for the initial stack
    //stackpages := 16
    //userStackBaseAddr := uintptr(0xffffc000)
    //var stackPages [16]uintptr
    //for i:=0; i < stackpages; i++ {
    //    stack := allocPage()
    //    Memclr(stack, PAGE_SIZE)
    //    user.MemorySpace.mapPage(stack, userStackBaseAddr-uintptr((i+1)*PAGE_SIZE), PAGE_RW | PAGE_PERM_USER)
    //    stackPages[i] = stack

    //}
    //var aux [32]auxVecEntry
    //nrVec := LoadAuxVector(aux[:], elfHdr, loadAddr)
    //nrVec += nrVec % 2
    //vecByteSize := nrVec*int(unsafe.Sizeof(aux[0]))

    //stack := (*[1 << 15]uint32)(unsafe.Pointer(stackPages[0]))[:PAGE_SIZE/4]
    //for i,n := range aux[:nrVec] {
    //    index := PAGE_SIZE/4-1-vecByteSize/4+i*2
    //    stack[index] = n.Type
    //    stack[index+1] = n.Value
    //}

    //initDomain.Segments = user.Segments
    //initDomain.MemorySpace = user.MemorySpace

    //initThread.StackStart = userStackBaseAddr
    //initDomain.EnqueueThread(&initThread)
    //EnqueueDomain(&initDomain)

    //currentTask = user.MemorySpace
    //switchPageDir(user.MemorySpace.PageDirectory)
    //JumpUserMode(user.Segments, uintptr(entry), userStackBaseAddr - 4 - uintptr(vecByteSize) -4-4-4)
    err := StartProgram("/usr/test", &initDomain, &initThread)
    if err != 0 {
        kernelPanic("Could not start /usr/hellogo")
    }
    err = StartProgram("/usr/test", &testDomain, &testThread)
    if err != 0 {
        kernelPanic("Could not start /usr/test")
    }
    EnqueueDomain(&initDomain)
    EnqueueDomain(&testDomain)
    startDomain := initDomain
    switchPageDir(startDomain.MemorySpace.PageDirectory)
    JumpUserMode(startDomain.Segments, uintptr(startDomain.CurThread.info.EIP), uintptr(startDomain.CurThread.info.ESP))
    kernelPanic("Could not jump to user space :/")
}

func testFunc() {
    text_mode_println("Hi, from user mode")
    //print("asdf")
}


func gpfPanic(info *InterruptInfo, regs *RegisterState){
    text_mode_print_char(0xa)
    text_mode_print_errorln("Received General Protection fault. Disabling Interrupts and halting")
    text_mode_print("Thread ID: ")
    text_mode_print_hex(uint8(currentDomain.CurThread.tid))
    text_mode_println("")
    text_mode_print("Errorcode: ")
    text_mode_print_hex32(info.ExceptionCode)
    text_mode_println("")
    panicHelper(info, regs)
}

func panicHelper(info *InterruptInfo, regs *RegisterState){
    printRegisters(info, regs)
    DisableInterrupts()
    Hlt()
}

func kernelPanic(msg string) {
    text_mode_print_errorln(msg)
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

func printTid() {
    text_mode_print("tid: ")
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
