package main

import (
	"path"
	"unsafe"
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
)

var (
	freePagesList  *page
	kernelMemSpace MemSpace = MemSpace{}
	allocatedPages int      = 0
	maxPages       int      = 0
)

type PageTable [ENTRIES_PER_TABLE]pageTableEntry

func (pt *PageTable) getEntryIndex(virtAddr uintptr) uintptr {
	return virtAddr >> 12 & ((1 << 10) - 1)
}

func (pt *PageTable) getEntry(virtAddr uintptr) *pageTableEntry {
	return &pt[pt.getEntryIndex((virtAddr))]
}

func (pt *PageTable) setEntry(virtAddr uintptr, physAddr uintptr, flags uintptr) {
	e := pageTableEntry(physAddr | flags)
	pt[pt.getEntryIndex((virtAddr))] = e
}

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

func (e pageTableEntry) getPhysicalAddress() uintptr {
	return uintptr(e &^ (PAGE_SIZE - 1))
}

func (e pageTableEntry) asPageTable() *PageTable {
	return (*PageTable)(unsafe.Pointer(e.getPhysicalAddress()))
}

type page struct {
	next *page
}

type MemSpace struct {
	PageDirectory *PageTable
	VmTop         uintptr
	Brk           uintptr
}

func (m *MemSpace) getPageTable(virtAddr uintptr) *PageTable {
	e := m.PageDirectory.getEntry(virtAddr >> 10)

	if !e.isPresent() {
		// No page table present
		addr := AllocPage()
		Memclr(addr, PAGE_SIZE)
		// User perm here to allow all entries in page table to be possibly accessed by
		// user. User does not have access to page table though.
		m.PageDirectory.setEntry(virtAddr>>10, addr, PAGE_PRESENT|PAGE_RW|PAGE_PERM_USER)

	}

	return e.asPageTable()
}

func (m *MemSpace) tryMapPage(page uintptr, virtAddr uintptr, flags uint8) bool {

	pt := m.getPageTable(virtAddr)
	e := pt.getEntry(virtAddr)
	if e.isPresent() {
		return false
	}
	pt.setEntry(virtAddr, page, uintptr(flags)|PAGE_PRESENT)
	if virtAddr >= m.VmTop && virtAddr < 0x8000000 {
		m.VmTop = virtAddr + PAGE_SIZE
	}
	return true
}

func (m *MemSpace) MapPage(page uintptr, virtAddr uintptr, flags uint8) {
	if PAGE_DEBUG {
		kdebugln("[PAGE] Mapping page ", page, " to virt addr ", virtAddr)
	}
	if !m.tryMapPage(page, virtAddr, flags) {
		kerrorln("Page already present")
		kprintln(page, " -> ", virtAddr)
		kernelPanic("Tried to remap a page")
	}
}

func (m *MemSpace) UnmapPage(virtAddr uintptr) {
	if PAGE_DEBUG {
		kdebugln("[PAGE] Unmapping page ", virtAddr)
	}
	pt := m.getPageTable(virtAddr)
	e := pt.getEntry(virtAddr)
	if e.isPresent() {
		e.unsetPresent()
		FreePage(e.getPhysicalAddress())
	} else {
		if PAGE_DEBUG {
			kdebugln("[PAGE] WARNING: Page was already unmapped")
		}
	}
}

func (m *MemSpace) getPageTableEntry(virtAddr uintptr) *pageTableEntry {
	pt := m.getPageTable(virtAddr)
	return pt.getEntry(virtAddr)
}

func (m *MemSpace) FindSpaceFor(startAddr uintptr, length uintptr) uintptr {
	if PAGE_DEBUG {
		kdebugln("[PAGE] Find space for ", startAddr, " with size ", length)
	}
	for startAddr < MAX_ALLOC_VIRT_ADDR {
		// TODO: Check if page table is not allocated and count it as free instead of getting table entry and causing the table to be allocated
		for ; m.getPageTableEntry(startAddr).isPresent() && startAddr < MAX_ALLOC_VIRT_ADDR; startAddr += PAGE_SIZE {
		}
		endAddr := startAddr + length
		if PAGE_DEBUG {
			kdebugln("[PAGE][FIND] Trying ", startAddr, " with endaddr ", endAddr)
		}
		if endAddr > MAX_ALLOC_VIRT_ADDR {
			break
		}
		isRangeFree := true
		for i := startAddr; i < endAddr && isRangeFree; i += PAGE_SIZE {
			entry := m.getPageTableEntry(i)
			isRangeFree = isRangeFree && !entry.isPresent()
			if !isRangeFree {
				startAddr = i
			}
		}
		if isRangeFree {
			if PAGE_DEBUG {
				kdebugln("[PAGE][FIND] Found ", startAddr)
			}
			return startAddr
		} else {
			if PAGE_DEBUG {
				kdebugln("[PAGE][FIND] position did not work")
			}
		}
	}
	if PAGE_DEBUG {
		kdebugln("[PAGE][FIND] Did not find suitable location ", startAddr)
	}
	return 0
}

func (m *MemSpace) GetPhysicalAddress(virtAddr uintptr) (uintptr, bool) {
	if !m.isAddressAccessible(virtAddr) {
		return 0, false
	} else {
		if PAGE_DEBUG {
			// kdebugln("[PAGING] Translated address: ", virtAddr, "->", uintptr(e&^(PAGE_SIZE-1)))
		}
		e := m.getPageTableEntry(virtAddr)
		return uintptr(e.getPhysicalAddress()) | (virtAddr & (PAGE_SIZE - 1)), true
	}
}

func (m *MemSpace) isAddressAccessible(virtAddr uintptr) bool {
	pt := m.getPageTable(virtAddr)
	e := pt.getEntry(virtAddr)
	return e.isPresent() && e.isUserAccessible()
}

func (m *MemSpace) isRangeAccessible(startAddr uintptr, endAddr uintptr) bool {
	for pageAddr := startAddr &^ PAGE_SIZE; pageAddr < endAddr; pageAddr += PAGE_SIZE {
		if !m.isAddressAccessible(pageAddr) {
			return false
		}
	}
	return true
}

func (m *MemSpace) FreeAllPages() {
	for tableIdx, tableEntry := range m.PageDirectory {
		if !tableEntry.isPresent() {
			// tables that are not present don't need to be freed
			continue
		}
		pta := tableEntry.asPageTable()
		if tableIdx > (KERNEL_RESERVED >> 22) {
			for i, entry := range pta {
				virtAddr := uintptr((tableIdx << 22) + (i << 12))
				if !entry.isPresent() || virtAddr <= KERNEL_RESERVED {
					// Ignore kernel reserved mappings and tables that are not present
					continue
				}
				m.UnmapPage(virtAddr)
			}
		}
		tableEntry.unsetPresent()
		FreePage(tableEntry.getPhysicalAddress())
	}
	FreePage(uintptr(unsafe.Pointer(m.PageDirectory)))
}

func enablePaging()

func switchPageDir(dir *PageTable)

func getCurrentPageDir() *PageTable

func getPageFaultAddr() uint32

func pageFaultHandler() {
	code := currentThread.info.ExceptionCode
	kerrorln("\nPage Fault! Disabling Interrupt and halting!")
	kprintln("Exception code: ", uintptr(code))
	kprintln("Present: ", (code&PAGE_FAULT_PRESENT)>>(PAGE_FAULT_PRESENT>>1),
		" Write: ", (code&PAGE_FAULT_WRITE)>>(PAGE_FAULT_WRITE>>1),
		" user: ", (code&PAGE_FAULT_USER)>>(PAGE_FAULT_USER>>1),
		" instruction: ", (code&PAGE_FAULT_INSTRUCTION_FETCH)>>(PAGE_FAULT_INSTRUCTION_FETCH>>1))
	causingAddr := getPageFaultAddr()
	kprint("Causing Address: ", uintptr(causingAddr))
	f := runtimeFindFunc(uintptr(causingAddr))
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

func FreePage(addr uintptr) {
	if addr%PAGE_SIZE != 0 {
		kdebugln("[PAGE] WARNING: freeingPage but is not page aligned: ", addr)
		return
	}
	// Just to check for immediate double free
	// If I were to check for double freeing correctly I would have to traverse the list
	// every time completely but that would make freeing O(n)
	if addr == uintptr(unsafe.Pointer(freePagesList)) {
		kdebugln("[Page] immediate double freeing page ", addr)
		kernelPanic("[Page] double freeing page")
	}
	p := (*page)(unsafe.Pointer(addr))
	p.next = freePagesList
	freePagesList = p
	allocatedPages--
}

func AllocPage() uintptr {
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

func CreateNewPageDirectory() MemSpace {
	var ret MemSpace
	addr := AllocPage()
	Memclr(addr, PAGE_SIZE)
	ret.PageDirectory = (*PageTable)(unsafe.Pointer(addr))
	for i := uint64(KERNEL_START); i < KERNEL_RESERVED; i += PAGE_SIZE {
		ret.tryMapPage(uintptr(i), uintptr(i), PAGE_RW|PAGE_PERM_KERNEL)
	}
	return ret
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
			FreePage(i)
		}
	}

	maxPages = -allocatedPages
	allocatedPages = 0
	addr := AllocPage()
	Memclr(addr, PAGE_SIZE)
	kernelPageDirectory := (*PageTable)(unsafe.Pointer(addr))
	kernelMemSpace.PageDirectory = kernelPageDirectory
	// printPageTable(kernelPageDirectory, 0, 1024)

	for i := uintptr(0); i < uintptr(topAddr); i += PAGE_SIZE {
		flag := uint8(PAGE_PERM_KERNEL)
		// TODO: Get address somewhere
		// if i < 0x150000 || i >= 0x199000 {
		// 	flag |= PAGE_RW
		// }
		kernelMemSpace.tryMapPage(i, i, flag)
	}
	switchPageDir(kernelPageDirectory)

	if PAGE_DEBUG {
		kdebugln("[PAGE] Got ", maxPages, " pages. Initialization took ", allocatedPages, " pages")
	}
	allocatedPages = 0
	SetInterruptHandler(0xE, pageFaultHandler, KCS_SELECTOR, PRIV_USER)
	enablePaging()
}

func printPageTable(table *PageTable, start, length int) {
	for i, n := range table[start : start+length] {
		kdebug(uintptr(n))
		if i%16 == 15 {
			kdebugln("")
		} else {
			kdebug(" ")
		}
	}
}
