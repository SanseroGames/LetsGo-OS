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
    defaultStackPages = 16
    defaultStackStart = uintptr(0xffffc000)
)

var (
    tss tssEntry
    defaultUserSegments SegmentList
)

// Don't forget to multiply by 8. it is not array index.
func flushTss(segmentIndex int)

func JumpUserMode (segments SegmentList, funcAddr uintptr, stackAddr uintptr)

func hackyGetFuncAddr(funcAddr func()) uintptr

func JumpUserModeFunc (segments SegmentList, funcAddr func(), stackAddr uintptr){
    JumpUserMode(segments, hackyGetFuncAddr(funcAddr), stackAddr)
}


// Need pointer as this function should not do any memory allocations
func StartProgram(path string, outDomain *domain, outMainThread *thread) int {
    if outDomain == nil || outMainThread == nil {
        text_mode_print_errorln("Cannot start program. Plase allocate the memory for me")
        return 1
    }
    outDomain.Segments = defaultUserSegments
    outDomain.MemorySpace = createNewPageDirectory()
    outDomain.CurThread = outMainThread

    outMainThread.domain = outDomain
    outMainThread.next = outMainThread
    outMainThread.info.CS = defaultUserSegments.cs | 3
    outMainThread.info.SS = defaultUserSegments.ss | 3
    outMainThread.regs.GS = defaultUserSegments.gs | 3
    outMainThread.regs.FS = defaultUserSegments.fs | 3
    outMainThread.regs.ES = defaultUserSegments.es | 3
    outMainThread.regs.DS = defaultUserSegments.ds | 3
    outMainThread.fpOffset = 0xffffffff

    elfHdr, loadAddr, topAddr := LoadElfFile(path, &outDomain.MemorySpace)

    outDomain.MemorySpace.Brk = topAddr

    if elfHdr == nil {
        text_mode_print_errorln("Could not load elf file")
        return 2
    }
    var stackPages [defaultStackPages]uintptr
    for i:=0; i < defaultStackPages; i++ {
        stack := allocPage()
        Memclr(stack, PAGE_SIZE)
        outDomain.MemorySpace.mapPage(stack, defaultStackStart-uintptr((i+1)*PAGE_SIZE), PAGE_RW | PAGE_PERM_USER)
        stackPages[i] = stack
    }

    var aux [32]auxVecEntry
    nrVec := LoadAuxVector(aux[:], elfHdr, loadAddr)
    nrVec += nrVec % 2
    vecByteSize := nrVec*int(unsafe.Sizeof(aux[0]))

    stack := (*[1 << 15]uint32)(unsafe.Pointer(stackPages[0]))[:PAGE_SIZE/4]
    for i,n := range aux[:nrVec] {
        index := PAGE_SIZE/4-1-vecByteSize/4+i*2
        stack[index] = n.Type
        stack[index+1] = n.Value
    }
    outMainThread.info.EIP = elfHdr.Entry
    outMainThread.info.ESP = uint32(defaultStackStart) - 4 - uint32(vecByteSize) -4-4-4
    return 0
}


func InitUserMode(kernelStackStart uintptr) {             // Add usermode code segment
    userModeCS := AddSegment(userModeBase, userModeLimit, PRIV_USER | SEG_EXEC | SEG_R | SEG_NORMAL, SEG_GRAN_4K_PAGE)
    // Add usermode data segment
    userModeDS := AddSegment(userModeBase, userModeLimit, PRIV_USER | SEG_NOEXEC | SEG_W | SEG_NORMAL, SEG_GRAN_4K_PAGE)

    defaultUserSegments.cs = uint32(userModeCS*8)
    defaultUserSegments.ss = uint32(userModeDS*8)
    defaultUserSegments.ds = uint32(userModeDS*8)
    defaultUserSegments.es = uint32(userModeDS*8)
    defaultUserSegments.fs = 0
    defaultUserSegments.gs = 0

    tssIndex := AddSegment(uintptr(unsafe.Pointer(&tss)), unsafe.Sizeof(tss), PRIV_KERNEL | SEG_NORW | TSS_32MODE | SEG_SYSTEM | TSS_IS_TSS, SEG_GRAN_BYTE)


    UpdateGdt()
    tss.ss0 = KDS_SELECTOR
    tss.esp0 = uint32(kernelStackStart)
    flushTss(tssIndex*8)
}

