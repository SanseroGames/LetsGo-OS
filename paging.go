package main

import (
	"path"
	"unsafe"
)

const (
	PAGE_DEBUG = true

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
)

var (
	freePagesList  *page
	kernelMemSpace MemSpace = MemSpace{}
	allocatedPages int      = 0
)

type PageTable [ENTRIES_PER_TABLE]pageTableEntry

type pageTableEntry uintptr

func (e pageTableEntry) isPresent() bool {
	return e&PAGE_PRESENT == 1
}

func (e *pageTableEntry) unsetPresent() {
	*e = (*e) &^ PAGE_PRESENT
}

func (e pageTableEntry) isUserAccessible() bool {
	return e&PAGE_PERM_USER > 0
}

type page struct {
	next *page
}

type MemSpace struct {
	PageDirectory *PageTable
	VmTop         uintptr
	Brk           uintptr
}

func enablePaging()

func switchPageDir(dir *PageTable)

func getCurrentPageDir() *PageTable

func getPageFaultAddr() uint32

func pageFaultHandler() {
	kerrorln("\nPage Fault! Disabling Interrupt and halting!")
	kprintln("Exception code: ", uintptr(currentThread.info.ExceptionCode))
	causingAddr := getPageFaultAddr()
	kprint("Causing Address: ", uintptr(causingAddr))
	f := findfuncTest(uintptr(causingAddr))
	if f.valid() {
		s := f._Func().Name()
		file, line := f._Func().FileLine(uintptr(causingAddr))
		_, filename := path.Split(file)
		kprint(" (", s, " (", filename, ":", line, ")", ")")
	}
	kprintln("")
	kprintln("Current Page Directory: ", (uintptr)(unsafe.Pointer(getCurrentPageDir())))
	kernelPanic("Page Fault")
}

func Memclr(p uintptr, n int) {
	s := (*(*[1 << 30]byte)(unsafe.Pointer(p)))[:n]
	// the compiler will emit runtime.memclrNoHeapPointers
	for i := range s {
		s[i] = 0
	}
}

func freePage(addr uintptr) {
	if addr%PAGE_SIZE != 0 {
		kdebugln("[PAGE] WARNING: freeingPage but is not page aligned: ", addr)
		return
	}
	// Just to check for immediate double free
	// If I were to check for double freeing correctly I would have to traverse the list
	// everytime completely but that would make freeing O(n)
	if addr == uintptr(unsafe.Pointer(freePagesList)) {
		kdebugln("[Page] immediate double freeing page ", addr)
		kernelPanic("[Page] double freeing page")
	}
	p := (*page)(unsafe.Pointer(addr))
	p.next = freePagesList
	freePagesList = p
	allocatedPages--
}

func allocPage() uintptr {
	if freePagesList == nil {
		kernelPanic("[PAGE] Out of pages to allocate")
	}
	p := freePagesList
	freePagesList = p.next
	allocatedPages++
	if PAGE_DEBUG {
		kdebugln("[PAGE]: Allocated ", unsafe.Pointer(p))
	}
	return uintptr(unsafe.Pointer(p))
}

func createNewPageDirectory() MemSpace {
	var ret MemSpace
	ret.PageDirectory = (*PageTable)(unsafe.Pointer(allocPage()))
	Memclr(uintptr(unsafe.Pointer(ret.PageDirectory)), PAGE_SIZE)
	for i := uint64(KERNEL_START); i < KERNEL_RESERVED; i += PAGE_SIZE {
		ret.tryMapPage(uintptr(i), uintptr(i), PAGE_RW|PAGE_PERM_KERNEL)
	}
	return ret
}

// Would be better to return addres when multitasking
func (m *MemSpace) getPageTable(page uintptr) pageTableEntry {
	directoryIndex := page >> 22
	e := m.PageDirectory[directoryIndex]
	if e.isPresent() {
		return e
	}
	// No pagetable present
	addr := allocPage()
	Memclr(addr, PAGE_SIZE)
	// User perm here to allow all entries in page table to be possibly accessed by
	// user. User does not have access to page table though.
	e = pageTableEntry(addr | PAGE_PRESENT | PAGE_RW | PAGE_PERM_USER)
	m.PageDirectory[directoryIndex] = e
	return e
}

func (m *MemSpace) tryMapPage(page uintptr, virtAddr uintptr, flags uint8) bool {
	pt := m.getPageTable(virtAddr)
	pta := (*PageTable)(unsafe.Pointer(pt &^ ((1 << 12) - 1)))
	pageIndex := virtAddr >> 12 & ((1 << 10) - 1)
	e := pta[pageIndex]
	if e.isPresent() {
		return false
	}
	e = pageTableEntry(page | uintptr(flags) | PAGE_PRESENT)
	pta[pageIndex] = e
	if virtAddr >= m.VmTop && virtAddr < 0x80000000 {
		m.VmTop = virtAddr + PAGE_SIZE
	}
	return true
}

func (m *MemSpace) mapPage(page uintptr, virtAddr uintptr, flags uint8) {
	if PAGE_DEBUG {
		kdebugln("[PAGE] Mapping page ", page, " to virt addr ", virtAddr)
	}
	if !m.tryMapPage(page, virtAddr, flags) {
		kerrorln("Page already present")
		kprintln(uint32(page), " -> ", uint32(virtAddr))
		kernelPanic("Tried to remap a page")
	}
}

func (m *MemSpace) unMapPage(virtAddr uintptr) {
	//kdebugln("[PAGE] Unmapping page ", virtAddr)
	pt := m.getPageTable(virtAddr)
	pta := (*PageTable)(unsafe.Pointer(pt &^ ((1 << 12) - 1)))
	pageIndex := virtAddr >> 12 & ((1 << 10) - 1)
	e := pta[pageIndex]
	if e.isPresent() {
		phAddr := uintptr(e&^(PAGE_SIZE-1)) | (virtAddr & (PAGE_SIZE - 1))
		pta[pageIndex].unsetPresent()
		freePage(phAddr)
	} else {
		if PAGE_DEBUG {
			kdebugln("[PAGE] WARNING: Page was already unmapped")
		}
	}
}

func (m *MemSpace) getPhysicalAddress(virtAddr uintptr) (uintptr, bool) {
	pageAddr := virtAddr &^ (PAGE_SIZE - 1)
	pt := m.getPageTable(pageAddr)
	pta := (*PageTable)(unsafe.Pointer(pt &^ ((1 << 12) - 1)))
	pageIndex := pageAddr >> 12 & ((1 << 10) - 1)
	e := pta[pageIndex]
	if !e.isPresent() || !e.isUserAccessible() {
		return 0, false
	} else {
		if PAGE_DEBUG {
			kdebugln("[PAGING] Translated address: ", virtAddr, "->", uintptr(e&^(PAGE_SIZE-1)))
		}
		return uintptr(e&^(PAGE_SIZE-1)) | (virtAddr & (PAGE_SIZE - 1)), true
	}
}

func (m *MemSpace) freeAllPages() {
	for tableIdx, table := range m.PageDirectory {
		if !table.isPresent() {
			// Ignore kernel reserved mappings and tables that are not present
			continue
		}
		pta := (*PageTable)(unsafe.Pointer(table &^ ((1 << 12) - 1)))
		if tableIdx > (KERNEL_RESERVED >> 22) {
			for i, entry := range pta {
				virtAddr := uintptr((tableIdx << 22) + (i << 12))
				if !entry.isPresent() || virtAddr <= KERNEL_RESERVED {
					continue
				}
				m.unMapPage(virtAddr)
			}
		}
		table.unsetPresent()
		freePage(uintptr(unsafe.Pointer(pta)))
	}
	freePage(uintptr(unsafe.Pointer(m.PageDirectory)))
}

func InitPaging() {
	var topAddr uint64 = 0
	for _, p := range memoryMaps {
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
			freePage(i)
		}
	}

	kernelPageDirectory := (*PageTable)(unsafe.Pointer(allocPage()))
	Memclr(uintptr(unsafe.Pointer(kernelPageDirectory)), PAGE_SIZE)
	kernelMemSpace.PageDirectory = kernelPageDirectory
	//for _,p := range memoryMaps {
	//    startAddr := p.BaseAddr &^ (PAGE_SIZE - 1)
	//    lastAddr := p.BaseAddr + p.Length - 1
	//    for i:= startAddr; i < lastAddr; i += PAGE_SIZE {
	//        tryMapPage(uintptr(i), uintptr(i), PAGE_RW | PAGE_PERM_KERNEL , kernelPageDirectory)
	//    }
	//}
	//text_mode_print_hex32(uint32(topAddr))
	//text_mode_println("")
	//text_mode_print_hex32(uint32(uintptr(unsafe.Pointer(kernelPageDirectory))))
	//text_mode_println("")
	for i := uint64(0); i < topAddr; i += PAGE_SIZE {
		flag := uint8(PAGE_PERM_KERNEL)
		// TODO: Get address somewhere
		// if i < 0x150000 || i >= 0x199000 {
		// 	flag |= PAGE_RW
		// }
		kernelMemSpace.tryMapPage(uintptr(i), uintptr(i), flag)
	}
	switchPageDir(kernelPageDirectory)
	allocatedPages = 0
	SetInterruptHandler(0xE, pageFaultHandler, KCS_SELECTOR, PRIV_USER)
	enablePaging()
	//printPageTable((*PageTable)(unsafe.Pointer(kernelPageDirectory[0] &^ ((1 << 12) - 1))))
}

func printPageTable(table *PageTable, start, length int) {
	// length of 192 fills the screen
	for i, n := range table[start : start+length] {
		text_mode_print_hex32(uint32(n))
		if i%8 == 7 {
			text_mode_println("")
		} else {
			text_mode_print(" ")
		}
	}
}
