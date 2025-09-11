package kernel

import (
	"path"
	"unsafe"

	"github.com/sanserogames/letsgo-os/kernel/log"
	"github.com/sanserogames/letsgo-os/kernel/mm"
)

const (
	PAGE_DEBUG = false

	// Reserve memory below 50 MB for kernel image
	KERNEL_START      = 1 << 20
	KERNEL_RESERVED   = 50 << 20
	PAGE_SIZE         = 4 << 10
	ENTRIES_PER_TABLE = PAGE_SIZE / 4

	PAGE_PRESENT       = 1 << 0
	PAGE_RW            = 1 << 1
	PAGE_PERM_USER     = 1 << 2
	PAGE_PERM_KERNEL   = 0 << 2
	PAGE_WRITETHROUGH  = 1 << 3
	PAGE_DISABLE_CACHE = 1 << 4

	PAGE_FAULT_PRESENT           = 1 << 0
	PAGE_FAULT_WRITE             = 1 << 1
	PAGE_FAULT_USER              = 1 << 2
	PAGE_FAULT_INSTRUCTION_FETCH = 1 << 4

	MAX_ALLOC_VIRT_ADDR = 0xf0000000
	MIN_ALLOC_VIRT_ADDR = 0x8000000
)

var (
	kernelMemSpace mm.MemSpace = mm.MemSpace{}
	maxPages       int         = 0
)

func enablePaging()

func switchPageDir(dir *mm.PageTable)

func getCurrentPageDir() *mm.PageTable

func getPageFaultAddr() uint32

func pageFaultHandler() {
	code := CurrentThread.info.ExceptionCode
	log.KErrorLn("\nPage Fault! Disabling Interrupt and halting!")
	log.KPrintLn("Exception code: ", uintptr(code))
	log.KPrintLn("Present: ", (code&PAGE_FAULT_PRESENT)>>(PAGE_FAULT_PRESENT>>1),
		" Write: ", (code&PAGE_FAULT_WRITE)>>(PAGE_FAULT_WRITE>>1),
		" user: ", (code&PAGE_FAULT_USER)>>(PAGE_FAULT_USER>>1),
		" instruction: ", (code&PAGE_FAULT_INSTRUCTION_FETCH)>>(PAGE_FAULT_INSTRUCTION_FETCH>>1))
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
	kernelPanic("Page Fault")
}

func CreateNewPageDirectory() mm.MemSpace {
	var ret mm.MemSpace
	addr := mm.AllocPage()
	addr.Clear()
	ret.PageDirectory = (*mm.PageTable)(addr.Pointer())
	for i := uint64(KERNEL_START); i < KERNEL_RESERVED; i += PAGE_SIZE {
		ret.TryMapPage(uintptr(i), uintptr(i), PAGE_RW|PAGE_PERM_KERNEL)
	}
	return ret
}

func InitPaging() {
	var topAddr uint64 = 0

	memMapsSlice := memoryMaps[:]

	for _, p := range memMapsSlice {
		if p.Type != MEM_MAP_AVAILABLE || p.Length < PAGE_SIZE || p.BaseAddr+p.Length < KERNEL_RESERVED {
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
			mm.FreePage(i)
		}
	}

	maxPages = -mm.AllocatedPages
	mm.AllocatedPages = 0
	addr := mm.AllocPage()
	addr.Clear()
	kernelPageDirectory := (*mm.PageTable)(addr.Pointer())
	kernelMemSpace.PageDirectory = kernelPageDirectory
	// printPageTable(kernelPageDirectory, 0, 1024)

	for i := uintptr(0); i < uintptr(topAddr); i += PAGE_SIZE {
		flag := uint8(PAGE_PERM_KERNEL)
		// TODO: Get address somewhere
		// if i < 0x150000 || i >= 0x199000 {
		// 	flag |= PAGE_RW
		// }
		kernelMemSpace.TryMapPage(i, i, flag)
	}
	switchPageDir(kernelPageDirectory)

	if PAGE_DEBUG {
		log.KDebugLn("[PAGE] Got ", maxPages, " pages. Initialization took ", mm.AllocatedPages, " pages")
	}
	mm.AllocatedPages = 0
	SetInterruptHandler(0xE, pageFaultHandler, KCS_SELECTOR, PRIV_USER)
	enablePaging()
}

func printPageTable(table *mm.PageTable, start, length int) {
	for i, n := range table[start : start+length] {
		log.KDebug(uintptr(n))
		if i%16 == 15 {
			log.KDebugLn("")
		} else {
			log.KDebug(" ")
		}
	}
}
