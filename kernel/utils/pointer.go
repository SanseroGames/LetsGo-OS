package utils

import (
	"unsafe"
)

func UIntToPointer[T any](pointer uintptr)*T {
	return (*T)(unsafe.Pointer(pointer))
}

func UIntToSlice[T any](pointer uintptr, len int )[]T {
	return unsafe.Slice((*T)(unsafe.Pointer(pointer)), len)
}