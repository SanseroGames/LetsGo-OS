package syscall

import (
	"syscall"
	"unsafe"

	"github.com/sanserogames/letsgo-os/kernel"
	"github.com/sanserogames/letsgo-os/kernel/log"
	"github.com/sanserogames/letsgo-os/kernel/mm"
	"github.com/sanserogames/letsgo-os/kernel/panic"
	"github.com/sanserogames/letsgo-os/kernel/utils"
)

const PRINT_SYSCALL = kernel.ENABLE_DEBUG

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
		panic.KernelPanic("Too many syscalls registered")
	}
	syscallList = append(syscallList, syscallEntry{handler: handler, name: name, numberOfArgs: 0})
	registeredSyscalls[number] = byte(len(syscallList)) // add 1 to distinguish uninitialized
}

func InitSyscall() {
	kernel.SetInterruptHandler(0x80, linuxSyscallHandler, kernel.KCS_SELECTOR, kernel.PRIV_USER)

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
	return kernel.CurrentThread.Tid, ESUCCESS
}

func getPidSyscall(args syscallArgs) (uint32, syscall.Errno) {
	return kernel.CurrentThread.Domain.Pid, ESUCCESS
}

func linuxSyscallHandler() {
	var ret uint32 = 0
	var err syscall.Errno = ESUCCESS
	syscallNr := kernel.CurrentThread.Regs.EAX
	args := syscallArgs{
		arg1: kernel.CurrentThread.Regs.EBX,
		arg2: kernel.CurrentThread.Regs.ECX,
		arg3: kernel.CurrentThread.Regs.EDX,
		arg4: kernel.CurrentThread.Regs.ESI,
		arg5: kernel.CurrentThread.Regs.EDI,
		arg6: kernel.CurrentThread.Regs.EBP,
	}
	if kernel.CurrentThread.IsKernelInterrupt {
		syscallNr = kernel.CurrentThread.KernelRegs.EAX
		args = syscallArgs{
			arg1: kernel.CurrentThread.KernelRegs.EBX,
			arg2: kernel.CurrentThread.KernelRegs.ECX,
			arg3: kernel.CurrentThread.KernelRegs.EDX,
			arg4: kernel.CurrentThread.KernelRegs.ESI,
			arg5: kernel.CurrentThread.KernelRegs.EDI,
			arg6: kernel.CurrentThread.KernelRegs.EBP,
		}
		if syscallNr == syscall.SYS_WRITE {
			// Write syscall. We probably upset the go runtime so it wants to complain
			linuxWriteSyscall(args)
		}
		log.KPrintLn("\nkernel syscallnr: ", syscallNr)
		panic.KernelPanic("Why is the kernel making a syscall?")
	}
	handlerIdx := registeredSyscalls[syscallNr]
	if handlerIdx == 0 {
		unsupportedSyscall()
		return
	}
	handler := syscallList[handlerIdx-1]
	if PRINT_SYSCALL {
		log.KDebugLn("pid: ",
			kernel.CurrentThread.Domain.Pid,
			" tid: ",
			kernel.CurrentThread.Tid,
			" :: ",
			handler.name,
			" (",
			syscallNr,
			" hex(",
			uintptr(syscallNr),
			")",
			")")
		log.KDebugLn(" |- arg1:",
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
		log.KDebugLn("SYSCALL RETURN pid: ",
			kernel.CurrentThread.Domain.Pid,
			" tid: ",
			kernel.CurrentThread.Tid,
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
	kernel.CurrentThread.Regs.EAX = ret
}

func linuxExecveSyscall(args syscallArgs) (uint32, syscall.Errno) {
	pathname := args.arg1
	argv := args.arg2

	// Create new domain
	newDomainMem := mm.AllocPage()
	newDomainMem.Clear()
	newDomain := (*kernel.Domain)(newDomainMem.Pointer())
	newThreadMem := mm.AllocPage()
	newThreadMem.Clear()
	newThread := (*kernel.Thread)(newThreadMem.Pointer())

	err := kernel.StartProgramUsr(uintptr(pathname), uintptr(argv), newDomain, newThread)
	if err != ESUCCESS {
		mm.FreePage(newDomainMem.Pointer())
		mm.FreePage(newThreadMem.Pointer())
		return 0, err
	}
	newDomain.MemorySpace.MapPage(newThreadMem.Address(), newThreadMem.Address(), kernel.PAGE_RW|kernel.PAGE_PERM_KERNEL)
	kernel.AddDomain(newDomain)

	// if !kernel.CurrentThread.isFork {
	// 	// We were not in a fork, so the process should be replaced
	// 	// We simulate this by just exiting the current process
	// 	//panic.KernelPanic("Execve in non fork thread")
	// 	//ExitDomain(kernel.CurrentThread.Domain)
	// } else {
	// 	kernel.ExitThread(kernel.CurrentThread)
	// }
	kernel.PerformSchedule = true
	// return kernel.CurrentThread.regs.EAX, ESUCCESS
	return newDomain.Pid, ESUCCESS
}

func linuxWaitPidSyscall(args syscallArgs) (uint32, syscall.Errno) {
	waitPid := args.arg1
	// status := args.arg2
	// options := args.arg3
	// usage := args.arg3
	// log.KDebugln("Wait for ", waitPid)

	for {
		kernel.Block()
		notStarted := true
		stillWaiting := false
		for cur := kernel.AllDomains.Head; notStarted || cur != kernel.AllDomains.Head; cur = cur.Next {
			if cur.Pid == waitPid {
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
	kernel.ExitDomain(kernel.CurrentThread.Domain)
	kernel.PerformSchedule = true
	// Already in new context so return value from last syscall from current domain
	return kernel.CurrentThread.Regs.EAX, ESUCCESS
}

func linuxExitSyscall(args syscallArgs) (uint32, syscall.Errno) {
	kernel.ExitThread(kernel.CurrentThread)
	kernel.PerformSchedule = true
	// Already in new context so return value from last syscall from current domain
	return kernel.CurrentThread.Regs.EAX, ESUCCESS
}

func rebootHandler(args syscallArgs) (uint32, syscall.Errno) {
	magic1 := args.arg1
	magic2 := args.arg2
	cmd := args.arg3
	if magic1 != REBOOT_MAGIC1 && magic2 != REBOOT_MAGIC2 {
		return 0, syscall.EINVAL
	}
	if cmd != REBOOT_CMD_POWEROFF {
		log.KErrorLn("invalid reboot command")
		return 0, syscall.EINVAL
	}
	kernel.Shutdown()
	return 0, syscall.EINVAL
}

func linuxStatxSyscall(args syscallArgs) (uint32, syscall.Errno) {
	//dirfd := args.arg1
	//path := args.arg2
	//flags := args.arg3
	//mask := args.arg4
	buf := args.arg5
	addr, ok := kernel.CurrentThread.Domain.MemorySpace.GetPhysicalAddress(uintptr(buf))
	if !ok {
		log.KErrorLn("invalid adress in statx")
		return 0, syscall.EFAULT
	}
	item := (*statxData)(unsafe.Pointer(addr))
	mm.Memclr(addr, int(unsafe.Sizeof(item)))
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
	addr, ok := kernel.CurrentThread.Domain.MemorySpace.GetPhysicalAddress(uintptr(buf))
	if !ok {
		log.KErrorLn("invalid adress in uname")
		return 0, syscall.EFAULT
	}
	provided := (*utsname)(unsafe.Pointer(addr))
	*provided = uts
	return 0, ESUCCESS
}

func linuxCloneSyscall(args syscallArgs) (uint32, syscall.Errno) {
	flags := args.arg1
	stack := args.arg2
	newThreadMem := mm.AllocPage()
	newThreadMem.Clear()
	newThread := (*kernel.Thread)(newThreadMem.Pointer())
	kernel.CreateNewThread(newThread, uintptr(stack), kernel.CurrentThread, kernel.CurrentThread.Domain)
	kernel.CurrentThread.Domain.MemorySpace.MapPage(newThreadMem.Address(), newThreadMem.Address(), kernel.PAGE_RW|kernel.PAGE_PERM_KERNEL)
	if flags&_CLONE_THREAD == 0 {
		// This is probably temporary as I don't want to implement COW right now to create a new process
		log.KDebugLn("[CLONE SYSCALL] Clone where the goal is not a thread does not behave like on linux")
		// newThread.isFork = true
		panic.KernelPanic("No forking here!")
	}
	// Need to make this better at some point
	return newThread.Tid, ESUCCESS
}

func linuxMincoreSyscall(args syscallArgs) (uint32, syscall.Errno) {
	//addr := args.arg1
	length := args.arg2
	vec := args.arg3
	vecAddr, ok := kernel.CurrentThread.Domain.MemorySpace.GetPhysicalAddress(uintptr(vec))
	if !ok {
		log.KErrorLn("Could not look up vec array")
		return 0, syscall.EFAULT
	}

	arr := (*[30 << 1]byte)(unsafe.Pointer(vecAddr))[:(length+kernel.PAGE_SIZE-1)/kernel.PAGE_SIZE]
	for i := range arr {
		arr[i] = 1
	}
	return 0, ESUCCESS
}

func linuxMunmapSyscall(args syscallArgs) (uint32, syscall.Errno) {
	baseAddr := args.arg1
	length := args.arg2
	//printRegisters(currentInfo, Regs)
	for i := uint32(0); i < length; i += kernel.PAGE_SIZE {
		addr := uintptr(baseAddr + i)

		if addr < kernel.KERNEL_RESERVED {
			return 0, syscall.EINVAL
		}
		kernel.CurrentThread.Domain.MemorySpace.UnmapPage(addr)
	}
	return 0, ESUCCESS
}

func linuxBrkSyscall(args syscallArgs) (uint32, syscall.Errno) {
	newBrk := uintptr(args.arg1)
	brk := kernel.CurrentThread.Domain.MemorySpace.Brk
	if newBrk == 0 {
		return uint32(brk), ESUCCESS
	}
	if newBrk == brk || newBrk < brk {
		return uint32(brk), ESUCCESS
	}
	//text_mode_print_hex32(brk)
	for i := (brk + kernel.PAGE_SIZE - 1) &^ (kernel.PAGE_SIZE - 1); i < newBrk; i += kernel.PAGE_SIZE {
		p := mm.AllocPage()
		p.Clear()
		flags := uint8(kernel.PAGE_PERM_USER | kernel.PAGE_RW)
		// log.KDebugln("[brk] Map page ", i, " -> ", p)

		kernel.CurrentThread.Domain.MemorySpace.MapPage(p.Address(), i, flags)
	}
	kernel.CurrentThread.Domain.MemorySpace.Brk = newBrk
	// log.KDebugln("BRK: ", newBrk)
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
		target = kernel.CurrentThread.Domain.MemorySpace.VmTop
	}
	if target+uintptr(size) < kernel.KERNEL_RESERVED {
		return 0, syscall.EINVAL
	}
	if prot == 0 {
		return uint32(target), ESUCCESS
	}

	startAddr := kernel.CurrentThread.Domain.MemorySpace.FindSpaceFor(target, size)

	if flags&MMAP_MAP_FIXED == MMAP_MAP_FIXED {
		startAddr = target
	} else if startAddr == 0 {
		return 0, syscall.EINVAL
	}

	for i := startAddr; i < startAddr+size; i += kernel.PAGE_SIZE {
		p := mm.AllocPage()
		p.Clear()
		pageFlags := uint8(kernel.PAGE_PERM_USER)
		if prot&MMAP_PROT_WRITE == MMAP_PROT_WRITE {
			pageFlags |= kernel.PAGE_RW
		}

		if kernel.CurrentThread.Domain.MemorySpace.GetPageTableEntry(i).IsPresent() {
			if flags&MMAP_MAP_FIXED == MMAP_MAP_FIXED {
				// TODO: Clear or free and remap page?
				// kernel.CurrentThread.Domain.MemorySpace.unMapPage(i)
			} else {
				// TODO: What to do here?
				panic.KernelPanic("Trying to map page which is already present without MAP_FIXED")
			}
		} else {
			kernel.CurrentThread.Domain.MemorySpace.MapPage(p.Address(), i, pageFlags)
		}
	}

	return uint32(startAddr), ESUCCESS
}

func linuxSetThreadAreaSyscall(args syscallArgs) (uint32, syscall.Errno) {
	u_info := args.arg1
	addr, ok := kernel.CurrentThread.Domain.MemorySpace.GetPhysicalAddress(uintptr(u_info))
	if !ok {
		log.KErrorLn("Could not look up user desc")
		return 0, syscall.EFAULT
	}
	desc := (*kernel.UserDesc)(unsafe.Pointer(addr))
	if desc.Flags&kernel.UDESC_SEG_NOT_PRESENT != 0 {
		log.KErrorLn("fixme: not handling updating entries")
		return 0, syscall.ENOSYS
	}

	slot := desc.EntryNumber
	if slot == 0xffffffff {
		slot = kernel.FindFreeTlsSlot()
		if slot == 0xffffffff {
			// There was no free slot
			return 0, syscall.ESRCH
		}
		desc.EntryNumber = slot
	}
	kernel.SetTlsSegment(slot, desc)

	return 0, ESUCCESS
}

func linuxOpenSyscall(args syscallArgs) (uint32, syscall.Errno) {
	//path := args.arg1
	flags := args.arg2
	addr, ok := kernel.CurrentThread.Domain.MemorySpace.GetPhysicalAddress(uintptr(flags))
	if !ok {
		return 0, syscall.EFAULT
	}
	s := utils.CString(addr)
	if PRINT_SYSCALL {
		log.KDebugLn("[SYS-OPEN] ", s)
	}
	//text_mode_println(s)
	return 0, syscall.ENOSYS
	//printRegisters(currentInfo, Regs)

}

func linuxOpenAtSyscall(args syscallArgs) (uint32, syscall.Errno) {
	fd := args.arg1
	path := args.arg2
	flags := args.arg3
	pathaddr, ok := kernel.CurrentThread.Domain.MemorySpace.GetPhysicalAddress(uintptr(path))
	if !ok {
		return 0, syscall.EFAULT
	}
	s1 := utils.CString(pathaddr)
	if PRINT_SYSCALL {
		log.KDebugLn("[SYS-OPENAT] fd:", fd)
		log.KDebugLn("[SYS-OPENAT] path:", s1)
		log.KDebugLn("[SYS-OPENAT] flags:", flags)
	}

	if s1 == "/dev/null" {
		return 42, ESUCCESS
	}

	//text_mode_println(s)
	return 0, syscall.ENOSYS
	//printRegisters(currentInfo, Regs)

}

func linuxWriteVSyscall(args syscallArgs) (uint32, syscall.Errno) {
	fd := args.arg1
	arr := uintptr(args.arg2)
	count := args.arg3
	if fd < 1 || fd > 2 {
		return 0, syscall.EBADF
	}

	if count <= 0 {
		return 0, syscall.EINVAL
	}

	processed := 0
	printed := uint32(0)
	for item, err := range mm.IterateUserSpaceType[ioVec](arr, &kernel.CurrentDomain.MemorySpace) {
		if err != ESUCCESS {
			return 0, err
		}
		written, err := writeSyscallWriteData(fd, item.iovBase, item.iovLen)
		if err != ESUCCESS {
			return 0, err
		}
		printed += written
		processed++
		if processed == int(count) {
			break
		}
	}
	return printed, ESUCCESS
}

func linuxWriteSyscall(args syscallArgs) (uint32, syscall.Errno) {
	fd := args.arg1
	text := uintptr(args.arg2)
	length := args.arg3
	if PRINT_SYSCALL {
		log.KDebugLn("FD: ", fd, " text: ", text, " length: ", length)
	}
	if fd < 1 || fd > 2 {
		return 0, syscall.EBADF
	}
	return writeSyscallWriteData(fd, text, length)
}

func writeSyscallWriteData(fd uint32, text uintptr, length uint32) (uint32, syscall.Errno) {
	var buf [100]byte

	s := unsafe.String(&buf[0], len(buf))

	index := 0
	writeLen := uint32(0)
	for value, err := range kernel.CurrentDomain.MemorySpace.IterateUserSpace(text) {
		if err != ESUCCESS {
			return 0, err
		}
		buf[index] = value
		index++
		if index == len(buf) {
			if fd == 2 {
				log.KError(s)
			} else {
				log.KPrint(s)
			}
			index = 0
		}
		writeLen++
		if writeLen == length {
			break
		}
	}
	rest := unsafe.String(&buf[0], index)
	if fd == 2 {
		log.KError(rest)
	} else {
		log.KPrint(rest)
	}
	return uint32(writeLen), ESUCCESS
}

func linuxReadSyscall(args syscallArgs) (uint32, syscall.Errno) {
	fd := args.arg1
	buf := args.arg2
	count := args.arg3
	addr, ok := kernel.CurrentThread.Domain.MemorySpace.GetPhysicalAddress(uintptr(buf))
	if !ok {
		log.KErrorLn("Could not look up read addr")
		return 0, syscall.EFAULT
	}
	arr := (*[1 << 30]byte)(unsafe.Pointer(addr))[:count]

	var num uint32 = 0

	if fd == 0 {
		for num == 0 {
			for !kernel.SerialDevice.HasReceivedData() {
				kernel.Block()
			}

			for kernel.SerialDevice.HasReceivedData() && num < count {
				char := kernel.SerialDevice.Read()
				arr[num] = char
				num++
			}
		}
	}
	// I don't use keyboard anymore, since I have serial.
	// When I implmentf device drivers, I might add this again
	// else if fd == 42 {
	// 	for num == 0 {
	// 		for buffer.Len() == 0 {
	// 			kernel.Block()
	// 		}

	// 		for buffer.Len() > 0 && num < count {
	// 			raw_key := buffer.Pop().Keycode
	// 			pressed := raw_key&0x80 == 0
	// 			key := raw_key & 0x7f
	// 			if pressed {
	// 				arr[num] = translateKeycode(key)
	// 				num++
	// 			}
	// 		}
	// 	}
	// }
	return num, ESUCCESS
}

func linuxFutexSyscall(args syscallArgs) (uint32, syscall.Errno) {
	uaddr := args.arg1
	futex_op := args.arg2
	val := args.arg3
	timeout := args.arg4

	if futex_op&FUTEX_PRIVATE_FLAG == 0 {
		log.KErrorLn("Futex on shared futexes is not supported")
		return 0, syscall.ENOSYS
	}

	addr, ok := kernel.CurrentThread.Domain.MemorySpace.GetPhysicalAddress(uintptr(uaddr))
	if !ok {
		log.KErrorLn("Could not look up read addr")
		return 0, syscall.EFAULT
	}
	futexAddr := (*uint32)(unsafe.Pointer(addr))
	switch futex_op & 0xf {
	case FUTEX_WAIT:
		if timeout != 0 {
			//log.KErrorln("Timeouts are not supported yet")
			// log.KDebugLn("Timeout futex")
			return 0, ESUCCESS //syscall.ENOSYS
		}
		// This should be atomically, but we're not multithreaded (as in multicore) yet so it does not matter
		if val != *futexAddr {
			return 0, syscall.EAGAIN
		}

		kernel.CurrentThread.IsBlocked = true
		kernel.CurrentThread.WaitAddress = futexAddr
		return 0, ESUCCESS
	case FUTEX_WAKE:
		var woken uint32 = 0
		cur := kernel.CurrentThread.Next
		for cur != kernel.CurrentThread && woken < val {
			if cur.IsBlocked && cur.WaitAddress == futexAddr {
				cur.IsBlocked = false
				cur.WaitAddress = nil
				woken++
			}
			cur = cur.Next
		}
		return woken, ESUCCESS
	default:
		log.KErrorLn("Unsupported futex op", futex_op)
		return 0, syscall.ENOSYS
	}
}

func linuxSchedYieldSyscall(rgs syscallArgs) (uint32, syscall.Errno) {
	kernel.Block()
	return 0, ESUCCESS
}

func unsupportedSyscall() {
	log.KErrorLn("\nUnsupported Linux syscall received! Disabling interrupts and halting")
	log.KPrintLn("Syscall Number: ", uintptr(kernel.CurrentThread.Regs.EAX), " (", uint32(kernel.CurrentThread.Regs.EAX), ")")
	panic.KernelPanic("Unsupported syscall")
}
