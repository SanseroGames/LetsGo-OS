package mm

import (
	"unsafe"

	"github.com/sanserogames/letsgo-os/kernel/log"
	"github.com/sanserogames/letsgo-os/kernel/panic"
)

const (
	PAGE_DEBUG = false

	// Reserve memory below 50 MB for kernel image
	PAGE_SIZE         = 4 << 10
	ENTRIES_PER_TABLE = PAGE_SIZE / 4

	PAGE_PRESENT       = 1 << 0
	PAGE_RW            = 1 << 1
	PAGE_PERM_USER     = 1 << 2
	PAGE_PERM_KERNEL   = 0 << 2
	PAGE_WRITETHROUGH  = 1 << 3
	PAGE_DISABLE_CACHE = 1 << 4
)

type PageTable [ENTRIES_PER_TABLE]PageTableEntry

func (pt *PageTable) GetEntryIndex(virtAddr uintptr) uintptr {
	return virtAddr >> 12 & ((1 << 10) - 1)
}

func (pt *PageTable) GetEntry(virtAddr uintptr) *PageTableEntry {
	return &pt[pt.GetEntryIndex((virtAddr))]
}

func (pt *PageTable) SetEntry(virtAddr uintptr, physAddr uintptr, flags uintptr) {
	e := PageTableEntry(physAddr | flags)
	pt[pt.GetEntryIndex((virtAddr))] = e
}

type PageTableEntry uintptr

func (e PageTableEntry) IsPresent() bool {
	return e&PAGE_PRESENT == 1
}

func (e *PageTableEntry) UnsetPresent() {
	*e = (*e) &^ PAGE_PRESENT
}

func (e PageTableEntry) IsUserAccessible() bool {
	return e&PAGE_PERM_USER > 0
}

func (e PageTableEntry) GetPhysicalAddress() uintptr {
	return uintptr(e &^ (PAGE_SIZE - 1))
}

func (e PageTableEntry) AsPageTable() *PageTable {
	return (*PageTable)(unsafe.Pointer(e.GetPhysicalAddress()))
}

type page struct {
	next *page
}

var freePagesList *page
var AllocatedPages int = 0

func FreePage(addr uintptr) {
	if addr%PAGE_SIZE != 0 {
		log.KDebugLn("[PAGE] WARNING: freeingPage but is not page aligned: ", addr)
		return
	}
	// Just to check for immediate double free
	// If I were to check for double freeing correctly I would have to traverse the list
	// every time completely but that would make freeing O(n)
	if addr == uintptr(unsafe.Pointer(freePagesList)) {
		log.KDebugLn("[Page] immediate double freeing page ", addr)
		panic.KernelPanic("[Page] double freeing page")
	}
	p := (*page)(unsafe.Pointer(addr))
	p.next = freePagesList
	freePagesList = p
	AllocatedPages--
}

func AllocPage() uintptr {
	if freePagesList == nil {
		panic.KernelPanic("[PAGE] Out of pages to allocate")
	}
	p := freePagesList
	freePagesList = p.next
	AllocatedPages++
	if PAGE_DEBUG {
		log.KDebugLn("[PAGE]: Allocated ", unsafe.Pointer(p))
	}
	return uintptr(unsafe.Pointer(p))
}

func Memclr(p uintptr, n int) {
	s := (*(*[1 << 30]byte)(unsafe.Pointer(p)))[:n]
	// the compiler will emit runtime.memclrNoHeapPointers
	for i := range s {
		s[i] = 0
	}
}
