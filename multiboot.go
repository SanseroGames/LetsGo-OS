package main

import (
	"unsafe"
)

const (
	MEM_MAP_AVAILABLE = 1
)

var (
	multibootInfo *MultibootInfo

	loadedModules [30]MultibootModule

	memoryMaps [6]MemoryMap
)

type MultibootInfo struct {
	TotalSize uint32
	reserved  uint32
}

type multibootType uint32

type MultibootTag struct {
	Type uint32
	Size uint32
}

type MultibootMemoryMap struct {
	MultibootTag // Type is 6
	EntrySize    uint32
	EntryVersion uint32
	Entries      MemoryMap // Take pointer of it and use it as slice with
}

// A module represents a module to be loaded along with the kernel.
type MultibootModule struct {
	MultibootTag
	// Start is the inclusive start of the Module memory location
	Start uint32

	// End is the exclusive end of the Module memory location.
	End uint32

	// Cmdline is a pointer to a null-terminated ASCII string.
	cmdline [100]byte
}

func (m *MultibootModule) Cmdline() string {
	if m.Size <= 17 {
		return ""
	}
	return unsafe.String(&m.cmdline[0], max(m.Size-16-1, 1))
}

type Module struct {
	Start   uint32
	End     uint32
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

	mbI := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(info), unsafe.Sizeof(*info))), info.TotalSize-uint32(unsafe.Sizeof(*info)))

	loadedModuleSlice := loadedModules[:]

	foundModules := 0
	for i := uint32(0); i < info.TotalSize; {
		mbTag := (*MultibootTag)(unsafe.Pointer(&mbI[i]))
		kdebugln("Type ", mbTag.Type, " size ", mbTag.Size, " i ", i, " next ", (i+mbTag.Size+7)&0xfffffff8)
		if mbTag.Type == 0 && mbTag.Size == 8 {
			break
		}
		if mbTag.Type == 3 {
			if foundModules < len(loadedModuleSlice) {
				mbMod := (*MultibootModule)(unsafe.Pointer(mbTag))

				loadedModuleSlice[foundModules] = *mbMod
				kdebugln(mbMod.Cmdline())
				kdebugln(loadedModuleSlice[foundModules].Cmdline())
				kdebugln(uintptr(loadedModuleSlice[foundModules].Start), " ", uintptr(loadedModuleSlice[foundModules].End), " ", loadedModuleSlice[foundModules].Size, " ", loadedModuleSlice[foundModules].cmdline[0])
				foundModules++
			} else {
				kerrorln("[WARNING] Not enough space to load all modules")
			}
		}
		if mbTag.Type == 6 {
			memTag := (*MultibootMemoryMap)(unsafe.Pointer(mbTag))
			nrentries := (memTag.Size - 16) / memTag.EntrySize
			maps := unsafe.Slice(&(memTag.Entries), nrentries)
			for i, v := range maps {
				if i > len(memoryMaps) {
					kerrorln("[WARNING] More memory maps than space in memorymap list")
					break
				}
				// kdebugln(uintptr(v.BaseAddr), " ", uintptr(v.Length), " ", v.Type)
				memoryMaps[i] = v
			}
		}
		oldi := i
		size := max(mbTag.Size, 8)
		i = (i + size + 7) & 0xfffffff8
		if oldi == i {
			kerrorln("[WARNING] Loading multiboot modules behaved weird")
			break
		}
	}
	kdebugln("Done")
	//printMemMaps()
}
