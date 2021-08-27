package main

import (
    "unsafe"
)

type tssEntry struct {
	prev_tss uint32 // The previous TSS - with hardware task switching these form a kind of backward linked list.
	esp0 uint32 // The stack pointer to load when changing to kernel mode.
	ss0 uint32     // The stack segment to load when changing to kernel mode.
	// Everything below here is unused.
	esp1 uint32 // esp and ss 1 and 2 would be used when switching to rings 1 or 2.
    ss1 uint32
	esp2 uint32
	ss2 uint32
	cr3 uint32
	eip uint32
	eflags uint32
	eax uint32
	ecx uint32
	edx uint32
	ebx uint32
	esp uint32
	ebp uint32
	esi uint32
	edi uint32
	es uint32
	cs uint32
	ss uint32
	ds uint32
	fs uint32
	gs uint32
	ldt uint32
	trap uint16
	iomap_base uint16
}

type SegmentList struct {
	es uint32
	cs uint32
	ss uint32
	ds uint32
	fs uint32
	gs uint32

}

const (
    userModeBase = 0x00000000
    userModeLimit = 0xFFFFFF
    TSS_IS_TSS = 1 << 0
    TSS_32MODE = 1 << 3
)

var (
    tss tssEntry
    user userspace = userspace{}
)

type userspace struct {
    Segments SegmentList
    MemorySpace MemSpace
}

// Don't forget to multiply by 8. it is not array index.
func flushTss(segmentIndex int)

func JumpUserMode (segments SegmentList, funcAddr uintptr, stackAddr uintptr)

func hackyGetFuncAddr(funcAddr func()) uintptr

func JumpUserModeFunc (segments SegmentList, funcAddr func(), stackAddr uintptr){
    JumpUserMode(segments, hackyGetFuncAddr(funcAddr), stackAddr)
}

func InitUserMode(kernelStackStart uintptr) {             // Add usermode code segment
    userModeCS := AddSegment(userModeBase, userModeLimit, PRIV_USER | SEG_EXEC | SEG_R | SEG_NORMAL, SEG_GRAN_4K_PAGE)
    // Add usermode data segment
    userModeDS := AddSegment(userModeBase, userModeLimit, PRIV_USER | SEG_NOEXEC | SEG_W | SEG_NORMAL, SEG_GRAN_4K_PAGE)

    user.Segments.cs = uint32(userModeCS*8)
    user.Segments.ss = uint32(userModeDS*8)
    user.Segments.ds = uint32(userModeDS*8)
    user.Segments.es = uint32(userModeDS*8)
    user.Segments.fs = uint32(userModeDS*8)
    user.Segments.gs = uint32(userModeDS*8)

    user.MemorySpace = createNewPageDirectory()

    tssIndex := AddSegment(uintptr(unsafe.Pointer(&tss)), unsafe.Sizeof(tss), PRIV_KERNEL | SEG_NORW | TSS_32MODE | SEG_SYSTEM | TSS_IS_TSS, SEG_GRAN_BYTE)


    UpdateGdt()
    tss.ss0 = KDS_SELECTOR
    tss.esp0 = uint32(kernelStackStart)
    flushTss(tssIndex*8)
}

