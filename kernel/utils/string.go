package utils

import (
	// "iter"
	// "syscall"
	"unsafe"
)

func CString(ptr uintptr) string {
	n := 0
	p := unsafe.Pointer(ptr)
	for ; *(*byte)(unsafe.Add(p, n)) != 0; n++ {
	}
	return unsafe.String((*byte)(p), n)
}

// func CStringLen(seq iter.Seq2[byte, syscall.Errno]) (int, syscall.Errno) {
// 	length := 0
// 	for v, err := range seq {
// 		if err != 0 {
// 			return length, err
// 		}
// 		length++
// 		if v == 0 {
// 			break
// 		}
// 	}
// 	return length, 0
// }
