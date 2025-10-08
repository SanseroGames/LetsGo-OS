package kernel

import (
	"path"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/sanserogames/letsgo-os/kernel/log"
	"github.com/sanserogames/letsgo-os/kernel/mm"
)

// TODO: Move somewhere else?
const ESUCCESS = syscall.Errno(0)

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
	pcHeader    unsafe.Pointer
	funcnametab []byte
	cutab       []uint32
	filetab     []byte
	pctab       []byte
	pclntable   []byte
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

//go:linkname runtimeFindFunc runtime.findfunc
func runtimeFindFunc(pc uintptr) funcInfo

type tssEntry struct {
	prev_tss uint32 // The previous TSS - with hardware task switching these form a kind of backward linked list.
	esp0     uint32 // The stack pointer to load when changing to kernel mode.
	ss0      uint32 // The stack segment to load when changing to kernel mode.
	// Everything below here is unused.
	esp1       uint32 // esp and ss 1 and 2 would be used when switching to rings 1 or 2.
	ss1        uint32
	esp2       uint32
	ss2        uint32
	cr3        uint32
	eip        uint32
	eflags     uint32
	eax        uint32
	ecx        uint32
	edx        uint32
	ebx        uint32
	esp        uint32
	ebp        uint32
	esi        uint32
	edi        uint32
	es         uint32
	cs         uint32
	ss         uint32
	ds         uint32
	fs         uint32
	gs         uint32
	ldt        uint32
	trap       uint16
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
	userModeBase      = 0x00000000
	userModeLimit     = 0xFFFFFF
	TSS_IS_TSS        = 1 << 0
	TSS_32MODE        = 1 << 3
	defaultStackPages = 16
	defaultStackStart = uintptr(0xffffc000)

	EFLAGS_IF = 0x200 // interrupt flag. enable interrupts
	EFLAGS_R  = 0x2   // Reserved. always 1
)

var (
	tss                 tssEntry
	defaultUserSegments SegmentList
)

// Don't forget to multiply by 8. it is not array index.
func flushTss(segmentIndex int)

func printFuncName(pc uintptr) {
	f := runtimeFindFunc(pc)
	if !f.valid() {
		log.KPrintLn("func: ", pc)
		return
	}
	s := f._Func().Name()
	file, line := f._Func().FileLine(pc)
	_, filename := path.Split(file)
	log.KPrintLn(s, " (", filename, ":", line, ")")
}

func KernelThreadInit() {
	SetInterruptStack(CurrentThread.kernelStack.hi)
	switchPageDir(CurrentThread.Domain.MemorySpace.PageDirectory)
	JumpUserMode(CurrentThread.Regs, CurrentThread.info)
}

func JumpUserMode(regs RegisterState, info InterruptInfo)

func hackyGetFuncAddr(funcAddr func()) uintptr

func SetInterruptStack(addr uintptr) {
	tss.esp0 = uint32(addr)
}

func CreateNewThread(outThread *Thread, newStack uintptr, cloneThread *Thread, targetDomain *Domain) {
	newThreadAddr := (uintptr)(unsafe.Pointer(outThread))
	mm.Memclr(newThreadAddr, int(unsafe.Sizeof(outThread)))
	outThread.fpOffset = 0xffffffff
	kernelStack := mm.AllocPage()
	if ENABLE_DEBUG {
		log.KDebugLn("thread stack ", kernelStack)
	}

	kernelStack.Clear()
	outThread.kernelStack.lo = kernelStack.Address()
	outThread.kernelStack.hi = kernelStack.Address() + PAGE_SIZE
	outThread.kernelInfo.ESP = uint32(outThread.kernelStack.hi)
	outThread.kernelInfo.EIP = hackyGetFuncAddr(KernelThreadInit)
	outThread.kernelInfo.CS = KCS_SELECTOR
	outThread.kernelInfo.SS = KDS_SELECTOR
	outThread.KernelRegs.GS = KGS_SELECTOR
	outThread.kernelInfo.EFLAGS = EFLAGS_R
	targetDomain.MemorySpace.MapPage(kernelStack.Address(), kernelStack.Address(), PAGE_RW|PAGE_PERM_KERNEL)

	outThread.info.CS = defaultUserSegments.cs | 3
	outThread.info.SS = defaultUserSegments.ss | 3
	outThread.Regs.GS = defaultUserSegments.gs | 3
	outThread.Regs.FS = defaultUserSegments.fs | 3
	outThread.Regs.ES = defaultUserSegments.es | 3
	outThread.Regs.DS = defaultUserSegments.ds | 3
	outThread.info.EFLAGS = EFLAGS_R | EFLAGS_IF
	if newStack != 0 {
		outThread.userStack.hi = newStack
		outThread.userStack.lo = newStack - 16*0x1000
		outThread.info.ESP = uint32(newStack)
	} else {
		outThread.userStack.hi = 0
		outThread.userStack.lo = 0
		outThread.info.ESP = 0
	}

	if cloneThread != nil {
		outThread.info.CS = cloneThread.info.CS
		outThread.info.SS = cloneThread.info.SS
		outThread.info.EIP = cloneThread.info.EIP
		outThread.info.EFLAGS = cloneThread.info.EFLAGS
		outThread.Regs = cloneThread.Regs
		outThread.Regs.EAX = 0
		if newStack == 0 {
			outThread.userStack.hi = cloneThread.userStack.hi
			outThread.userStack.lo = cloneThread.userStack.lo
			outThread.info.ESP = cloneThread.info.ESP
		}
		//outThread.fpState = cloneThread.fpState
		copy(outThread.tlsSegments[TLS_START:], cloneThread.tlsSegments[TLS_START:])
	}
	if outThread.Next != nil || outThread.prev != nil {
		kernelPanic("thread should not be in a list yet")
	}
	targetDomain.AddThread(outThread)
}

// Need pointer as this function should not do any memory allocations
func StartProgram(path string, outDomain *Domain, outMainThread *Thread) syscall.Errno {
	module, err := FindMultibootModule(path)
	if err != ESUCCESS {
		log.KErrorLn("Could not load elf file")
		return err
	}
	return startProgramInternal(module, 0, 0, outDomain, outMainThread)
}

func StartProgramUsr(path uintptr, argv uintptr, envp uintptr, outDomain *Domain, outMainThread *Thread) syscall.Errno {
	module, err := FindMultibootModuleUsr(path)
	if err != ESUCCESS {
		log.KErrorLn("Could not load elf file")
		return err
	}
	return startProgramInternal(module, argv, envp, outDomain, outMainThread)
}

func startProgramInternal(module *MultibootModule, argv uintptr, envp uintptr, outDomain *Domain, outMainThread *Thread) syscall.Errno {
	if outDomain == nil || outMainThread == nil {
		log.KErrorLn("Cannot start program. Please allocate the memory for me")
		return syscall.ENOMEM
	}
	outDomain.Segments = defaultUserSegments
	outDomain.MemorySpace = CreateNewPageDirectory()

	elfHdr, loadAddr, topAddr, err := LoadElfFile(module, &outDomain.MemorySpace)

	if err != ESUCCESS {
		outDomain.MemorySpace.FreeAllPages()
		// Assumption: LoadElfFile cannot fail if it started allocating pages
		log.KErrorLn("Could not load elf file")
		return err
	}
	outDomain.ProgramName = module.Cmdline()
	outDomain.MemorySpace.Brk = topAddr

	var stackPages [defaultStackPages]mm.Page
	for i := 0; i < defaultStackPages; i++ {
		stack := mm.AllocPage()
		stack.Clear()
		outDomain.MemorySpace.MapPage(stack.Address(), defaultStackStart-uintptr((i+1)*PAGE_SIZE), PAGE_RW|PAGE_PERM_USER)
		stackPages[i] = stack
	}

	CreateNewThread(outMainThread, defaultStackStart, nil, outDomain)

	argcOffset := 0
	argpOffset := argcOffset + 4

	// For now restrict process env to single page
	argPage := stackPages[0]
	argPageUintPtr := unsafe.Slice((*uintptr)(argPage.Pointer()), (PAGE_SIZE)/4)
	argOffset := PAGE_SIZE
	pointerPage := argPageUintPtr[argpOffset/4:]
	pointerOffset := 0
	argCount := 0
	if argv != 0 {
		for argPointer, err := range mm.IterateUserSpaceType[uintptr](argv, &CurrentDomain.MemorySpace) {
			if err != ESUCCESS {
				outDomain.MemorySpace.FreeAllPages()
				return err
			}
			if argPointer == 0 {
				break
			}
			argLength := 0
			for value, err := range CurrentDomain.MemorySpace.IterateUserSpace(argPointer) {
				if err != ESUCCESS {
					outDomain.MemorySpace.FreeAllPages()
					return err
				}
				argLength++
				if value == 0 {
					break
				}
			}
			argSlice := argPage[argOffset-argLength : argOffset]
			CurrentDomain.MemorySpace.ReadBytesFromUserSpace(argPointer, argSlice)
			argOffset = argOffset - argLength

			// TODO: Check that it wont overlap
			argCount++
			pointerPage[pointerOffset] = defaultStackStart - PAGE_SIZE + uintptr(argOffset)
			pointerOffset++
		}
	}
	argPageUintPtr[argcOffset/4] = uintptr(argCount)

	envCount := 0
	envpOffset := argpOffset + (argCount+1)*4
	pointerPage = argPageUintPtr[envpOffset/4:]
	pointerOffset = 0
	if envp != 0 {
		for envPointer, err := range mm.IterateUserSpaceType[uintptr](envp, &CurrentDomain.MemorySpace) {
			if err != ESUCCESS {
				outDomain.MemorySpace.FreeAllPages()
				return err
			}
			if envPointer == 0 {
				break
			}
			envLength := 0
			for value, err := range CurrentDomain.MemorySpace.IterateUserSpace(envPointer) {
				if err != ESUCCESS {
					outDomain.MemorySpace.FreeAllPages()
					return err
				}
				envLength++
				if value == 0 {
					break
				}
			}
			envSlice := argPage[argOffset-envLength : argOffset]
			CurrentDomain.MemorySpace.ReadBytesFromUserSpace(envPointer, envSlice)
			argOffset = argOffset - envLength

			// TODO: Check that it wont overlap
			envCount++
			pointerPage[pointerOffset] = defaultStackStart - PAGE_SIZE + uintptr(argOffset)
			pointerOffset++
		}
	}

	var aux [32]auxVecEntry
	nrVec := LoadAuxVector(aux[:], elfHdr, loadAddr)
	nrVec += nrVec % 2
	vecByteSize := nrVec * int(unsafe.Sizeof(aux[0]))

	auxOffset := envpOffset + (envCount+1)*4

	auxVectorBytes := argPage[auxOffset : auxOffset+vecByteSize]
	auxVector := unsafe.Slice((*auxVecEntry)(unsafe.Pointer(unsafe.SliceData(auxVectorBytes))), len(auxVectorBytes)/int(unsafe.Sizeof(aux[0])))

	copy(auxVector, aux[:nrVec])

	outMainThread.info.EIP = uintptr(elfHdr.Entry)
	outMainThread.info.ESP = uint32(defaultStackStart) - PAGE_SIZE
	return ESUCCESS
}

func InitUserMode(kernelStackStart uintptr, kernelStackEnd uintptr) { // Add usermode code segment
	userModeCS := AddSegment(userModeBase, userModeLimit, PRIV_USER|SEG_EXEC|SEG_R|SEG_NORMAL, SEG_GRAN_4K_PAGE)
	// Add usermode data segment
	userModeDS := AddSegment(userModeBase, userModeLimit, PRIV_USER|SEG_NOEXEC|SEG_W|SEG_NORMAL, SEG_GRAN_4K_PAGE)

	defaultUserSegments.cs = uint32(userModeCS * 8)
	defaultUserSegments.ss = uint32(userModeDS * 8)
	defaultUserSegments.ds = uint32(userModeDS * 8)
	defaultUserSegments.es = uint32(userModeDS * 8)
	defaultUserSegments.fs = 0
	defaultUserSegments.gs = 0

	tssIndex := AddSegment(uintptr(unsafe.Pointer(&tss)), unsafe.Sizeof(tss), PRIV_KERNEL|SEG_NORW|TSS_32MODE|SEG_SYSTEM|TSS_IS_TSS, SEG_GRAN_BYTE)

	tss.ss0 = KDS_SELECTOR
	tss.esp0 = uint32(kernelStackStart)
	scheduleThread.kernelStack.hi = kernelStackStart
	scheduleThread.kernelStack.lo = kernelStackEnd
	flushTss(tssIndex * 8)
}
