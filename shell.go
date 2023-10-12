package main

import (
	"reflect"
	"unsafe"
)

var (
	lineData  [fbWidth]byte
	line      string
	lineIndex = 0
	lineHdr   reflect.StringHeader
)

func InitShell() {
	lineHdr := (*reflect.StringHeader)(unsafe.Pointer(&line)) // case 1
	lineHdr.Data = uintptr(unsafe.Pointer(&lineData))         // case 6 (this case)
	lineHdr.Len = 0
}

func UpdateShell() {
	for {
		if buffer.Len() == 0 {
			break
		}
		raw_key := buffer.Pop().Keycode
		pressed := raw_key&0x80 == 0
		key := raw_key & 0x7f
		if pressed {
			if key == 0x1C {
				// Enter
				text_mode_print_char(0x0a)
				lineIndex = 0
				lineHdr.Len = 0
			} else if key == 0x0E {
				// Backspace
			} else {
				c := translateKeycode(key)
				if c != 0 {
					lineData[lineIndex] = c
					lineIndex++
					lineHdr.Len++
				}
			}
		}
		text_mode_println(line)
		text_mode_print_hex(uint8(lineHdr.Len))
		text_mode_print_hex(uint8(lineData[0]))
		//text_mode_print(" ")
		//text_mode_print_hex(key)
		//text_mode_print("  ")
	}
}
