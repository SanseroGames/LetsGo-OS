package kernel

import (
	"unsafe"
)

type TextModeWriter struct {
}

func (e TextModeWriter) Write(p []byte) (int, error) {
	textModePrintBytes(p)
	return len(p), nil
}

type TextModeErrorWriter struct {
}

func (e TextModeErrorWriter) Write(p []byte) (int, error) {
	textModePrintErrorBytes(p)
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

func TextModeInit() {
	fb = unsafe.Slice((*uint16)(unsafe.Pointer(fbPhysAddr)), fbWidth*fbHeight)
	//text_mode_enable_cursor()
	//text_mode_update_cursor(0,0)
}

func TextModeFlushScreen() {
	for i := range fb {
		fb[i] = 0xf00
	}
}

func textModeCheckFbMove() {
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

func textModePrintLnCol(s string, attr uint8) {
	for _, b := range s {
		textModePrintCharCol(uint8(b), attr)
	}
	textModePrintChar(0xa)
}

func textModePrintBytes(a []byte) {
	for _, b := range a {
		textModePrintChar(uint8(b))
	}
}

func textModePrintErrorBytes(a []byte) {
	for _, b := range a {
		textModePrintCharCol(uint8(b), 4<<4|0xf)
	}
}

func textModePrintChar(char uint8) {
	textModePrintCharCol(char, 0<<4|0xf)
}

func textModePrintCharCol(char uint8, attr uint8) {
	textModeCheckFbMove()
	if char == '\n' {
		fbCurLine++
		fbCurCol = 0
		textModeCheckFbMove()
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
