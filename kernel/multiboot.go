package kernel

import (
	"unsafe"

	"github.com/sanserogames/letsgo-os/kernel/log"
	"github.com/sanserogames/letsgo-os/kernel/multiboot"
)

var (
	multibootInfo *multiboot.MultibootInfo

	loadedModules [30]multiboot.MultibootModule

	MemoryMaps [6]multiboot.MemoryMap
)

func InitMultiboot(info *multiboot.MultibootInfo) {
	multibootInfo = info

	mbI := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(info), unsafe.Sizeof(*info))), info.TotalSize-uint32(unsafe.Sizeof(*info)))

	loadedModuleSlice := loadedModules[:]

	foundModules := 0
	for i := uint32(0); i < info.TotalSize; {
		mbTag := (*multiboot.MultibootTag)(unsafe.Pointer(&mbI[i]))
		log.KDebugLn("Type ", mbTag.Type, " size ", mbTag.Size, " i ", i, " next ", (i+mbTag.Size+7)&0xfffffff8)
		if mbTag.Type == 0 && mbTag.Size == 8 {
			break
		}
		if mbTag.Type == 3 {
			if foundModules < len(loadedModuleSlice) {
				mbMod := (*multiboot.MultibootModule)(unsafe.Pointer(mbTag))

				loadedModuleSlice[foundModules] = *mbMod
				log.KDebugLn(mbMod.Cmdline())
				log.KDebugLn(loadedModuleSlice[foundModules].Cmdline())
				foundModules++
			} else {
				log.KErrorLn("[WARNING] Not enough space to load all modules")
			}
		}
		if mbTag.Type == 6 {
			memTag := (*multiboot.MultibootMemoryMap)(unsafe.Pointer(mbTag))
			nrentries := (memTag.Size - 16) / memTag.EntrySize
			maps := unsafe.Slice(&(memTag.Entries), nrentries)
			for i, v := range maps {
				if i > len(MemoryMaps) {
					log.KErrorLn("[WARNING] More memory maps than space in memorymap list")
					break
				}
				// log.KDebugln(uintptr(v.BaseAddr), " ", uintptr(v.Length), " ", v.Type)
				MemoryMaps[i] = v
			}
		}
		oldi := i
		size := max(mbTag.Size, 8)
		i = (i + size + 7) & 0xfffffff8
		if oldi == i {
			log.KErrorLn("[WARNING] Loading multiboot modules behaved weird")
			break
		}
	}
	log.KDebugLn("Done")
	//printMemMaps()
}
