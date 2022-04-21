package main

import (
    "reflect"
    "unsafe"
)

type text_mode_writer struct {
}

func (e text_mode_writer) Write(p []byte) (int, error) {
    text_mode_print_bytes(p)
    return len(p), nil
}

type text_mode_error_writer struct {
}

func (e text_mode_error_writer) Write(p []byte) (int, error) {
    text_mode_print_error_bytes(p)
    return len(p), nil
}

const (
	fbWidth            = 80
	fbHeight           = 25
	fbPhysAddr uintptr = 0xb8000
    cursorHeight       = 1 // scanlines
    cursorStart        = 11 // scanlines
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
    //text_mode_enable_cursor()
    //text_mode_update_cursor(0,0)
}

func text_mode_enable_cursor() {
    Outb(0x3D4, 0x0A);
	Outb(0x3D5, (Inb(0x3D5) & 0xC0) | cursorStart); // cursor start scanline

	Outb(0x3D4, 0x0B);
	Outb(0x3D5, (Inb(0x3D5) & 0xE0) | cursorHeight+cursorStart); // cursor end scanline
}

func text_mode_disable_cursor() {
    Outb(0x3D4, 0x0A);
	Outb(0x3D5, 0x20);
}

func text_mode_update_cursor(x int, y int) {
    pos := y * fbWidth + x;
	Outb(0x3D4, 0x0F);
	Outb(0x3D5, uint8(pos & 0xFF));
	Outb(0x3D4, 0x0E);
	Outb(0x3D5, uint8((pos >> 8) & 0xFF));
}


func text_mode_flush_screen(){
    for i := range fb{
        fb[i] = 0xf00
    }
}

func text_mode_check_fb_move(){
    for fbCurLine >= fbHeight{
        copy(fb[:len(fb)-fbWidth], fb[fbWidth:])
        //for i := 1; i<fbHeight; i++{
        //    copy(fb[(i-1)*fbWidth:(i-1)*fbWidth+fbWidth], fb[i*fbWidth:i*fbWidth+fbWidth])
        //}
        for i:=0; i < fbWidth; i++{
            fb[(fbHeight-1)*fbWidth + i] = 0xf00
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

func text_mode_print_bytes(a []byte) {
    for _, b := range a {
        text_mode_print_char(uint8(b))
    }
}

func text_mode_print_error_bytes(a []byte) {
    for _, b := range a {
        text_mode_print_char_col(uint8(b), 4<<4 | 0xf)
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
    if(char == '\n'){
        fbCurLine++
        fbCurCol=0
        text_mode_check_fb_move()
    } else if char == '\b' {
        fb[fbCurCol+fbCurLine*fbWidth] = 0xf00
        fbCurCol = fbCurCol-1
        if fbCurCol < 0 {
            fbCurCol = 0
        }
    } else if char == '\r' {
        fbCurCol = 0
    } else {
        if(fbCurCol>=fbWidth){ return }
        fb[fbCurCol+fbCurLine*fbWidth] = uint16(attr)<<8 | uint16(char)
        fbCurCol++
    }
    //text_mode_update_cursor(fbCurCol, fbCurLine)
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
