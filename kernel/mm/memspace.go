package mm

import (
	"unsafe"

	"github.com/sanserogames/letsgo-os/kernel/log"
	"github.com/sanserogames/letsgo-os/kernel/panic"
)

const (

	// Reserve memory below 50 MB for kernel image
	KERNEL_START    = 1 << 20
	KERNEL_RESERVED = 50 << 20

	MAX_ALLOC_VIRT_ADDR = 0xf0000000
	MIN_ALLOC_VIRT_ADDR = 0x8000000
)

type MemSpace struct {
	PageDirectory *PageTable
	VmTop         uintptr
	Brk           uintptr
}

func (m *MemSpace) getPageTable(virtAddr uintptr) *PageTable {
	e := m.PageDirectory.GetEntry(virtAddr >> 10)

	if !e.IsPresent() {
		// No page table present
		addr := AllocPage()
		Memclr(addr, PAGE_SIZE)
		// User perm here to allow all entries in page table to be possibly accessed by
		// user. User does not have access to page table though.
		m.PageDirectory.SetEntry(virtAddr>>10, addr, PAGE_PRESENT|PAGE_RW|PAGE_PERM_USER)

	}

	return e.AsPageTable()
}

func (m *MemSpace) TryMapPage(page uintptr, virtAddr uintptr, flags uint8) bool {

	pt := m.getPageTable(virtAddr)
	e := pt.GetEntry(virtAddr)
	if e.IsPresent() {
		return false
	}
	pt.SetEntry(virtAddr, page, uintptr(flags)|PAGE_PRESENT)
	if virtAddr >= m.VmTop && virtAddr < 0x8000000 {
		m.VmTop = virtAddr + PAGE_SIZE
	}
	return true
}

func (m *MemSpace) MapPage(page uintptr, virtAddr uintptr, flags uint8) {
	if PAGE_DEBUG {
		log.KDebugLn("[PAGE] Mapping page ", page, " to virt addr ", virtAddr)
	}
	if !m.TryMapPage(page, virtAddr, flags) {
		log.KErrorLn("Page already present")
		log.KPrintLn(page, " -> ", virtAddr)
		panic.KernelPanic("Tried to remap a page")
	}
}

func (m *MemSpace) UnmapPage(virtAddr uintptr) {
	if PAGE_DEBUG {
		log.KDebug("[PAGE] Unmapping page ", virtAddr)
	}
	pt := m.getPageTable(virtAddr)
	e := pt.GetEntry(virtAddr)
	if e.IsPresent() {
		e.UnsetPresent()
		FreePage(e.GetPhysicalAddress())
		if PAGE_DEBUG {
			log.KDebugLn("(phys-addr: ", e.GetPhysicalAddress(), ")")
		}
	} else {
		if PAGE_DEBUG {
			log.KDebugLn("\n[PAGE] WARNING: Page was already unmapped")
		}
	}
}

func (m *MemSpace) GetPageTableEntry(virtAddr uintptr) *PageTableEntry {
	pt := m.getPageTable(virtAddr)
	return pt.GetEntry(virtAddr)
}

func (m *MemSpace) FindSpaceFor(startAddr uintptr, length uintptr) uintptr {
	if PAGE_DEBUG {
		log.KDebugLn("[PAGE] Find space for ", startAddr, " with size ", length)
	}
	if startAddr < MIN_ALLOC_VIRT_ADDR {
		if PAGE_DEBUG {
			log.KDebugLn("[PAGE] startAddr was below MIN_ALLOC_VIRT_ADDR")
		}
		startAddr = MIN_ALLOC_VIRT_ADDR
	}
	for startAddr < MAX_ALLOC_VIRT_ADDR {
		// TODO: Check if page table is not allocated and count it as free instead of getting table entry and causing the table to be allocated
		for ; m.GetPageTableEntry(startAddr).IsPresent() && startAddr < MAX_ALLOC_VIRT_ADDR; startAddr += PAGE_SIZE {
		}
		endAddr := startAddr + length
		if PAGE_DEBUG {
			log.KDebugLn("[PAGE][FIND] Trying ", startAddr, " with endaddr ", endAddr)
		}
		if endAddr > MAX_ALLOC_VIRT_ADDR {
			break
		}
		isRangeFree := true
		for i := startAddr; i < endAddr && isRangeFree; i += PAGE_SIZE {
			entry := m.GetPageTableEntry(i)
			isRangeFree = isRangeFree && !entry.IsPresent()
			if !isRangeFree {
				startAddr = i
			}
		}
		if isRangeFree {
			if PAGE_DEBUG {
				log.KDebugLn("[PAGE][FIND] Found ", startAddr)
			}
			return startAddr
		} else {
			if PAGE_DEBUG {
				log.KDebugLn("[PAGE][FIND] position did not work")
			}
		}
	}
	if PAGE_DEBUG {
		log.KDebugLn("[PAGE][FIND] Did not find suitable location ", startAddr)
	}
	return 0
}

func (m *MemSpace) GetPhysicalAddress(virtAddr uintptr) (uintptr, bool) {
	if !m.IsAddressAccessible(virtAddr) {
		return 0, false
	} else {
		if PAGE_DEBUG {
			// log.KDebugln("[PAGING] Translated address: ", virtAddr, "->", uintptr(e&^(PAGE_SIZE-1)))
		}
		e := m.GetPageTableEntry(virtAddr)
		return uintptr(e.GetPhysicalAddress()) | (virtAddr & (PAGE_SIZE - 1)), true
	}
}

func (m *MemSpace) IsAddressAccessible(virtAddr uintptr) bool {
	pt := m.getPageTable(virtAddr)
	e := pt.GetEntry(virtAddr)
	return e.IsPresent() && e.IsUserAccessible()
}

func (m *MemSpace) IsRangeAccessible(startAddr uintptr, endAddr uintptr) bool {
	for pageAddr := startAddr &^ PAGE_SIZE; pageAddr < endAddr; pageAddr += PAGE_SIZE {
		if !m.IsAddressAccessible(pageAddr) {
			return false
		}
	}
	return true
}

func (m *MemSpace) FreeAllPages() {
	for tableIdx, tableEntry := range m.PageDirectory {
		if !tableEntry.IsPresent() {
			// tables that are not present don't need to be freed
			continue
		}
		pta := tableEntry.AsPageTable()
		if tableIdx > (KERNEL_RESERVED >> 22) {
			for i, entry := range pta {
				virtAddr := uintptr((tableIdx << 22) + (i << 12))
				if !entry.IsPresent() || virtAddr <= KERNEL_RESERVED {
					// Ignore kernel reserved mappings and tables that are not present
					continue
				}
				m.UnmapPage(virtAddr)
			}
		}
		tableEntry.UnsetPresent()
		FreePage(tableEntry.GetPhysicalAddress())
	}
	FreePage(uintptr(unsafe.Pointer(m.PageDirectory)))
}
