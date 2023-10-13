package main

import (
	"reflect"
	"unsafe"
)

const (
	MEM_MAP_AVAILABLE = 1
)

var (
	multibootInfo *MultibootInfo

	loadedModules [20]MultibootModule

	memoryMaps [6]MemoryMap
)

type MultibootInfo struct {
	total_size uint32
	reserved   uint32
}

type multibootType uint32

type MultibootTag struct {
	typ  multibootType
	size uint32
}

// A module represents a module to be loaded along with the kernel.
type MultibootModule struct {
	// Start is the inclusive start of the Module memory location
	Start uint32

	// End is the exclusive end of the Module memory location.
	End uint32

	// Cmdline is a pointer to a null-terminated ASCII string.
	Cmdline string
}

type MemoryMap struct {
	BaseAddr uint64

	Length uint64

	Type uint32

	reserved uint32
}

func InitMultiboot(info *MultibootInfo) {
	multibootInfo = info

	mbI := *(*[]uint32)(unsafe.Pointer(&reflect.SliceHeader{
		Len:  int(info.total_size),
		Cap:  int(info.total_size),
		Data: uintptr(unsafe.Pointer(info)) + 8,
	}))

	foundModules := 0
	for i := uint32(0); i < (*info).total_size; {
		if mbI[i] == 0 && mbI[i+1] == 8 {
			break
		}
		if mbI[i] == 3 && foundModules < len(loadedModules) {
			loadedModules[foundModules].Start = mbI[i+2]
			loadedModules[foundModules].End = mbI[i+3]
			hdr := (*reflect.StringHeader)(unsafe.Pointer(&loadedModules[foundModules].Cmdline))
			hdr.Data = uintptr(unsafe.Pointer(info)) + 8 + uintptr(i)*4 + 16
			hdr.Len = int(mbI[i+1]-16) - 1 //Possible underflow
			foundModules++
		}
		if mbI[i] == 6 {
			size := mbI[i+1]
			esize := mbI[i+2]
			nrentries := (size - 16) / esize
			maps := *(*[]MemoryMap)(unsafe.Pointer(&reflect.SliceHeader{
				Len:  int(nrentries),
				Cap:  int(nrentries),
				Data: uintptr(unsafe.Pointer(info)) + 8 + uintptr(i)*4 + 16,
			}))
			copy(memoryMaps[:], maps)
		}
		oldi := i
		size := mbI[i+1]
		i += size / 4
		if size%4 != 0 {
			i++
		}
		if i%2 == 1 {
			i++
		}
		if oldi == i {
			break
		}
	}
	//printMemMaps()
}

func printMemMaps() {
	for _, n := range memoryMaps {
		text_mode_print_hex32(uint32(n.BaseAddr))
		text_mode_print(" ")
		text_mode_print_hex32(uint32(n.Length))
		text_mode_print(" ")
		text_mode_print_hex32(n.Type)
		text_mode_println("")
	}
}
