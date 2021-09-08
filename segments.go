package main

import(
    "unsafe"
    "reflect"
)

type GdtEntry struct {
    limitLow          uint16
    baseLow           uint16
    baseMid           uint8
    access            uint8
    limitHighAndFlags uint8
    baseHigh          uint8
}

type GdtSegment struct {
    base uintptr
    limit uintptr
    access uint8
    flags uint8
}

type GdtDescriptor struct {
    GdtSize        uint16
    GdtAddressLow  uint16
    GdtAddressHigh uint16
}

type UserDesc struct {
    EntryNumber uint32
    BaseAddr uint32
    Limit uint32
    Flags uint8
}

const (
    KCS_INDEX = 1
    KDS_INDEX = 2
    KGS_INDEX = 3

    KCS_SELECTOR = KCS_INDEX * 8
    KDS_SELECTOR = KDS_INDEX * 8
    KGS_SELECTOR = KGS_INDEX * 8

    // flags
    SEG_GRAN_BYTE = 0 << 7
    SEG_GRAN_4K_PAGE = 1 << 7

    // access
    SEG_NORW   = 0 << 1
    SEG_R      = 1 << 1
    SEG_W      = 1 << 1

    SEG_SYSTEM = 0 << 4
    SEG_NORMAL = 1 << 4

    SEG_NOEXEC = 0 << 3
    SEG_EXEC   = 1 << 3

    PRIV_KERNEL = 0 << 5
    PRIV_USER   = 3 << 5

    PRESENT     = 1 << 7

    // userDesc flags
    UDESC_32SEG = 1 << 0 // I ignore this
    UDESC_CONTENTS = 3 << 1 // don't know what those are
    UDESC_RX_ONLY = 1 << 3
    UDESC_LIMIT_IN_PAGES = 1 << 4
    UDESC_SEG_NOT_PRESENT = 1 << 5
    UDESC_USABLE = 1 << 6
)

var (
    gdtTable = [256]GdtEntry{}
    gdtDescriptor GdtDescriptor
    gdtTableLen = 0
)

func installGDT(descriptor *GdtDescriptor)

func getGDT() *GdtDescriptor

func InitSegments() {
    var gdtEntry GdtEntry
    gdtDescriptor = *getGDT()
    oldGdtAddr := uintptr(uint32(gdtDescriptor.GdtAddressLow) |
                    uint32(gdtDescriptor.GdtAddressHigh) << 16)
    oldGdtLen := int(uintptr(gdtDescriptor.GdtSize+1) / unsafe.Sizeof(gdtEntry))
    oldGdt := *(*[]GdtEntry)(unsafe.Pointer(&reflect.SliceHeader{
        Len:  oldGdtLen,
	    Cap:  oldGdtLen,
	    Data: oldGdtAddr,
    }))
    copy(gdtTable[:], oldGdt)
    gdtTableLen += oldGdtLen
    gdtAddr := uint32(uintptr(unsafe.Pointer(&gdtTable)))
    gdtDescriptor.GdtAddressLow = uint16(gdtAddr)
    gdtDescriptor.GdtAddressHigh = uint16(gdtAddr >> 16)
    installGDT(&gdtDescriptor)
    //printGdt(gdtTable[:gdtTableLen])
}

func GetSegment(index int, res *GdtSegment) int {
    if index < 0 || index > gdtTableLen { return -1 }
    (*res).limit = uintptr(gdtTable[index].limitLow)
    (*res).limit += uintptr((gdtTable[index].limitHighAndFlags & 0xF)) << 16
    (*res).base = uintptr(gdtTable[index].baseLow)
    (*res).base += uintptr(gdtTable[index].baseMid) << 16
    (*res).base += uintptr(gdtTable[index].baseHigh) << 24
    (*res).access = gdtTable[index].access
    (*res).flags = gdtTable[index].limitHighAndFlags & 0xF0
    return 0

}

func AddSegment(base uintptr, limit uintptr, access uint8, flags uint8) int {
    gdtTable[gdtTableLen].limitHighAndFlags = uint8((limit >> 16) & 0x000F)
    bigMode := uint8(0)

    if access & SEG_NORMAL != 0 {
        bigMode = 1 << 6
    }
    gdtTable[gdtTableLen].limitHighAndFlags |= flags | bigMode
    gdtTable[gdtTableLen].access = access | PRESENT
    gdtTable[gdtTableLen].baseMid = uint8(base >> 16)
    gdtTable[gdtTableLen].baseHigh = uint8(base >> 24)

    gdtTable[gdtTableLen].baseLow = uint16(base & 0xFFFF)
    gdtTable[gdtTableLen].limitLow = uint16(limit  & 0x0000FFFF)
    gdtTableLen++
    return gdtTableLen-1
}

func UpdateGdt(){
    gdtDescriptor.GdtSize = uint16(gdtTableLen*8 - 1)
    installGDT(&gdtDescriptor)
}

func printGdt(gdt []GdtEntry){
    for _,n := range gdt {
        text_mode_print_hex16(n.limitLow)
        text_mode_print(" ")
        text_mode_print_hex16(n.baseLow)
        text_mode_print(" ")
        text_mode_print_hex(n.baseMid)
        text_mode_print(" ")
        text_mode_print_hex(n.access)
        text_mode_print(" ")
        text_mode_print_hex(n.limitHighAndFlags)
        text_mode_print(" ")
        text_mode_print_hex(n.baseHigh)
        text_mode_println("")
    }
}
