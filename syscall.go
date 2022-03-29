package main

import (
    "unsafe"
    "reflect"
    "syscall"
)

const PRINT_SYSCALL = false

type ioVec struct {
    iovBase uintptr    /* Starting address */
    iovLen uint32    /* Number of bytes to transfer */
};

const _UTSNAME_LENGTH = 65

type utsname struct {
    /* Name of the implementation of the operating system.  */
    sysname [_UTSNAME_LENGTH]byte
    /* Name of this node on the network.  */
    nodename[_UTSNAME_LENGTH]byte
    /* Current release level of this implementation.  */
    release[_UTSNAME_LENGTH]byte
    /* Current version level of this release.  */
    version[_UTSNAME_LENGTH]byte
    /* Name of the hardware type the system is running on.  */
    machine[_UTSNAME_LENGTH]byte

    domainname [_UTSNAME_LENGTH]byte
}


type statxDataTimestamp struct {
    tv_sec int64    /* Seconds since the Epoch (UNIX time) */
    tv_nsec uint32   /* Nanoseconds since tv_sec */
    reserved uint32
};

type statxData struct {

    stx_mask uint32        /* Mask of bits indicating
                                   filled fields */
    stx_blksize uint32     /* Block size for filesystem I/O */
    stx_attributes uint64  /* Extra file attribute indicators */
    stx_nlink uint32       /* Number of hard links */
    stx_uid uint32         /* User ID of owner */
    stx_gid uint32         /* Group ID of owner */
    stx_mode uint16        /* File type and mode */
    reserved uint16
    stx_ino uint64        /* Inode number */
    stx_size uint64        /* Total size in bytes */
    stx_blocks uint64      /* Number of 512B blocks allocated */
    stx_attributes_mask uint64
                           /* Mask to show what's supported
                              in stx_attributes */

    /* The following fields are file timestamps */
    stx_atime statxDataTimestamp  /* Last access */
    stx_btime statxDataTimestamp /* Creation */
    stx_ctime statxDataTimestamp /* Last status change */
    stx_mtime statxDataTimestamp /* Last modification */

    /* If this file represents a device, then the next two
       fields contain the ID of the device */
    stx_rdev_major uint32  /* Major ID */
    stx_rdev_minor uint32  /* Minor ID */

    /* The next two fields contain the ID of the device
       containing the filesystem where the file resides */
    stx_dev_major uint32   /* Major ID */
    stx_dev_minor uint32   /* Minor ID */
}

type syscallHandler func(syscallArgs) (uint32, syscall.Errno)
type syscallList struct {
    handler syscallHandler
    name string
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

    __S_IFDIR = 0040000	/* Directory.  */
    __S_IFCHR = 0020000	/* Character device.  */
    __S_IFBLK = 0060000	/* Block device.  */
    __S_IFREG = 0100000	/* Regular file.  */
    __S_IFIFO = 0010000	/* FIFO.  */

    STATX_TYPE = 0x00000001	/* Want/got stx_mode & S_IFMT */
    STATX_MODE = 0x00000002	/* Want/got stx_mode & ~S_IFMT */
    STATX_NLINK = 0x00000004	/* Want/got stx_nlink */
    STATX_UID = 0x00000008	/* Want/got stx_uid */
    STATX_GID = 0x00000010	/* Want/got stx_gid */
    STATX_ATIME = 0x00000020	/* Want/got stx_atime */
    STATX_MTIME = 0x00000040	/* Want/got stx_mtime */
    STATX_CTIME = 0x00000080	/* Want/got stx_ctime */
    STATX_INO = 0x00000100	/* Want/got stx_ino */
    STATX_SIZE = 0x00000200	/* Want/got stx_size */
    STATX_BLOCKS = 0x00000400	/* Want/got stx_blocks */
    STATX_BASIC_STATS = 0x000007ff	/* The stuff in the normal stat struct */
    STATX_BTIME = 0x00000800	/* Want/got stx_btime */
    STATX_MNT_ID = 0x00001000	/* Got stx_mnt_id */

    FUTEX_WAIT = 0
    FUTEX_WAKE = 1
    FUTEX_FD = 2
    FUTEX_REQUEUE = 3
    FUTEX_CMP_REQUEUE = 4
    FUTEX_WAKE_OP = 5
    FUTEX_LOCK_PI = 6
    FUTEX_UNLOCK_PI	= 7
    FUTEX_TRYLOCK_PI = 8
    FUTEX_WAIT_BITSET = 9
    FUTEX_WAKE_BITSET = 10
    FUTEX_WAIT_REQUEUE_PI = 11
    FUTEX_CMP_REQUEUE_PI = 12

    FUTEX_PRIVATE_FLAG = 128

    ESUCCESS = syscall.Errno(0)
)

var (
    uts = utsname{
        sysname: [_UTSNAME_LENGTH]byte{'L','e','t','\'','s','G','o','!',' ','O','S',0},
        nodename: [_UTSNAME_LENGTH]byte{0},
        release: [_UTSNAME_LENGTH]byte{'6','9','.','4','.','2','0',0},
        version: [_UTSNAME_LENGTH]byte{'0','.','1',0},
        machine: [_UTSNAME_LENGTH]byte{0},
        domainname: [_UTSNAME_LENGTH]byte{0},
    }

    // Currently this wastes an awful lot of memory, but maps don't work. They don't initialise themself...
    registeredSyscalls = [0x200]syscallList{}
    okHandler = func(args syscallArgs) (uint32, syscall.Errno) {return 0, ESUCCESS}
    invalHandler = func(args syscallArgs) (uint32, syscall.Errno) {return 0, syscall.EINVAL}
)

// Not quite happy with the args here. Annoying that you have to redeclare variables
// in the implementation if you want meaningful names. Also you can't have variables
// that are for docu or implemented later due to go not liking unused variables.
func syscalltest(args syscallArgs) (uint32, syscall.Errno) {
    return 0, syscall.EFAULT
}

func RegisterSyscall(number int, name string, handler syscallHandler) {
    registeredSyscalls[number] = syscallList{handler: handler, name: name}
}

func InitSyscall() {
    SetInterruptHandler(0x80, linuxSyscallHandler, KCS_SELECTOR, PRIV_USER)

    RegisterSyscall(syscall.SYS_WRITE, "write syscall", linuxWriteSyscall)
    RegisterSyscall(syscall.SYS_WRITEV, "writeV syscall", linuxWriteVSyscall)
    RegisterSyscall(syscall.SYS_SET_THREAD_AREA, "set thread area syscall", linuxSetThreadAreaSyscall)
    RegisterSyscall(syscall.SYS_OPEN, "open syscall", linuxOpenSyscall)
    RegisterSyscall(syscall.SYS_OPENAT, "open at syscall", func(args syscallArgs) (uint32, syscall.Errno) {return 0, syscall.ENOSYS})
    RegisterSyscall(syscall.SYS_CLOSE, "close syscall", okHandler)
    RegisterSyscall(syscall.SYS_READ, "read syscall", linuxReadSyscall)
    RegisterSyscall(syscall.SYS_READLINK, "readlink syscall", okHandler)
    RegisterSyscall(syscall.SYS_READLINKAT, "read link at syscall", invalHandler)
    RegisterSyscall(syscall.SYS_FCNTL64, "fcntl64 syscall", invalHandler)
    RegisterSyscall(syscall.SYS_SCHED_GETAFFINITY, "sched get affinity syscall", func(args syscallArgs) (uint32, syscall.Errno) {return 0xffffffff, ESUCCESS})
    RegisterSyscall(syscall.SYS_NANOSLEEP, "nano sleep syscall", invalHandler)
    RegisterSyscall(syscall.SYS_EXIT_GROUP, "exit group syscall", linuxExitGroupSyscall)
    RegisterSyscall(syscall.SYS_EXIT, "exit syscall", okHandler)
    RegisterSyscall(syscall.SYS_BRK, "brk syscall", linuxBrkSyscall)
    RegisterSyscall(syscall.SYS_MMAP2, "mmap2 syscall", linuxMmap2Syscall)
    RegisterSyscall(syscall.SYS_MINCORE, "mincore syscall", linuxMincoreSyscall)
    RegisterSyscall(syscall.SYS_MUNMAP, "munmap syscall", linuxMunmapSyscall)
    RegisterSyscall(syscall.SYS_CLOCK_GETTIME, "clock get time syscall", func(args syscallArgs) (uint32, syscall.Errno) {return 0, syscall.ENOTSUP})
    RegisterSyscall(syscall.SYS_RT_SIGPROCMASK, "sig proc mask syscall", okHandler)
    RegisterSyscall(syscall.SYS_SIGALTSTACK, "sig alt stack syscall", okHandler)
    RegisterSyscall(syscall.SYS_RT_SIGACTION, "rt sig action syscall", okHandler)
    RegisterSyscall(syscall.SYS_GETTID, "gettid syscall", getTidSyscall)
    RegisterSyscall(syscall.SYS_GETPID, "get pid syscall", getPidSyscall)
    RegisterSyscall(syscall.SYS_SET_TID_ADDRESS, "set tid address syscall", okHandler)
    RegisterSyscall(syscall.SYS_POLL, "poll syscall", okHandler)
    RegisterSyscall(syscall.SYS_CLONE, "clone syscall", linuxCloneSyscall)
    RegisterSyscall(syscall.SYS_FUTEX, "futex syscall", linuxFutexSyscall)
    RegisterSyscall(syscall.SYS_SCHED_YIELD, "sched yield syscall", okHandler)
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
    RegisterSyscall(0x193, "clock gettime 64 syscall", invalHandler)
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
    if kernelInterrupt {
        syscallNr = currentThread.kernelRegs.EAX
        args = syscallArgs{
            arg1: currentThread.kernelRegs.EBX,
            arg2: currentThread.kernelRegs.ECX,
            arg3: currentThread.kernelRegs.EDX,
            arg4: currentThread.kernelRegs.ESI,
            arg5: currentThread.kernelRegs.EDI,
            arg6: currentThread.kernelRegs.EBP,
        }
        if false {
        text_mode_print("kernel syscallnr: ")
        text_mode_print_hex32(syscallNr)
        text_mode_println("")
        kernelPanic("Why is the kernel making a syscall?")
        }
    }
    handler := registeredSyscalls[syscallNr]
    if handler.handler == nil {
        unsupportedSyscall()
        return
    }
    if PRINT_SYSCALL {
        kprintln("pid: ",
                currentThread.domain.pid,
                " tid: ",
                currentThread.tid,
                " :: ",
                handler.name,
                "(",
                syscallNr,
                ")")
    }
    ret, err = handler.handler(args)
    if err != ESUCCESS {
        ret = ^uint32(err)+1
    }
    currentThread.regs.EAX = ret
}

func linuxExitGroupSyscall(args syscallArgs) (uint32, syscall.Errno) {
    //text_mode_print("Exiting domain ")
    //text_mode_print_hex(uint8(currentThread.domain.pid))
    //text_mode_println("")
    ExitDomain(currentThread.domain)
    PerformSchedule = true
    // Already in new context so return value from last syscall from current domain
    return currentThread.regs.EAX, ESUCCESS
}

func linuxStatxSyscall(args syscallArgs) (uint32, syscall.Errno) {
    //dirfd := args.arg1
    //path := args.arg2
    //flags := args.arg3
    //mask := args.arg4
    buf := args.arg5
    addr, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(buf))
    if !ok {
        text_mode_print_errorln("invalid adress in statx")
        return 0, syscall.EFAULT
    }
    item := (* statxData)(unsafe.Pointer(addr))
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
    addr, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(buf))
    if !ok {
        text_mode_print_errorln("invalid adress in uname")
        return 0, syscall.EFAULT
    }
    provided := (*utsname)(unsafe.Pointer(addr))
    *provided = uts
    return 0, ESUCCESS
}

func linuxCloneSyscall(args syscallArgs) (uint32, syscall.Errno) {
    flags := args.arg1
    stack := args.arg2
    //parent_tid := currentThread.regs.EDX
    //tls := currentThread.regs.ESI
    //child_tid := currentThread.regs.EDI
    //text_mode_print("flags:")
    //text_mode_print_hex32(flags)
    //text_mode_print(" stack:")
    //text_mode_print_hex32(stack)
    //text_mode_print(" parent:")
    //text_mode_print_hex32(parent_tid)
    //text_mode_print(" tls:")
    //text_mode_print_hex32(tls)
    //text_mode_print(" child:")
    //text_mode_print_hex32(child_tid)
    //text_mode_println("")
    if flags & _CLONE_THREAD == 0 {
        text_mode_print_errorln("Clone where the goal is not a thread is not supported")
        return 0, syscall.EINVAL
    }
    // Need to make this better at some point
    newThreadMem := allocPage()
    Memclr(newThreadMem, PAGE_SIZE)
    newThread := (* thread)(unsafe.Pointer(newThreadMem))
    CreateNewThread(newThread, uintptr(stack), currentThread, currentThread.domain)
    currentThread.domain.MemorySpace.mapPage(newThreadMem, newThreadMem, PAGE_RW | PAGE_PERM_KERNEL)
    return newThread.tid, ESUCCESS
}

func linuxMincoreSyscall(args syscallArgs) (uint32, syscall.Errno) {
    //addr := args.arg1
    length := args.arg2
    vec := args.arg3
    //text_mode_print(" addr: ")
    //text_mode_print_hex32(currentThread.regs.EBX)
    //text_mode_print(" length: ")
    //text_mode_print_hex32(currentThread.regs.ECX)
    //text_mode_print(" vec: ")
    //text_mode_print_hex32(currentThread.regs.EDX)
    //text_mode_println("")
    //addr := currentThread.regs.EBX
    vecAddr, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(vec))
    if !ok {
        text_mode_print_errorln("Could not look up vec array")
        return 0, syscall.EFAULT
    }

    arr := (*[30 << 1]byte)(unsafe.Pointer(vecAddr))[:(length+PAGE_SIZE-1) / PAGE_SIZE]
    for i := range arr {
        arr[i] = 1
    }
    return 0, ESUCCESS
}

func linuxMunmapSyscall(args syscallArgs) (uint32, syscall.Errno) {
    baseAddr := args.arg1
    length := args.arg2
    //printRegisters(currentInfo, regs)
    for i:= uint32(0); i < length; i += PAGE_SIZE {
        addr := uintptr(baseAddr + i)

        if addr < KERNEL_RESERVED {
            return 0, syscall.EINVAL
        }
        currentThread.domain.MemorySpace.unMapPage(addr)
    }
    return 0, ESUCCESS
}

func linuxBrkSyscall(args syscallArgs) (uint32, syscall.Errno) {
    newBrk := args.arg1
    brk := uint32(currentThread.domain.MemorySpace.Brk)
    if newBrk == 0 {
        return brk, ESUCCESS
    }
    if newBrk == brk || newBrk < brk {
        return brk, ESUCCESS
    }
    //text_mode_print_hex32(brk)
    for i:= (brk + PAGE_SIZE - 1) &^ (PAGE_SIZE - 1); i < newBrk; i+=PAGE_SIZE {
        p := allocPage()
        Memclr(p, PAGE_SIZE)
        flags := uint8(PAGE_PERM_USER | PAGE_RW)
        currentThread.domain.MemorySpace.mapPage(p, uintptr(i), flags)
    }
    currentThread.domain.MemorySpace.Brk = uintptr(newBrk)
    return newBrk, ESUCCESS
}

func linuxMmap2Syscall(args syscallArgs) (uint32, syscall.Errno) {
    target := args.arg1
    size := args.arg2
    prot := args.arg3
    if target == 0 {
        target = uint32(currentThread.domain.MemorySpace.VmTop)
    }
    if prot == 0 {
        return uint32(currentThread.domain.MemorySpace.VmTop), ESUCCESS
    }
    //text_mode_print("vmtop: ")
    //text_mode_print_hex32(uint32(currentThread.domain.MemorySpace.VmTop))
    //text_mode_print(" target: ")
    //text_mode_print_hex32(currentThread.regs.EBX)
    //text_mode_print(" size: ")
    //text_mode_print_hex32(size)
    //text_mode_print(" prot: ")
    //text_mode_print_hex32(currentThread.regs.EDX)
    //text_mode_print(" flags: ")
    //text_mode_print_hex32(currentThread.regs.ESI)
    //text_mode_println("")
    for i:=uint32(0); i < size; i+=PAGE_SIZE {
        p := allocPage()
        Memclr(p, PAGE_SIZE)
        flags := uint8(PAGE_PERM_USER | PAGE_RW)
        if(target+i < KERNEL_RESERVED){
            return 0, syscall.EINVAL
        }
        currentThread.domain.MemorySpace.mapPage(p, uintptr(target+i), flags)
    }
    //text_mode_print_hex32(target)
    return target, ESUCCESS
}

func linuxSetThreadAreaSyscall(args syscallArgs) (uint32, syscall.Errno) {
    u_info := args.arg1
    addr, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(u_info))
    if !ok {
        text_mode_print_errorln("Could not look up user desc")
        return 0, syscall.EFAULT
    }
    desc := (*UserDesc)(unsafe.Pointer(addr))
    if desc.Flags & UDESC_SEG_NOT_PRESENT != 0{
        text_mode_print_errorln("fixme: not handling updating entries")
        return  0, syscall.ENOSYS
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
    _, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(flags))
    if !ok {
        return 0, syscall.EFAULT
    }
    //s := cstring(addr)
    //text_mode_println(s)
    return 0, syscall.ENOSYS
    //printRegisters(currentInfo, regs)

}

func linuxWriteVSyscall(args syscallArgs) (uint32, syscall.Errno) {
    fd := args.arg1
    arr := args.arg2
    count := args.arg3
    if fd < 1 || fd > 2 {
        return 0, syscall.EBADF
    }

    addr, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(arr))
    if !ok {
        text_mode_print_errorln("Could not look up string addr")
        return 0, syscall.EFAULT
    }
    iovecs := *(*[]ioVec)(unsafe.Pointer(&reflect.SliceHeader{
        Len:  int(count),
	    Cap:  int(count),
	    Data: addr,
    }))
    printed := 0
    for _,n := range iovecs {
        addr, ok = currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(n.iovBase))
        var s string
        hdr := (*reflect.StringHeader)(unsafe.Pointer(&s)) // case 1
        hdr.Data = uintptr(unsafe.Pointer(uintptr(addr))) // case 6 (this case)
        hdr.Len = int(n.iovLen)
        if fd == 2 {
            text_mode_print_error(s)
        } else {
            text_mode_print(s)
        }
        printed += len(s)
    }
    return uint32(printed), ESUCCESS //TODO: Return number of bytes written
}

func linuxWriteSyscall(args syscallArgs) (uint32, syscall.Errno) {
    fd := args.arg1
    text := args.arg2
    length := args.arg3
    if fd < 1 || fd > 2 {
        return 0, syscall.EBADF
    }

    addr, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(text))
    if !ok {
        text_mode_print_errorln("Could not look up string addr")
        return 0, syscall.EFAULT
    }
    var s string
    hdr := (*reflect.StringHeader)(unsafe.Pointer(&s)) // case 1
    hdr.Data = uintptr(unsafe.Pointer(uintptr(addr))) // case 6 (this case)
    hdr.Len = int(length)
    // Test if it is a go panic and print infos to debug:

    //prevent stack trace
    //if stop > 10 {
    //    return uint32(len(s))
    //}
    //stop++


    if fd == 2 {
        text_mode_print_error(s)
    } else {
        text_mode_print(s)
    }
    return uint32(len(s)), ESUCCESS //TODO: nr of bytes written
}

func linuxReadSyscall(args syscallArgs) (uint32, syscall.Errno) {
    //fd := args.arg1
    buf := args.arg2
    count := args.arg3
    addr, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(buf))
    if !ok {
        text_mode_print_errorln("Could not look up read addr")
        return 0, syscall.EFAULT
    }
    arr := (*[1 <<30]byte)(unsafe.Pointer(addr))[:count]

    var num  uint32 = 0

    for num == 0 {
        for buffer.Len() == 0 {
            Block()
        }

        for buffer.Len() > 0 && num < count {
            raw_key := buffer.Pop().Keycode
            pressed := raw_key & 0x80 == 0
            key := raw_key & 0x7f
            if pressed {
                arr[num] = translateKeycode(key)
                num++
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
    //uaddr2 := args.arg5
    //val_3 := args.arg6
    //text_mode_print("uaddr: ")
    //text_mode_print_hex32(uaddr)
    //text_mode_print(" futex_op: ")
    //text_mode_print_hex32(futex_op)
    //text_mode_print(" val: ")
    //text_mode_print_hex32(val)
    //text_mode_print(" timeout: ")
    //text_mode_print_hex32(timeout)
    //text_mode_print(" uaddr2: ")
    //text_mode_print_hex32(uaddr2)
    //text_mode_print(" val_3: ")
    //text_mode_print_hex32(val_3)
    //text_mode_println("")

    if futex_op & FUTEX_PRIVATE_FLAG == 0 {
        text_mode_print_error("Futex on shared futexes is not supported")
        return 0, syscall.ENOSYS
    }

    addr, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(uaddr))
    if !ok {
        text_mode_print_errorln("Could not look up read addr")
        return 0, syscall.EFAULT
    }
    futexAddr := (*uint32)(unsafe.Pointer(addr))
    switch (futex_op & 0xf) {
        case FUTEX_WAIT:
            if timeout != 0 {
                text_mode_print_error("Timeouts are not supported yet")
                return 0, syscall.ENOSYS
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
            text_mode_print_error("Unsupported futex op")
            text_mode_print_hex32(futex_op)
            text_mode_println("")
            return 0, syscall.ENOSYS
    }
}

func unsupportedSyscall(){
    text_mode_print_char(0xa)
    text_mode_print_errorln("Unsupported Linux Syscall received! Disabling Interrupts and halting")
    text_mode_print("Syscall Number: ")
    text_mode_print_hex32(currentThread.regs.EAX)
    text_mode_println("")
    panicHelper(currentThread)
}


