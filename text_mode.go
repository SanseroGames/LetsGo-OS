package main

import (
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
	fbWidth              = 80
	fbHeight             = 25
	fbPhysAddr   uintptr = 0xb8000
	cursorHeight         = 1  // scanlines
	cursorStart          = 11 // scanlines
)

var (
	fbCurLine = 0
	fbCurCol  = 0
)

var fb []uint16

func text_mode_init() {
	fb = unsafe.Slice((*uint16)(unsafe.Pointer(fbPhysAddr)), fbWidth*fbHeight)
	//text_mode_enable_cursor()
	//text_mode_update_cursor(0,0)
}

func text_mode_flush_screen() {
	for i := range fb {
		fb[i] = 0xf00
	}
}

func text_mode_check_fb_move() {
	for fbCurLine >= fbHeight {
		copy(fb[:len(fb)-fbWidth], fb[fbWidth:])
		//for i := 1; i<fbHeight; i++{
		//    copy(fb[(i-1)*fbWidth:(i-1)*fbWidth+fbWidth], fb[i*fbWidth:i*fbWidth+fbWidth])
		//}
		for i := 0; i < fbWidth; i++ {
			fb[(fbHeight-1)*fbWidth+i] = 0xf00
		}
		fbCurLine--
	}
}

func text_mode_println_col(s string, attr uint8) {
	for _, b := range s {
		text_mode_print_char_col(uint8(b), attr)
	}
	text_mode_print_char(0xa)
}

func text_mode_print_bytes(a []byte) {
	for _, b := range a {
		text_mode_print_char(uint8(b))
	}
}

func text_mode_print_error_bytes(a []byte) {
	for _, b := range a {
		text_mode_print_char_col(uint8(b), 4<<4|0xf)
	}
}

func text_mode_print_char(char uint8) {
	text_mode_print_char_col(char, 0<<4|0xf)
}

func text_mode_print_char_col(char uint8, attr uint8) {
	text_mode_check_fb_move()
	if char == '\n' {
		fbCurLine++
		fbCurCol = 0
		text_mode_check_fb_move()
	} else if char == '\b' {
		fb[fbCurCol+fbCurLine*fbWidth] = 0xf00
		fbCurCol = fbCurCol - 1
		if fbCurCol < 0 {
			fbCurCol = 0
		}
	} else if char == '\r' {
		fbCurCol = 0
	} else {
		if fbCurCol >= fbWidth {
			return
		}
		fb[fbCurCol+fbCurLine*fbWidth] = uint16(attr)<<8 | uint16(char)
		fbCurCol++
	}
	//text_mode_update_cursor(fbCurCol, fbCurLine)
}
