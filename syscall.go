package main

import (
	"syscall"
	"unsafe"
)

const PRINT_SYSCALL = ENABLE_DEBUG

type ioVec struct {
	iovBase uintptr /* Starting address */
	iovLen  uint32  /* Number of bytes to transfer */
}

const _UTSNAME_LENGTH = 65

type utsname struct {
	/* Name of the implementation of the operating system.  */
	sysname [_UTSNAME_LENGTH]byte
	/* Name of this node on the network.  */
	nodename [_UTSNAME_LENGTH]byte
	/* Current release level of this implementation.  */
	release [_UTSNAME_LENGTH]byte
	/* Current version level of this release.  */
	version [_UTSNAME_LENGTH]byte
	/* Name of the hardware type the system is running on.  */
	machine [_UTSNAME_LENGTH]byte

	domainname [_UTSNAME_LENGTH]byte
}

type statxDataTimestamp struct {
	tv_sec   int64  /* Seconds since the Epoch (UNIX time) */
	tv_nsec  uint32 /* Nanoseconds since tv_sec */
	reserved uint32
}

type statxData struct {
	stx_mask uint32 /* Mask of bits indicating
	   filled fields */
	stx_blksize         uint32 /* Block size for filesystem I/O */
	stx_attributes      uint64 /* Extra file attribute indicators */
	stx_nlink           uint32 /* Number of hard links */
	stx_uid             uint32 /* User ID of owner */
	stx_gid             uint32 /* Group ID of owner */
	stx_mode            uint16 /* File type and mode */
	reserved            uint16
	stx_ino             uint64 /* Inode number */
	stx_size            uint64 /* Total size in bytes */
	stx_blocks          uint64 /* Number of 512B blocks allocated */
	stx_attributes_mask uint64
	/* Mask to show what's supported
	   in stx_attributes */

	/* The following fields are file timestamps */
	stx_atime statxDataTimestamp /* Last access */
	stx_btime statxDataTimestamp /* Creation */
	stx_ctime statxDataTimestamp /* Last status change */
	stx_mtime statxDataTimestamp /* Last modification */

	/* If this file represents a device, then the next two
	   fields contain the ID of the device */
	stx_rdev_major uint32 /* Major ID */
	stx_rdev_minor uint32 /* Minor ID */

	/* The next two fields contain the ID of the device
	   containing the filesystem where the file resides */
	stx_dev_major uint32 /* Major ID */
	stx_dev_minor uint32 /* Minor ID */
}

type syscallHandler func(syscallArgs) (uint32, syscall.Errno)
type syscallEntry struct {
	handler      syscallHandler
	name         string
	numberOfArgs byte
}

type syscallArgs struct {
	arg1 uint32
	arg2 uint32
	arg3 uint32
	arg4 uint32
	arg5 uint32
	arg6 uint32
}

const (
	_CLONE_VM             = 0x100
	_CLONE_FS             = 0x200
	_CLONE_FILES          = 0x400
	_CLONE_SIGHAND        = 0x800
	_CLONE_PTRACE         = 0x2000
	_CLONE_VFORK          = 0x4000
	_CLONE_PARENT         = 0x8000
	_CLONE_THREAD         = 0x10000
	_CLONE_NEWNS          = 0x20000
	_CLONE_SYSVSEM        = 0x40000
	_CLONE_SETTLS         = 0x80000
	_CLONE_PARENT_SETTID  = 0x100000
	_CLONE_CHILD_CLEARTID = 0x200000
	_CLONE_UNTRACED       = 0x800000
	_CLONE_CHILD_SETTID   = 0x1000000
	_CLONE_STOPPED        = 0x2000000
	_CLONE_NEWUTS         = 0x4000000
	_CLONE_NEWIPC         = 0x8000000

	__S_IFDIR = 0040000 /* Directory.  */
	__S_IFCHR = 0020000 /* Character device.  */
	__S_IFBLK = 0060000 /* Block device.  */
	__S_IFREG = 0100000 /* Regular file.  */
	__S_IFIFO = 0010000 /* FIFO.  */

	STATX_TYPE        = 0x00000001 /* Want/got stx_mode & S_IFMT */
	STATX_MODE        = 0x00000002 /* Want/got stx_mode & ~S_IFMT */
	STATX_NLINK       = 0x00000004 /* Want/got stx_nlink */
	STATX_UID         = 0x00000008 /* Want/got stx_uid */
	STATX_GID         = 0x00000010 /* Want/got stx_gid */
	STATX_ATIME       = 0x00000020 /* Want/got stx_atime */
	STATX_MTIME       = 0x00000040 /* Want/got stx_mtime */
	STATX_CTIME       = 0x00000080 /* Want/got stx_ctime */
	STATX_INO         = 0x00000100 /* Want/got stx_ino */
	STATX_SIZE        = 0x00000200 /* Want/got stx_size */
	STATX_BLOCKS      = 0x00000400 /* Want/got stx_blocks */
	STATX_BASIC_STATS = 0x000007ff /* The stuff in the normal stat struct */
	STATX_BTIME       = 0x00000800 /* Want/got stx_btime */
	STATX_MNT_ID      = 0x00001000 /* Got stx_mnt_id */

	FUTEX_WAIT            = 0
	FUTEX_WAKE            = 1
	FUTEX_FD              = 2
	FUTEX_REQUEUE         = 3
	FUTEX_CMP_REQUEUE     = 4
	FUTEX_WAKE_OP         = 5
	FUTEX_LOCK_PI         = 6
	FUTEX_UNLOCK_PI       = 7
	FUTEX_TRYLOCK_PI      = 8
	FUTEX_WAIT_BITSET     = 9
	FUTEX_WAKE_BITSET     = 10
	FUTEX_WAIT_REQUEUE_PI = 11
	FUTEX_CMP_REQUEUE_PI  = 12

	FUTEX_PRIVATE_FLAG = 128

	ESUCCESS = syscall.Errno(0)

	MMAP_PROT_READ  = 1
	MMAP_PROT_WRITE = 2
	MMAP_PROT_EXEC  = 4

	MMAP_MAP_FIXED     = 0x10
	MMAP_MAP_ANONYMOUS = 0x20

	REBOOT_MAGIC1       = 0xfee1dead
	REBOOT_MAGIC2       = 0x28121969
	REBOOT_CMD_POWEROFF = 0x4321fedc
)

var (
	uts = utsname{
		sysname:    [_UTSNAME_LENGTH]byte{'L', 'e', 't', '\'', 's', 'G', 'o', '!', ' ', 'O', 'S', 0},
		nodename:   [_UTSNAME_LENGTH]byte{0},
		release:    [_UTSNAME_LENGTH]byte{'6', '9', '.', '4', '.', '2', '0', 0},
		version:    [_UTSNAME_LENGTH]byte{'0', '.', '1', 0},
		machine:    [_UTSNAME_LENGTH]byte{0},
		domainname: [_UTSNAME_LENGTH]byte{0},
	}

	// As I don't implement the full set of linux syscalls I try to save memory by using a lookup in a pointer table
	// and not sort them in the list.
	registeredSyscalls = [0x200](byte){}
	syscallListRaw     = [64]syscallEntry{}
	syscallList        = []syscallEntry{}
	okHandler          = func(args syscallArgs) (uint32, syscall.Errno) { return 0, ESUCCESS }
	invalHandler       = func(args syscallArgs) (uint32, syscall.Errno) { return 0, syscall.EINVAL }
)

// Not quite happy with the args here. Annoying that you have to redeclare variables
// in the implementation if you want meaningful names. Also you can't have variables
// that are for docu or implemented later due to go not liking unused variables.
func syscalltest(args syscallArgs) (uint32, syscall.Errno) {
	return 0, syscall.EFAULT
}

func RegisterSyscall(number int, name string, handler syscallHandler) {
	if len(syscallList) == cap(syscallList) {
		kernelPanic("Too many syscalls registered")
	}
	syscallList = append(syscallList, syscallEntry{handler: handler, name: name, numberOfArgs: 0})
	registeredSyscalls[number] = byte(len(syscallList)) // add 1 to distinguish uninitialized
}

func InitSyscall() {
	SetInterruptHandler(0x80, linuxSyscallHandler, KCS_SELECTOR, PRIV_USER)

	// Somehow this does not work when written in the variables above
	syscallList = syscallListRaw[:0]

	RegisterSyscall(syscall.SYS_WRITE, "write syscall", linuxWriteSyscall)
	RegisterSyscall(syscall.SYS_WRITEV, "writeV syscall", linuxWriteVSyscall)
	RegisterSyscall(syscall.SYS_SET_THREAD_AREA, "set thread area syscall", linuxSetThreadAreaSyscall)
	RegisterSyscall(syscall.SYS_OPEN, "open syscall", linuxOpenSyscall)
	RegisterSyscall(syscall.SYS_OPENAT, "open at syscall", linuxOpenAtSyscall)
	RegisterSyscall(syscall.SYS_CLOSE, "close syscall", okHandler)
	RegisterSyscall(syscall.SYS_READ, "read syscall", linuxReadSyscall)
	RegisterSyscall(syscall.SYS_READLINK, "readlink syscall", okHandler)
	RegisterSyscall(syscall.SYS_READLINKAT, "read link at syscall", invalHandler)
	RegisterSyscall(syscall.SYS_SCHED_GETAFFINITY, "sched get affinity syscall", func(args syscallArgs) (uint32, syscall.Errno) { return 0xffffffff, ESUCCESS })
	RegisterSyscall(syscall.SYS_NANOSLEEP, "nano sleep syscall", invalHandler)
	RegisterSyscall(syscall.SYS_EXIT_GROUP, "exit group syscall", linuxExitGroupSyscall)
	RegisterSyscall(syscall.SYS_EXIT, "exit syscall", linuxExitSyscall)
	RegisterSyscall(syscall.SYS_BRK, "brk syscall", linuxBrkSyscall)
	RegisterSyscall(syscall.SYS_MMAP2, "mmap2 syscall", linuxMmap2Syscall)
	RegisterSyscall(syscall.SYS_MINCORE, "mincore syscall", linuxMincoreSyscall)
	RegisterSyscall(syscall.SYS_MUNMAP, "munmap syscall", linuxMunmapSyscall)
	RegisterSyscall(syscall.SYS_CLOCK_GETTIME, "clock get time syscall", func(args syscallArgs) (uint32, syscall.Errno) { return 0, syscall.ENOTSUP })
	RegisterSyscall(syscall.SYS_RT_SIGPROCMASK, "sig proc mask syscall", okHandler)
	RegisterSyscall(syscall.SYS_SIGALTSTACK, "sig alt stack syscall", okHandler)
	RegisterSyscall(syscall.SYS_RT_SIGACTION, "rt sig action syscall", okHandler)
	RegisterSyscall(syscall.SYS_GETTID, "gettid syscall", getTidSyscall)
	RegisterSyscall(syscall.SYS_GETPID, "get pid syscall", getPidSyscall)
	RegisterSyscall(syscall.SYS_SET_TID_ADDRESS, "set tid address syscall", okHandler)
	RegisterSyscall(syscall.SYS_POLL, "poll syscall", okHandler)
	RegisterSyscall(syscall.SYS_CLONE, "clone syscall", linuxCloneSyscall)
	RegisterSyscall(syscall.SYS_FUTEX, "futex syscall", linuxFutexSyscall)
	RegisterSyscall(syscall.SYS_SCHED_YIELD, "sched yield syscall", linuxSchedYieldSyscall)
	RegisterSyscall(syscall.SYS_GETEUID32, "get euid syscall", okHandler)
	RegisterSyscall(syscall.SYS_GETUID32, "get uid syscall", okHandler)
	RegisterSyscall(syscall.SYS_GETEGID32, "get egid syscall", okHandler)
	RegisterSyscall(syscall.SYS_GETGID32, "get gid syscall", okHandler)
	RegisterSyscall(syscall.SYS_UNAME, "uname syscall", linuxUnameSyscall)
	RegisterSyscall(syscall.SYS_TGKILL, "tgkill syscall", okHandler)
	RegisterSyscall(syscall.SYS_MPROTECT, "mprotect syscall", okHandler)
	RegisterSyscall(syscall.SYS_SET_ROBUST_LIST, "set robust list sycall", invalHandler)
	RegisterSyscall(syscall.SYS_UGETRLIMIT, "get upper limit syscall", invalHandler)
	RegisterSyscall(0x163, "get random syscall", invalHandler)
	RegisterSyscall(0x17f, "statx syscall", linuxStatxSyscall)
	RegisterSyscall(0x180, "arch ptrctl syscall", invalHandler)
	RegisterSyscall(0x182, "rseq syscall", invalHandler)
	RegisterSyscall(0x193, "clock gettime 64 syscall", invalHandler)
	RegisterSyscall(syscall.SYS_EPOLL_CREATE1, "epoll_create1 syscall", invalHandler)
	RegisterSyscall(syscall.SYS_EPOLL_WAIT, "epoll wait syscall", okHandler)
	RegisterSyscall(syscall.SYS_EPOLL_CREATE, "epoll_create syscall", linuxEpollCreateSyscall)
	RegisterSyscall(syscall.SYS_FCNTL64, "fcntl64 syscall", okHandler)
	RegisterSyscall(syscall.SYS_FCNTL, "fctnl syscall", okHandler)
	RegisterSyscall(syscall.SYS_PRCTL, "prctl syscall", invalHandler)
	RegisterSyscall(syscall.SYS_PIPE2, "pipe2 syscall", okHandler)
	RegisterSyscall(syscall.SYS_EPOLL_CTL, "epoll_ctl syscall", okHandler)
	RegisterSyscall(syscall.SYS_DUP3, "dup3 syscall", okHandler)
	RegisterSyscall(syscall.SYS_DUP2, "dup2 syscall", okHandler)
	RegisterSyscall(syscall.SYS_EXECVE, "execve syscall", linuxExecveSyscall)
	RegisterSyscall(syscall.SYS_MADVISE, "madvise syscall", okHandler)
	RegisterSyscall(syscall.SYS_PRLIMIT64, "prlimit64 syscall", okHandler)
	RegisterSyscall(syscall.SYS_REBOOT, "reboot syscall", rebootHandler)
	RegisterSyscall(syscall.SYS_WAIT4, "wait4 syscall", linuxWaitPidSyscall)
	RegisterSyscall(syscall.SYS_FSTATAT64, "fstatat64 syscall", okHandler)
	RegisterSyscall(syscall.SYS_GETCWD, "fstatat64 syscall", okHandler)
}

func getTidSyscall(args syscallArgs) (uint32, syscall.Errno) {
	return currentThread.tid, ESUCCESS
}

func getPidSyscall(args syscallArgs) (uint32, syscall.Errno) {
	return currentThread.domain.pid, ESUCCESS
}

func linuxSyscallHandler() {
	var ret uint32 = 0
	var err syscall.Errno = ESUCCESS
	syscallNr := currentThread.regs.EAX
	args := syscallArgs{
		arg1: currentThread.regs.EBX,
		arg2: currentThread.regs.ECX,
		arg3: currentThread.regs.EDX,
		arg4: currentThread.regs.ESI,
		arg5: currentThread.regs.EDI,
		arg6: currentThread.regs.EBP,
	}
	if currentThread.isKernelInterrupt {
		syscallNr = currentThread.kernelRegs.EAX
		args = syscallArgs{
			arg1: currentThread.kernelRegs.EBX,
			arg2: currentThread.kernelRegs.ECX,
			arg3: currentThread.kernelRegs.EDX,
			arg4: currentThread.kernelRegs.ESI,
			arg5: currentThread.kernelRegs.EDI,
			arg6: currentThread.kernelRegs.EBP,
		}
		if syscallNr == syscall.SYS_WRITE {
			// Write syscall. We probably upset the go runtime so it wants to complain
			linuxWriteSyscall(args)
		}
		kprintln("\nkernel syscallnr: ", syscallNr)
		kernelPanic("Why is the kernel making a syscall?")
	}
	handlerIdx := registeredSyscalls[syscallNr]
	if handlerIdx == 0 {
		unsupportedSyscall()
		return
	}
	handler := syscallList[handlerIdx-1]
	if PRINT_SYSCALL {
		kdebugln("pid: ",
			currentThread.domain.pid,
			" tid: ",
			currentThread.tid,
			" :: ",
			handler.name,
			" (",
			syscallNr,
			" hex(",
			uintptr(syscallNr),
			")",
			")")
		kdebugln(" |- arg1:",
			uintptr(args.arg1),
			", arg2:",
			uintptr(args.arg2),
			", arg3:",
			uintptr(args.arg3),
			", arg4:",
			uintptr(args.arg4),
			", arg5:",
			uintptr(args.arg5),
			", arg6:",
			uintptr(args.arg6),
		)
	}
	ret, err = handler.handler(args)
	if PRINT_SYSCALL {
		kdebugln("SYSCALL RETURN pid: ",
			currentThread.domain.pid,
			" tid: ",
			currentThread.tid,
			" :: ",
			ret,
			" (",
			uint32(err),
			")",
		)
	}
	if err != ESUCCESS {
		ret = ^uint32(err) + 1
	}
	currentThread.regs.EAX = ret
}

func linuxExecveSyscall(args syscallArgs) (uint32, syscall.Errno) {
	arr := args.arg1
	addr, ok := currentThread.domain.MemorySpace.GetPhysicalAddress(uintptr(arr))
	if !ok {
		kerrorln("Could not look up string pathname")
		return 0, syscall.EFAULT
	}
	pathname := cstring(addr)

	// Create new domain
	newDomainMem := AllocPage()
	Memclr(newDomainMem, PAGE_SIZE)
	newDomain := (*domain)(unsafe.Pointer(newDomainMem))
	newThreadMem := AllocPage()
	Memclr(newThreadMem, PAGE_SIZE)
	newThread := (*thread)(unsafe.Pointer(newThreadMem))
	err := StartProgram(pathname, newDomain, newThread)
	if err != 0 {
		FreePage(newDomainMem)
		FreePage(newThreadMem)
		return 0, syscall.ENOENT
	}
	newDomain.MemorySpace.MapPage(newThreadMem, newThreadMem, PAGE_RW|PAGE_PERM_KERNEL)
	AddDomain(newDomain)

	if !currentThread.isFork {
		// We were not in a fork, so the process should be replaced
		// We simulate this by just exiting the current process
		//kernelPanic("Execve in non fork thread")
		//ExitDomain(currentThread.domain)
	} else {
		ExitThread(currentThread)
	}
	PerformSchedule = true
	// return currentThread.regs.EAX, ESUCCESS
	return newDomain.pid, ESUCCESS
}

func linuxWaitPidSyscall(args syscallArgs) (uint32, syscall.Errno) {
	waitPid := args.arg1
	// status := args.arg2
	// options := args.arg3
	// usage := args.arg3
	// kdebugln("Wait for ", waitPid)

	for {
		Block()
		notStarted := true
		stillWaiting := false
		for cur := allDomains.head; notStarted || cur != allDomains.head; cur = cur.next {
			if cur.pid == waitPid {
				stillWaiting = true
				break
			}
			notStarted = false
		}
		if !stillWaiting {
			break
		}
	}

	return 0, ESUCCESS
}

func linuxEpollCreateSyscall(args syscallArgs) (uint32, syscall.Errno) {
	// TODO
	return 0, ESUCCESS
}

func linuxExitGroupSyscall(args syscallArgs) (uint32, syscall.Errno) {
	ExitDomain(currentThread.domain)
	PerformSchedule = true
	// Already in new context so return value from last syscall from current domain
	return currentThread.regs.EAX, ESUCCESS
}

func linuxExitSyscall(args syscallArgs) (uint32, syscall.Errno) {
	ExitThread(currentThread)
	PerformSchedule = true
	// Already in new context so return value from last syscall from current domain
	return currentThread.regs.EAX, ESUCCESS
}

func rebootHandler(args syscallArgs) (uint32, syscall.Errno) {
	magic1 := args.arg1
	magic2 := args.arg2
	cmd := args.arg3
	if magic1 != REBOOT_MAGIC1 && magic2 != REBOOT_MAGIC2 {
		return 0, syscall.EINVAL
	}
	if cmd != REBOOT_CMD_POWEROFF {
		kerrorln("invalid reboot command")
		return 0, syscall.EINVAL
	}
	Shutdown()
	return 0, syscall.EINVAL
}

func linuxStatxSyscall(args syscallArgs) (uint32, syscall.Errno) {
	//dirfd := args.arg1
	//path := args.arg2
	//flags := args.arg3
	//mask := args.arg4
	buf := args.arg5
	addr, ok := currentThread.domain.MemorySpace.GetPhysicalAddress(uintptr(buf))
	if !ok {
		kerrorln("invalid adress in statx")
		return 0, syscall.EFAULT
	}
	item := (*statxData)(unsafe.Pointer(addr))
	Memclr(addr, int(unsafe.Sizeof(item)))
	// Trying to replicate linux behavior
	item.stx_mask = STATX_BASIC_STATS
	item.stx_blksize = 1024
	item.stx_mode = __S_IFCHR | 0620
	item.stx_size = 0
	item.stx_blocks = 0
	item.stx_dev_major = 0
	item.stx_dev_minor = 0x18
	item.stx_ino = 0
	item.stx_nlink = 1
	item.stx_rdev_major = 136
	item.stx_rdev_minor = 0
	item.stx_uid = 0
	item.stx_gid = 0
	item.stx_attributes_mask = 0x0000000000203000
	item.stx_attributes = 0
	return 0, ESUCCESS
}

func linuxUnameSyscall(args syscallArgs) (uint32, syscall.Errno) {
	buf := args.arg1
	addr, ok := currentThread.domain.MemorySpace.GetPhysicalAddress(uintptr(buf))
	if !ok {
		kerrorln("invalid adress in uname")
		return 0, syscall.EFAULT
	}
	provided := (*utsname)(unsafe.Pointer(addr))
	*provided = uts
	return 0, ESUCCESS
}

func linuxCloneSyscall(args syscallArgs) (uint32, syscall.Errno) {
	flags := args.arg1
	stack := args.arg2
	newThreadMem := AllocPage()
	Memclr(newThreadMem, PAGE_SIZE)
	newThread := (*thread)(unsafe.Pointer(newThreadMem))
	CreateNewThread(newThread, uintptr(stack), currentThread, currentThread.domain)
	currentThread.domain.MemorySpace.MapPage(newThreadMem, newThreadMem, PAGE_RW|PAGE_PERM_KERNEL)
	if flags&_CLONE_THREAD == 0 {
		// This is probably temporary as I don't want to implement COW right now to create a new process
		kdebugln("[CLONE SYSCALL] Clone where the goal is not a thread does not behave like on linux")
		newThread.isFork = true
	}
	// Need to make this better at some point
	return newThread.tid, ESUCCESS
}

func linuxMincoreSyscall(args syscallArgs) (uint32, syscall.Errno) {
	//addr := args.arg1
	length := args.arg2
	vec := args.arg3
	vecAddr, ok := currentThread.domain.MemorySpace.GetPhysicalAddress(uintptr(vec))
	if !ok {
		kerrorln("Could not look up vec array")
		return 0, syscall.EFAULT
	}

	arr := (*[30 << 1]byte)(unsafe.Pointer(vecAddr))[:(length+PAGE_SIZE-1)/PAGE_SIZE]
	for i := range arr {
		arr[i] = 1
	}
	return 0, ESUCCESS
}

func linuxMunmapSyscall(args syscallArgs) (uint32, syscall.Errno) {
	baseAddr := args.arg1
	length := args.arg2
	//printRegisters(currentInfo, regs)
	for i := uint32(0); i < length; i += PAGE_SIZE {
		addr := uintptr(baseAddr + i)

		if addr < KERNEL_RESERVED {
			return 0, syscall.EINVAL
		}
		currentThread.domain.MemorySpace.UnmapPage(addr)
	}
	return 0, ESUCCESS
}

func linuxBrkSyscall(args syscallArgs) (uint32, syscall.Errno) {
	newBrk := uintptr(args.arg1)
	brk := currentThread.domain.MemorySpace.Brk
	if newBrk == 0 {
		return uint32(brk), ESUCCESS
	}
	if newBrk == brk || newBrk < brk {
		return uint32(brk), ESUCCESS
	}
	//text_mode_print_hex32(brk)
	for i := (brk + PAGE_SIZE - 1) &^ (PAGE_SIZE - 1); i < newBrk; i += PAGE_SIZE {
		p := AllocPage()
		Memclr(p, PAGE_SIZE)
		flags := uint8(PAGE_PERM_USER | PAGE_RW)
		// kdebugln("[brk] Map page ", i, " -> ", p)

		currentThread.domain.MemorySpace.MapPage(p, i, flags)
	}
	currentThread.domain.MemorySpace.Brk = newBrk
	// kdebugln("BRK: ", newBrk)
	return uint32(newBrk), ESUCCESS
}

func linuxMmap2Syscall(args syscallArgs) (uint32, syscall.Errno) {
	target := uintptr(args.arg1)
	size := uintptr(args.arg2)
	prot := args.arg3
	flags := args.arg4

	if flags&MMAP_MAP_ANONYMOUS == 0 {
		return 0, syscall.EINVAL
	}

	if target == 0 {
		target = currentThread.domain.MemorySpace.VmTop
	}
	if target+uintptr(size) < KERNEL_RESERVED {
		return 0, syscall.EINVAL
	}
	if prot == 0 {
		return uint32(target), ESUCCESS
	}

	startAddr := currentThread.domain.MemorySpace.FindSpaceFor(target, size)

	if flags&MMAP_MAP_FIXED == MMAP_MAP_FIXED {
		startAddr = target
	} else if startAddr == 0 {
		return 0, syscall.EINVAL
	}

	for i := startAddr; i < startAddr+size; i += PAGE_SIZE {
		p := AllocPage()
		Memclr(p, PAGE_SIZE)
		pageFlags := uint8(PAGE_PERM_USER)
		if prot&MMAP_PROT_WRITE == MMAP_PROT_WRITE {
			pageFlags |= PAGE_RW
		}

		if currentThread.domain.MemorySpace.getPageTableEntry(i).isPresent() {
			if flags&MMAP_MAP_FIXED == MMAP_MAP_FIXED {
				// TODO: Clear or free and remap page?
				// currentThread.domain.MemorySpace.unMapPage(i)
			} else {
				// TODO: What to do here?
				kernelPanic("Trying to map page which is already present without MAP_FIXED")
			}
		} else {
			currentThread.domain.MemorySpace.MapPage(p, i, pageFlags)
		}
	}

	return uint32(startAddr), ESUCCESS
}

func linuxSetThreadAreaSyscall(args syscallArgs) (uint32, syscall.Errno) {
	u_info := args.arg1
	addr, ok := currentThread.domain.MemorySpace.GetPhysicalAddress(uintptr(u_info))
	if !ok {
		kerrorln("Could not look up user desc")
		return 0, syscall.EFAULT
	}
	desc := (*UserDesc)(unsafe.Pointer(addr))
	if desc.Flags&UDESC_SEG_NOT_PRESENT != 0 {
		kerrorln("fixme: not handling updating entries")
		return 0, syscall.ENOSYS
	}

	slot := desc.EntryNumber
	if slot == 0xffffffff {
		// Find free slot
		for i := TLS_START; i < len(currentThread.tlsSegments); i++ {
			if !currentThread.tlsSegments[i].IsPresent() {
				slot = uint32(i)
				break
			}
		}
		if slot == 0xffffffff {
			// There was no free slot
			return 0, syscall.ESRCH
		}
		desc.EntryNumber = slot
	}
	SetTlsSegment(slot, desc, currentThread.tlsSegments[:])

	return 0, ESUCCESS
}

func linuxOpenSyscall(args syscallArgs) (uint32, syscall.Errno) {
	//path := args.arg1
	flags := args.arg2
	addr, ok := currentThread.domain.MemorySpace.GetPhysicalAddress(uintptr(flags))
	if !ok {
		return 0, syscall.EFAULT
	}
	s := cstring(addr)
	if PRINT_SYSCALL {
		kdebugln("[SYS-OPEN] ", s)
	}
	//text_mode_println(s)
	return 0, syscall.ENOSYS
	//printRegisters(currentInfo, regs)

}

func linuxOpenAtSyscall(args syscallArgs) (uint32, syscall.Errno) {
	fd := args.arg1
	path := args.arg2
	flags := args.arg3
	pathaddr, ok := currentThread.domain.MemorySpace.GetPhysicalAddress(uintptr(path))
	if !ok {
		return 0, syscall.EFAULT
	}
	s1 := cstring(pathaddr)
	if PRINT_SYSCALL {
		kdebugln("[SYS-OPENAT] fd:", fd)
		kdebugln("[SYS-OPENAT] path:", s1)
		kdebugln("[SYS-OPENAT] flags:", flags)
	}

	if s1 == "/dev/null" {
		return 42, ESUCCESS
	}

	//text_mode_println(s)
	return 0, syscall.ENOSYS
	//printRegisters(currentInfo, regs)

}

func linuxWriteVSyscall(args syscallArgs) (uint32, syscall.Errno) {
	fd := args.arg1
	arr := uintptr(args.arg2)
	count := args.arg3
	if fd < 1 || fd > 2 {
		return 0, syscall.EBADF
	}

	if !currentThread.domain.MemorySpace.isRangeAccessible(arr, arr+uintptr(count*8)) { // todo: sizeof?
		return 0, syscall.EFAULT
	}

	addr, ok := currentThread.domain.MemorySpace.GetPhysicalAddress(arr)
	if !ok {
		kerrorln("Could not look up string addr")
		return 0, syscall.EFAULT
	}
	iovecs := unsafe.Slice((*ioVec)(unsafe.Pointer(addr)), count)
	printed := 0
	for _, n := range iovecs {
		if !currentThread.domain.MemorySpace.isRangeAccessible(uintptr(n.iovBase), uintptr(n.iovBase)+uintptr(n.iovLen)) {
			return 0, syscall.EFAULT
		}
		addr, ok = currentThread.domain.MemorySpace.GetPhysicalAddress(uintptr(n.iovBase))
		if !ok {
			return 0, syscall.EFAULT
		}
		s := unsafe.String((*byte)(unsafe.Pointer(uintptr(addr))), n.iovLen)
		if fd == 2 {
			kerror(s)
		} else {
			kprint(s)
		}
		printed += len(s)
	}
	return uint32(printed), ESUCCESS //TODO: Return number of bytes written
}

func linuxWriteSyscall(args syscallArgs) (uint32, syscall.Errno) {
	fd := args.arg1
	text := uintptr(args.arg2)
	length := args.arg3
	if PRINT_SYSCALL {
		kdebugln("FD: ", fd, " text: ", text, " length: ", length)
	}
	if fd < 1 || fd > 2 {
		return 0, syscall.EBADF
	}

	if !currentThread.domain.MemorySpace.isRangeAccessible(text, text+uintptr(length)) {
		return 0, syscall.EFAULT
	}

	addr, ok := currentThread.domain.MemorySpace.GetPhysicalAddress(text)
	if !ok {
		kerrorln("Could not look up string addr")
		return 0, syscall.EFAULT
	}
	s := unsafe.String((*byte)(unsafe.Pointer(addr)), length)

	if fd == 2 {
		kerror(s)
	} else {
		kprint(s)
	}
	return uint32(len(s)), ESUCCESS //TODO: nr of bytes written
}

func linuxReadSyscall(args syscallArgs) (uint32, syscall.Errno) {
	fd := args.arg1
	buf := args.arg2
	count := args.arg3
	addr, ok := currentThread.domain.MemorySpace.GetPhysicalAddress(uintptr(buf))
	if !ok {
		kerrorln("Could not look up read addr")
		return 0, syscall.EFAULT
	}
	arr := (*[1 << 30]byte)(unsafe.Pointer(addr))[:count]

	var num uint32 = 0

	if fd == 0 {
		for num == 0 {
			for !serialDevice.HasReceivedData() {
				Block()
			}

			for serialDevice.HasReceivedData() && num < count {
				char := serialDevice.Read()
				arr[num] = char
				num++
			}
		}
	} else if fd == 42 {
		for num == 0 {
			for buffer.Len() == 0 {
				Block()
			}

			for buffer.Len() > 0 && num < count {
				raw_key := buffer.Pop().Keycode
				pressed := raw_key&0x80 == 0
				key := raw_key & 0x7f
				if pressed {
					arr[num] = translateKeycode(key)
					num++
				}
			}
		}
	}
	return num, ESUCCESS
}

func linuxFutexSyscall(args syscallArgs) (uint32, syscall.Errno) {
	uaddr := args.arg1
	futex_op := args.arg2
	val := args.arg3
	timeout := args.arg4

	if futex_op&FUTEX_PRIVATE_FLAG == 0 {
		kerrorln("Futex on shared futexes is not supported")
		return 0, syscall.ENOSYS
	}

	addr, ok := currentThread.domain.MemorySpace.GetPhysicalAddress(uintptr(uaddr))
	if !ok {
		kerrorln("Could not look up read addr")
		return 0, syscall.EFAULT
	}
	futexAddr := (*uint32)(unsafe.Pointer(addr))
	switch futex_op & 0xf {
	case FUTEX_WAIT:
		if timeout != 0 {
			//kerrorln("Timeouts are not supported yet")
			return 0, ESUCCESS //syscall.ENOSYS
		}
		// This should be atomically
		if val != *futexAddr {
			return 0, syscall.EAGAIN
		}

		currentThread.isBlocked = true
		currentThread.waitAddress = futexAddr
		return 0, ESUCCESS
	case FUTEX_WAKE:
		var woken uint32 = 0
		cur := currentThread.next
		for cur != currentThread && woken < val {
			if cur.isBlocked && cur.waitAddress == futexAddr {
				cur.isBlocked = false
				cur.waitAddress = nil
				woken++
			}
			cur = cur.next
		}
		return woken, ESUCCESS
	default:
		kerrorln("Unsupported futex op", futex_op)
		return 0, syscall.ENOSYS
	}
}

func linuxSchedYieldSyscall(rgs syscallArgs) (uint32, syscall.Errno) {
	Block()
	return 0, ESUCCESS
}

func unsupportedSyscall() {
	kerrorln("\nUnsupported Linux syscall received! Disabling interrupts and halting")
	kprintln("Syscall Number: ", uintptr(currentThread.regs.EAX), " (", uint32(currentThread.regs.EAX), ")")
	panicHelper(currentThread)
}
