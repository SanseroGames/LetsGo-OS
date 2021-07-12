package main

import (
	"unsafe"
    "reflect"
)

const (
	fbWidth            = 80
	fbHeight           = 25
	fbPhysAddr uintptr = 0xb8000
)

var (
    fbCurLine = 0
    fbCurCol = 0
)

var fb []uint16

func main() {
    fb = *(*[]uint16)(unsafe.Pointer(&reflect.SliceHeader{
        Len:  fbWidth * fbHeight,
	    Cap:  fbWidth * fbHeight,
	    Data: fbPhysAddr,
    }))

    text_mode_flush_screen()
    s :=  "Hi and welcome to Let's-Go OS"
    s2 := "An Operating System written in the GO programming language"
    s3 := "It can't do much, but I hope you enjoy your stay"

    text_mode_println(s)
    text_mode_println(s2)
    text_mode_println(s3)
    text_mode_print_char(0x0a)

    InitInterrupts()
    SetInterruptHandler(0x80, linuxSyscallHandler)
    InitPIC()
    EnableInterrupts()
    InitKeyboard()

    for {
        Hlt()
        temp()
    }
    //TODO: Initialize go runtime
}

func temp(){
    text_mode_print_hex(uint8(buffer.Len()))
    text_mode_print(" ")
    text_mode_print_hex(uint8(buffer.Pop().test))
    text_mode_print_char(0x0a)
}

func linuxSyscallHandler(info *InterruptInfo, regs *RegisterState){
    switch (regs.EAX) {
        case 4: {
            // Linux write syscall
            linuxWriteSyscall(info, regs)
        }
    default: panicHandler(info, regs)
    }
}

func linuxWriteSyscall(info *InterruptInfo, regs *RegisterState){
    // Not safe
    var s string
    hdr := (*reflect.StringHeader)(unsafe.Pointer(&s)) // case 1
    hdr.Data = uintptr(unsafe.Pointer(uintptr(regs.ECX))) // case 6 (this case)
    hdr.Len = int(regs.EDX)
    text_mode_print(s)
}

func panicHandler(info *InterruptInfo, regs *RegisterState){
    text_mode_print_char(0xa)
    text_mode_print_error("Unsupported Linux Syscall received! Disabling Interrupts and halting")
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

    DisableInterrupts()
    Hlt()
}

func text_mode_flush_screen(){
    for i := range fb{
        fb[i] = 0
    }
}

func text_mode_check_fb_move(){
    if(fbCurLine == fbHeight){
        for i := 1; i<fbHeight; i++{
            copy(fb[(i-1)*fbWidth:(i-1)*fbWidth+fbWidth], fb[i*fbWidth:i*fbWidth+fbWidth])
        }
        for i:=0; i < fbWidth; i++{
            fb[(fbHeight-1)*fbWidth + i] = 0
        }
        fbCurLine--
    }
}

func text_mode_print(s string) {
    for _, b := range s {
        text_mode_print_char(uint8(b))
    }
}

func text_mode_println(s string) {
    text_mode_check_fb_move()
    attr := uint16(0<<4 | 0xf)
    for i, b := range s {
        t := i + fbCurCol
        if(t >= fbWidth) {break}
        fb[t+fbCurLine*fbWidth] = attr<<8 | uint16(b)
    }
    fbCurLine++
    fbCurCol = 0
}

func text_mode_print_error(s string) {
    text_mode_check_fb_move()
    attr := uint16(4<<4 | 0xf)
    for i, b := range s {
        t := i + fbCurCol
        if(t >= fbWidth) {break}
        fb[t+fbCurLine*fbWidth] = attr<<8 | uint16(b)
    }
    fbCurLine++
    fbCurCol = 0
}

func text_mode_print_char(char uint8){
    if(fbCurCol>=fbWidth){ return }
    text_mode_check_fb_move()
    attr := uint16(0<<4 | 0xf)
    if(char == 0x0a){
        fbCurLine++
        fbCurCol=0
    } else {
        fb[fbCurCol+fbCurLine*fbWidth] = attr<<8 | uint16(char)
        fbCurCol++
    }
}

func text_mode_print_hex32(num uint32){
    text_mode_print_hex16(uint16(num >> 16))
    text_mode_print_hex16(uint16(num))
}

func text_mode_print_hex16(num uint16){
    text_mode_print_hex(uint8(num >> 8))
    text_mode_print_hex(uint8(num))
}


func text_mode_print_hex(num uint8){
    text_mode_print_hex_char(num >> 4)
    text_mode_print_hex_char(num)
}

func text_mode_print_hex_char(nibble uint8){
    n := nibble & 0xf
    if(n<10){
        text_mode_print_char(0x30+n)
    } else {
        text_mode_print_char(0x41+n-10)
    }
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

