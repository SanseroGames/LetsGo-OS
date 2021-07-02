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

    // Display a string to the top-left corner of the screen one character
    // at a time.
    //for i := 0; i < 684000; i++ {
	//	for j := 0; j < 1500; j++ {
	//	}
	//}
    text_mode_flush_screen()
    s := "Hello World. This is the kernel speaking!"

    for i := 0; i < fbHeight-2; i++{
       text_mode_print(s)
    }
    text_mode_print("overlfow!!")
    text_mode_print("adsfl")
    text_mode_print("adsflijdsakfdfkjdsafakdsflkasdflkdsaflkdsalfkdsalkfdsalkfdsalksalkf 00000 difiiidfslllsf 1111  333 3 q")
    text_mode_print_hex(0x5f)
    text_mode_print_char(0x0a)
    InitInterrupts()
    SetInterruptHandler(0x80, panicHandler)
    InitPIC()
    EnableInterrupts()
    InitKeyboard()

    for {
        Hlt()
    }
    //TODO: Initialize go runtime
}

func panicHandler(nop uint8){
    text_mode_print_error("Linux Syscall received. GO probably paniced. Disabling Interrupts and halting")
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

