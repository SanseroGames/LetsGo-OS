package main

import (
    "reflect"
    "unsafe"
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

func text_mode_init(){
    fb = *(*[]uint16)(unsafe.Pointer(&reflect.SliceHeader{
        Len:  fbWidth * fbHeight,
	    Cap:  fbWidth * fbHeight,
	    Data: fbPhysAddr,
    }))
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

func text_mode_print_col(s string, attr uint8){
    for _, b := range s {
        text_mode_print_char_col(uint8(b), attr)
    }
}

func text_mode_println_col(s string, attr uint8){
    for _, b := range s {
        text_mode_print_char_col(uint8(b), attr)
    }
    text_mode_print_char(0xa)

}


func text_mode_print(s string) {
    for _, b := range s {
        text_mode_print_char(uint8(b))
    }
}

func text_mode_println(s string) {
    for _, b := range s {
        text_mode_print_char(uint8(b))
    }
    text_mode_print_char(0xa)
}

func text_mode_print_error(s string) {
    for _, b := range s {
        text_mode_print_char_col(uint8(b), 4<<4 | 0xf)
    }
}

func text_mode_print_errorln(s string){
    text_mode_print_error(s)
    text_mode_print_char(0xa)
}

func text_mode_print_char(char uint8){
    text_mode_print_char_col(char, 0<<4 | 0xf)
}

func text_mode_print_char_col(char uint8, attr uint8){
    text_mode_check_fb_move()
    if(char == 0x0a){
        fbCurLine++
        fbCurCol=0
        text_mode_check_fb_move()
    } else {
        if(fbCurCol>=fbWidth){ return }
        fb[fbCurCol+fbCurLine*fbWidth] = uint16(attr)<<8 | uint16(char)
        fbCurCol++
    }
}

func text_mode_print_hex64(num uint64){
    text_mode_print_hex32(uint32(num >> 32))
    text_mode_print_hex32(uint32(num))
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
