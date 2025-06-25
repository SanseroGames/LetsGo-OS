package mm

import (
	"github.com/sanserogames/letsgo-os/kernel/utils"
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
	return utils.UIntToPointer[PageTable](e.GetPhysicalAddress())
}
