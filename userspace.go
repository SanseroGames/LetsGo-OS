package main

import (
    "unsafe"
    "runtime"
)

type _func struct {
	entry   uintptr // start pc
	nameoff int32   // function name

	args        int32  // in/out args size
	deferreturn uint32 // offset of start of a deferreturn call instruction from entry, if any.

	pcsp      uint32
	pcfile    uint32
	pcln      uint32
	npcdata   uint32
	cuOffset  uint32 // runtime.cutab offset of this function's CU
	funcID    uint64 // set for certain special runtime functions
	flag      uint8
	_         [1]byte // pad
	nfuncdata uint8   // must be last, must end on a uint32-aligned boundary
}

type moduledata struct {
	pcHeader     unsafe.Pointer
    funcnametab  []byte
	cutab        []uint32
	filetab      []byte
	pctab        []byte
	pclntable    []byte
}

type funcInfo struct {
	*_func
	datap *moduledata
}

func (f funcInfo) _Func() *runtime.Func {
	return (*runtime.Func)(unsafe.Pointer(f._func))
}

func (f funcInfo) valid() bool {
	return f._func != nil
}

//go:linkname findfuncTest runtime.findfunc
func findfuncTest(pc uintptr) funcInfo

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

func kernelThreadInit() {
    SetInterruptStack(currentThread.kernelStack.hi)
    switchPageDir(currentThread.domain.MemorySpace.PageDirectory)
    JumpUserMode(currentThread.regs, currentThread.info)
}

func JumpUserMode (regs RegisterState, info InterruptInfo)

func hackyGetFuncAddr(funcAddr func()) uintptr

func SetInterruptStack(addr uintptr) {
    tss.esp0 = uint32(addr)
}

func CreateNewThread(outThread *thread, newStack uintptr, cloneThread *thread, targetDomain *domain) {
    newThreadAddr := (uintptr)(unsafe.Pointer(outThread))
    Memclr(newThreadAddr, int(unsafe.Sizeof(outThread)))
    outThread.fpOffset = 0xffffffff
    outThread.userStack.hi = newStack
    outThread.userStack.lo = newStack-16*0x1000
    kernelStack := allocPage()
    Memclr(kernelStack, PAGE_SIZE)
    outThread.kernelStack.lo = kernelStack
    outThread.kernelStack.hi = kernelStack+PAGE_SIZE
    outThread.kernelInfo.ESP = uint32(outThread.kernelStack.hi)
    targetDomain.MemorySpace.mapPage(kernelStack, kernelStack, PAGE_RW | PAGE_PERM_KERNEL)
    outThread.kernelInfo.EIP = uint32(hackyGetFuncAddr(kernelThreadInit))
    outThread.info.CS = defaultUserSegments.cs | 3
    outThread.info.SS = defaultUserSegments.ss | 3
    outThread.regs.GS = defaultUserSegments.gs | 3
    outThread.regs.FS = defaultUserSegments.fs | 3
    outThread.regs.ES = defaultUserSegments.es | 3
    outThread.regs.DS = defaultUserSegments.ds | 3
    outThread.info.EFLAGS = 0x202
    if cloneThread != nil {
        outThread.info.CS = cloneThread.info.CS
        outThread.info.SS = cloneThread.info.SS
        outThread.info.EIP = cloneThread.info.EIP
        outThread.info.EFLAGS = cloneThread.info.EFLAGS
        outThread.regs = cloneThread.regs
        outThread.regs.EAX = 0
        //outThread.fpState = cloneThread.fpState
        copy(outThread.tlsSegments[TLS_START:], cloneThread.tlsSegments[TLS_START:])
    }
    if outThread.next != nil || outThread.prev != nil {
        kernelPanic("thread should not be in a list yet")
    }
    outThread.info.ESP = uint32(newStack)
    targetDomain.AddThread(outThread)
}
// Need pointer as this function should not do any memory allocations
func StartProgram(path string, outDomain *domain, outMainThread *thread) int {
    if outDomain == nil || outMainThread == nil {
        text_mode_print_errorln("Cannot start program. Plase allocate the memory for me")
        return 1
    }
    outDomain.Segments = defaultUserSegments
    outDomain.MemorySpace = createNewPageDirectory()

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

    CreateNewThread(outMainThread, defaultStackStart, nil, outDomain)

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


func InitUserMode(kernelStackStart uintptr, kernelStackEnd uintptr) {             // Add usermode code segment
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


    tss.ss0 = KDS_SELECTOR
    tss.esp0 = uint32(kernelStackStart)
    scheduleThread.kernelStack.hi = kernelStackStart
    scheduleThread.kernelStack.lo = kernelStackEnd
    flushTss(tssIndex*8)
}

