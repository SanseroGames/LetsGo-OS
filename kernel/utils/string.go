package utils

import (
	"iter"
	"syscall"
	"unsafe"
)

func CString(ptr uintptr) string {
	var n int
	for p := ptr; *(*byte)(unsafe.Pointer(p)) != 0; p++ {
		n++
	}
	return unsafe.String((*byte)(unsafe.Pointer(ptr)), n)
}

func CStringLen(seq iter.Seq2[byte, syscall.Errno]) (int, syscall.Errno) {
	length := 0
	for v, err := range seq {
		if err != 0 {
			return length, err
		}
		length++
		if v == 0 {
			break
		}
	}
	return length, 0
}
