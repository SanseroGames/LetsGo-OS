package main

import (
	"unsafe"
)

type GdtEntry struct {
	limitLow          uint16
	baseLow           uint16
	baseMid           uint8
	access            uint8
	limitHighAndFlags uint8
	baseHigh          uint8
}

func (e *GdtEntry) IsPresent() bool {
	return e.access&PRESENT != 0
}

func (e *GdtEntry) Fill(base uint32, limit uint32, access uint8, flags uint8) {
	e.limitHighAndFlags = uint8((limit >> 16) & 0x000F)
	e.limitHighAndFlags |= flags
	e.access = access
	e.baseMid = uint8(base >> 16)
	e.baseHigh = uint8(base >> 24)
	e.baseLow = uint16(base & 0xFFFF)
	e.limitLow = uint16(limit & 0x0000FFFF)
}

func (e *GdtEntry) Clear() {
	e.limitHighAndFlags = 0
	e.access = 0
	e.baseMid = 0
	e.baseHigh = 0
	e.baseLow = 0
	e.limitLow = 0

}

type GdtSegment struct {
	base   uintptr
	limit  uintptr
	access uint8
	flags  uint8
}

func (s *GdtSegment) IsPresent() bool {
	return s.access&PRESENT != 0
}

func (s *GdtSegment) Clear() {
	s.base = 0
	s.limit = 0
	s.access = 0
	s.flags = 0
}

type GdtDescriptor struct {
	GdtSize        uint16
	GdtAddressLow  uint16
	GdtAddressHigh uint16
}

type UserDesc struct {
	EntryNumber uint32
	BaseAddr    uint32
	Limit       uint32
	Flags       uint8
}

type TlsEntry struct {
	Desc    UserDesc
	Present bool
}

const (
	KCS_INDEX = 1
	KDS_INDEX = 2
	KGS_INDEX = 3

	KCS_SELECTOR = KCS_INDEX * 8
	KDS_SELECTOR = KDS_INDEX * 8
	KGS_SELECTOR = KGS_INDEX * 8

	// flags
	SEG_GRAN_BYTE    = 0 << 7
	SEG_GRAN_4K_PAGE = 1 << 7

	// access
	SEG_NORW = 0 << 1
	SEG_R    = 1 << 1
	SEG_W    = 1 << 1

	SEG_SYSTEM = 0 << 4
	SEG_NORMAL = 1 << 4

	SEG_NOEXEC = 0 << 3
	SEG_EXEC   = 1 << 3

	PRIV_KERNEL = 0 << 5
	PRIV_USER   = 3 << 5

	SEG_BIG_MODE = 1 << 6

	PRESENT = 1 << 7

	// userDesc flags
	UDESC_32SEG           = 1 << 0 // I ignore this
	UDESC_CONTENTS        = 3 << 1 // don't know what those are
	UDESC_RX_ONLY         = 1 << 3
	UDESC_LIMIT_IN_PAGES  = 1 << 4
	UDESC_SEG_NOT_PRESENT = 1 << 5
	UDESC_USABLE          = 1 << 6

	// Misc
	GDT_ENTRIES = 256
	TLS_START   = 16
)

var (
	gdtTable      = [GDT_ENTRIES]GdtEntry{}
	gdtDescriptor GdtDescriptor
	gdtTableLen   = 0
)

func installGDT(descriptor *GdtDescriptor)

func getGDT() *GdtDescriptor

func InitSegments() {
	var gdtEntry GdtEntry

	gdtTableSlice := gdtTable[:]

	Memclr(uintptr(unsafe.Pointer(&gdtTable)), len(gdtTable))

	gdtDescriptor = *getGDT()
	oldGdtAddr := uintptr(uint32(gdtDescriptor.GdtAddressLow) |
		uint32(gdtDescriptor.GdtAddressHigh)<<16)
	oldGdtLen := int(uintptr(gdtDescriptor.GdtSize+1) / unsafe.Sizeof(gdtEntry))
	oldGdt := unsafe.Slice((*GdtEntry)(unsafe.Pointer(oldGdtAddr)), oldGdtLen)
	copy(gdtTableSlice, oldGdt)
	gdtAddr := uintptr(unsafe.Pointer(&gdtTable))
	gdtDescriptor.GdtAddressLow = uint16(gdtAddr)
	gdtDescriptor.GdtAddressHigh = uint16(gdtAddr >> 16)
	gdtDescriptor.GdtSize = GDT_ENTRIES - 1
	gdtTableLen += oldGdtLen
	installGDT(&gdtDescriptor)
	//printGdt(gdtTable[:gdtTableLen])
}

func GetSegment(index int, res *GdtSegment) int {
	if index < 0 || index > len(gdtTable) {
		return -1
	}
	res.limit = uintptr(gdtTable[index].limitLow)
	res.limit += uintptr((gdtTable[index].limitHighAndFlags & 0xF)) << 16
	res.base = uintptr(gdtTable[index].baseLow)
	res.base += uintptr(gdtTable[index].baseMid) << 16
	res.base += uintptr(gdtTable[index].baseHigh) << 24
	res.access = gdtTable[index].access
	res.flags = gdtTable[index].limitHighAndFlags & 0xF0
	return 0

}

func AddSegment(base uintptr, limit uintptr, access uint8, flags uint8) int {
	if gdtTableLen < 0 || gdtTableLen > TLS_START {
		return -1
	}
	doUpdateSegment(gdtTableLen, base, limit, access, flags)
	gdtTableLen++
	return gdtTableLen - 1
}

func UpdateSegment(index int, base uintptr, limit uintptr, access uint8, flags uint8) bool {
	if index < 0 || index > gdtTableLen {
		return false
	}
	doUpdateSegment(index, base, limit, access, flags)
	return true
}

// Does not check if index in range
func doUpdateSegment(index int, base uintptr, limit uintptr, access uint8, flags uint8) {
	gdtTable[index].limitHighAndFlags = uint8((limit >> 16) & 0x000F)
	bigMode := uint8(0)

	if access&SEG_NORMAL != 0 {
		bigMode = 1 << 6
	}
	gdtTable[index].limitHighAndFlags |= flags | bigMode
	gdtTable[index].access = access | PRESENT
	gdtTable[index].baseMid = uint8(base >> 16)
	gdtTable[index].baseHigh = uint8(base >> 24)

	gdtTable[index].baseLow = uint16(base & 0xFFFF)
	gdtTable[index].limitLow = uint16(limit & 0x0000FFFF)

}

func SetTlsSegment(index uint32, desc *UserDesc, table []GdtEntry) bool {
	if index < TLS_START || index > uint32(len(gdtTable)) {
		return false
	}

	if desc.Flags == UDESC_RX_ONLY|UDESC_SEG_NOT_PRESENT {
		// Desc is considered emtpy. Clear entry
		table[index].Clear()
		return true
	}

	flags := uint8(SEG_GRAN_BYTE)
	if desc.Flags&UDESC_LIMIT_IN_PAGES != 0 {
		flags = SEG_GRAN_4K_PAGE
	}
	access := uint8(PRIV_USER | SEG_NORMAL | PRESENT)
	if desc.Flags&UDESC_RX_ONLY != 0 {
		access |= SEG_EXEC
	} else {
		access |= SEG_W
	}
	if access&SEG_NORMAL != 0 {
		flags |= SEG_BIG_MODE
	}

	table[index].Fill(desc.BaseAddr, desc.Limit, access, flags)
	FlushTlsTable(table)

	return true
}

func FlushTlsTable(table []GdtEntry) {
	copy(gdtTable[TLS_START:], table[TLS_START:])
}

func printGdt(gdt []GdtEntry) {
	for _, n := range gdt {
		kdebugln(n.limitLow, " ", n.baseLow, " ", n.baseMid, " ", n.access, " ", n.limitHighAndFlags, " ", n.baseHigh)
	}
}
