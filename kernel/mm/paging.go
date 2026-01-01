package mm

import (
	"path"
	"runtime"
	"unsafe"

	"github.com/sanserogames/letsgo-os/kernel/log"
	"github.com/sanserogames/letsgo-os/kernel/multiboot"
	"github.com/sanserogames/letsgo-os/kernel/panic"
)

const (
	PAGE_FAULT_PRESENT           = 1 << 0
	PAGE_FAULT_WRITE             = 1 << 1
	PAGE_FAULT_USER              = 1 << 2
	PAGE_FAULT_INSTRUCTION_FETCH = 1 << 4
)

var (
	KernelMemSpace MemSpace = MemSpace{}
	maxPages       int      = 0
)

func enablePaging()

func SwitchPageDir(dir *PageTable)

func getCurrentPageDir() *PageTable

func getPageFaultAddr() uint32

type _func struct {
	entry   uintptr // start pc
	nameoff int32   // function name

	args        int32  // in/out args size
	deferreturn uint32 // offset of start of a deferreturn call instruction from entry, if any.

	pcsp      uint32
	pcfile    uint32
	pcln      uint32
	npcdata   uint32
	cuOffset  uint32 // runtime.cutab offset of this function's CU
	funcID    uint64 // set for certain special runtime functions
	flag      uint8
	_         [1]byte // pad
	nfuncdata uint8   // must be last, must end on a uint32-aligned boundary
}

type funcInfo struct {
	*_func
}

func (f funcInfo) valid() bool {
	return f._func != nil
}

func (f funcInfo) _Func() *runtime.Func {
	return (*runtime.Func)(unsafe.Pointer(f._func))
}

//go:linkname runtimeFindFunc runtime.findfunc
func runtimeFindFunc(pc uintptr) funcInfo

func PageFaultHandler(exceptionCode uintptr) {
	log.KErrorLn("\nPage Fault! Disabling Interrupt and halting!")
	log.KPrintLn("Exception code: ", uintptr(exceptionCode))
	log.KPrintLn("Present: ", (exceptionCode&PAGE_FAULT_PRESENT)>>(PAGE_FAULT_PRESENT>>1),
		" Write: ", (exceptionCode&PAGE_FAULT_WRITE)>>(PAGE_FAULT_WRITE>>1),
		" user: ", (exceptionCode&PAGE_FAULT_USER)>>(PAGE_FAULT_USER>>1),
		" instruction: ", (exceptionCode&PAGE_FAULT_INSTRUCTION_FETCH)>>(PAGE_FAULT_INSTRUCTION_FETCH>>1))
	causingAddr := getPageFaultAddr()
	log.KPrint("Causing Address: ", uintptr(causingAddr))
	f := runtimeFindFunc(uintptr(causingAddr))
	if f.valid() {
		s := f._Func().Name()
		file, line := f._Func().FileLine(uintptr(causingAddr))
		_, filename := path.Split(file)
		log.KPrint(" (", s, " (", filename, ":", line, ")", ")")
	}
	log.KPrintLn("")
	log.KPrintLn("Current Page Directory: ", (uintptr)(unsafe.Pointer(getCurrentPageDir())))
	panic.KernelPanic("Page Fault")
}

func CreateNewPageDirectory() MemSpace {
	var ret MemSpace
	addr := AllocPage()
	addr.Clear()
	ret.PageDirectory = (*PageTable)(addr.Pointer())
	for i := uint64(KERNEL_START); i < KERNEL_RESERVED; i += PAGE_SIZE {
		ret.TryMapPage(uintptr(i), uintptr(i), PAGE_RW|PAGE_PERM_KERNEL)
	}
	return ret
}

func InitPaging(memoryMaps []multiboot.MemoryMap) {
	var topAddr uint64 = 0

	memMapsSlice := memoryMaps[:]

	for _, p := range memMapsSlice {
		if p.Type != multiboot.MEM_MAP_AVAILABLE || p.Length < PAGE_SIZE || p.BaseAddr+p.Length < KERNEL_RESERVED {
			continue
		}
		if p.BaseAddr+p.Length > topAddr && p.BaseAddr+p.Length < 0x10000000 {
			topAddr = p.BaseAddr + p.Length
		}

		startAddr := uintptr(p.BaseAddr)
		if startAddr%PAGE_SIZE != 0 {
			startAddr = (startAddr + PAGE_SIZE - 1) &^ (PAGE_SIZE - 1)
		}
		if startAddr < KERNEL_RESERVED {
			startAddr = KERNEL_RESERVED
		}
		for i := startAddr; i < uintptr(p.BaseAddr+p.Length); i += PAGE_SIZE {
			FreePage(unsafe.Pointer(i))
		}
	}

	maxPages = -AllocatedPages
	AllocatedPages = 0
	addr := AllocPage()
	addr.Clear()
	kernelPageDirectory := (*PageTable)(addr.Pointer())
	KernelMemSpace.PageDirectory = kernelPageDirectory
	// printPageTable(kernelPageDirectory, 0, 1024)

	for i := uintptr(0); i < uintptr(topAddr); i += PAGE_SIZE {
		flag := uint8(PAGE_PERM_KERNEL)
		// TODO: Get address somewhere
		// if i < 0x150000 || i >= 0x199000 {
		// 	flag |= PAGE_RW
		// }
		KernelMemSpace.TryMapPage(i, i, flag)
	}
	SwitchPageDir(kernelPageDirectory)

	if PAGE_DEBUG {
		log.KDebugLn("[PAGE] Got ", maxPages, " pages. Initialization took ", AllocatedPages, " pages")
	}
	AllocatedPages = 0
	enablePaging()
}

func printPageTable(table *PageTable, start, length int) {
	for i, n := range table[start : start+length] {
		log.KDebug(uintptr(n))
		if i%16 == 15 {
			log.KDebugLn("")
		} else {
			log.KDebug(" ")
		}
	}
}
