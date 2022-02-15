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
)

func linuxSyscallHandler(){
    var ret uint32 = 0
    syscallNr := currentThread.regs.EAX
    arg1 := currentThread.regs.EBX
    arg2 := currentThread.regs.ECX
    arg3 := currentThread.regs.EDX
    arg4 := currentThread.regs.ESI
    arg5 := currentThread.regs.EDI
    arg6 := currentThread.regs.EBP
    if kernelInterrupt {
        syscallNr = currentThread.kernelRegs.EAX
        arg1 = currentThread.kernelRegs.EBX
        arg2 = currentThread.kernelRegs.ECX
        arg3 = currentThread.kernelRegs.EDX
        arg4 = currentThread.kernelRegs.ESI
        arg5 = currentThread.kernelRegs.EDI
        arg6 = currentThread.kernelRegs.EBP

        if false {
        text_mode_print("kernel syscallnr: ")
        text_mode_print_hex32(syscallNr)
        text_mode_println("")
        kernelPanic("Why is the kernel making a syscall?")
        }
    }

    switch (syscallNr) {
        case syscall.SYS_WRITE: {
            // Linux write syscall
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("write syscall")
            }
            ret = linuxWriteSyscall(arg1, arg2, arg3)
        }
        case syscall.SYS_WRITEV: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("writeV syscall")
            }
            ret = linuxWriteVSyscall(arg1, arg2, arg3)
        }
        case syscall.SYS_SET_THREAD_AREA: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("set thread area syscall")
            }
            ret = linuxSetThreadAreaSyscall(arg1)
        }
        case syscall.SYS_OPEN: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("open syscall")
            }

            ret = linuxOpenSyscall(arg1, arg2)
        }
        case syscall.SYS_OPENAT: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("openat syscall")
            }

            ret = ^uint32(syscall.ENOSYS)+1
        }
        case syscall.SYS_CLOSE: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("close syscall")
            }
        }
        case syscall.SYS_READ: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("read syscall")
            }
            ret = linuxReadSyscall(arg1, arg2, arg3)
        }
        case syscall.SYS_READLINK: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("readlink syscall")
            }
        }
        case syscall.SYS_READLINKAT: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("readlinkat syscall")
            }
            ret = ^uint32(syscall.EINVAL)+1
        }
        case syscall.SYS_FCNTL64: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("fcntl64 syscall")
            }
            ret = ^uint32(syscall.EINVAL)+1

        }
        case syscall.SYS_SCHED_GETAFFINITY: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("sched getaffinity syscall")
            }
            ret = 0xffffffff
        }
        case syscall.SYS_NANOSLEEP: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("nanosleep syscall")
            }
        }
        case syscall.SYS_EXIT_GROUP: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("exit group syscall")
            }
            ret = linuxExitGroupSyscall(arg1)
        }
        case syscall.SYS_EXIT: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("exit syscall")
            }

        }
        case syscall.SYS_BRK: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("brk syscall")
            }
            ret = linuxBrkSyscall(arg1)
        }
        case syscall.SYS_MMAP2: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("mmap2 syscall")
            }
            ret = linuxMmap2Syscall(arg1, arg2, arg3)
        }
        case syscall.SYS_MINCORE: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("mincore syscall")
            }
            ret = linuxMincoreSyscall(arg1, arg2, arg3)
        }
        case syscall.SYS_MUNMAP: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("munmap syscall")
            }
            ret = linuxMunmapSyscall(arg1, arg2)
        }
        case syscall.SYS_CLOCK_GETTIME: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("clock get time syscall")
            }
        }

        case syscall.SYS_RT_SIGPROCMASK: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("sig proc mask syscall")
            }
        }

        case syscall.SYS_SIGALTSTACK: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("sig alt stack syscall")
            }
        }

        case syscall.SYS_RT_SIGACTION: {
            if PRINT_SYSCALL {
                //printTid()
                //text_mode_println("rt sigaction syscall")
            }
        }
        case syscall.SYS_GETTID: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("gettid syscall")
            }
            ret = currentThread.tid
        }
        case syscall.SYS_GETPID: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("getpid syscall")
            }
            ret = currentThread.domain.pid
        }
        case syscall.SYS_SET_TID_ADDRESS: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("set_tid_address syscall")
            }
        }
        case syscall.SYS_POLL: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("poll syscall")
            }
        }
        case syscall.SYS_CLONE: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("clone syscall")
            }
            ret = linuxCloneSyscall(arg1, arg2)
        }
        case syscall.SYS_FUTEX: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("futex syscall")
            }
            ret = linuxFutexSyscall(arg1, arg2, arg3, arg4, arg5, arg6)
        }
        case syscall.SYS_SCHED_YIELD: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("sched yield syscall")
            }
        }
        case syscall.SYS_GETEUID32: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("get euid syscall")
            }
        }

        case syscall.SYS_GETUID32: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("get uid syscall")
            }
        }

        case syscall.SYS_GETEGID32: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("get egid syscall")
            }
        }

        case syscall.SYS_GETGID32: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("get gid syscall")
            }
        }
        case syscall.SYS_UNAME: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("uname syscall")
            }
            ret = linuxUnameSyscall(arg1)
        }
        case syscall.SYS_TGKILL: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("tgkill syscall")
            }

        }
        case syscall.SYS_MPROTECT: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("mprotect syscall")
            }

        }
        case syscall.SYS_SET_ROBUST_LIST: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("Set robust syscall")
            }
            ret = ^uint32(syscall.EINVAL)+1

        }
        case syscall.SYS_UGETRLIMIT: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("get upper r limit syscall")
            }
            ret = ^uint32(syscall.EINVAL)+1
        }
        // get random
        case 0x163: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("get random syscall")
            }
            ret = ^uint32(syscall.EINVAL)+1

        }
        // __NR_statx
        case 0x17f: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("statx syscall")
            }
            ret = linuxStatxSyscall(arg1, arg2, arg3, arg4, arg5)
        }
        // __NR_arch_prctl
        case 0x180: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("arch_prctl")
            }
            //printRegisters(currentInfo, currentThread.regs)
            ret = ^uint32(syscall.EINVAL)+1
            //ret = linuxSetThreadAreaSyscall()
        }

        // __NR_clock_gettime64
        case 0x193: {
            if PRINT_SYSCALL {
                printTid()
                text_mode_println("clock gettime 64 syscall")
            }
            ret = ^uint32(syscall.EINVAL)+1

        }

        default: unsupportedSyscall()
    }
    currentThread.regs.EAX = ret
}

func linuxExitGroupSyscall(status uint32) uint32 {
    //text_mode_print("Exiting domain ")
    //text_mode_print_hex(uint8(currentThread.domain.pid))
    //text_mode_println("")
    ExitDomain(currentThread.domain)
    PerformSchedule = true
    // Already in new context so return value from last syscall from current domain
    return currentThread.regs.EAX
}

func linuxStatxSyscall(dirfd uint32, path uint32, flags uint32, mask uint32, buf uint32, ) uint32 {
    addr, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(buf))
    if !ok {
        text_mode_print_errorln("invalid adress in statx")
        return ^uint32(syscall.EFAULT)+1
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
    return 0
}

func linuxUnameSyscall(buf uint32) uint32 {
    addr, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(buf))
    if !ok {
        text_mode_print_errorln("invalid adress in uname")
        return ^uint32(syscall.EFAULT)+1
    }
    provided := (*utsname)(unsafe.Pointer(addr))
    *provided = uts
    return 0
}

func linuxCloneSyscall(flags uint32, stack uint32) uint32 {
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
        return ^uint32(syscall.EINVAL)+1
    }
    // Need to make this better at some point
    newThreadMem := allocPage()
    Memclr(newThreadMem, PAGE_SIZE)
    newThread := (* thread)(unsafe.Pointer(newThreadMem))
    CreateNewThread(newThread, uintptr(stack), currentThread, currentThread.domain)

    currentThread.domain.AddThread(newThread)
    return newThread.tid
}

func linuxMincoreSyscall(addr uint32, length uint32, vec uint32) uint32 {
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
        return ^uint32(syscall.EFAULT)+1
    }

    arr := (*[30 << 1]byte)(unsafe.Pointer(vecAddr))[:(length+PAGE_SIZE-1) / PAGE_SIZE]
    for i := range arr {
        arr[i] = 1
    }
    return 0
}

func linuxMunmapSyscall(baseAddr uint32, length uint32) uint32 {
    //printRegisters(currentInfo, regs)
    for i:= uint32(0); i < length; i += PAGE_SIZE {
        addr := uintptr(baseAddr + i)

        if addr < KERNEL_RESERVED {
            return ^uint32(syscall.EINVAL)+1
        }
        currentThread.domain.MemorySpace.unMapPage(addr)
    }
    return 0
}

func linuxBrkSyscall(newBrk uint32) uint32 {
    brk := uint32(currentThread.domain.MemorySpace.Brk)
    if newBrk == 0 {
        return brk
    }
    if newBrk == brk || newBrk < brk {
        return brk
    }
    //text_mode_print_hex32(brk)
    for i:= (brk + PAGE_SIZE - 1) &^ (PAGE_SIZE - 1); i < newBrk; i+=PAGE_SIZE {
        p := allocPage()
        Memclr(p, PAGE_SIZE)
        flags := uint8(PAGE_PERM_USER | PAGE_RW)
        currentThread.domain.MemorySpace.mapPage(p, uintptr(i), flags)
    }
    currentThread.domain.MemorySpace.Brk = uintptr(newBrk)
    return newBrk
}

func linuxMmap2Syscall(target uint32, size uint32, prot uint32) uint32 {
    if target == 0 {
        target = uint32(currentThread.domain.MemorySpace.VmTop)
    }
    if prot == 0 {
        return uint32(currentThread.domain.MemorySpace.VmTop)
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
            return ^uint32(syscall.EINVAL)+1
        }
        currentThread.domain.MemorySpace.mapPage(p, uintptr(target+i), flags)
    }
    //text_mode_print_hex32(target)
    return target
}

func linuxSetThreadAreaSyscall(u_info uint32) uint32 {
    addr, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(u_info))
    if !ok {
        text_mode_print_errorln("Could not look up user desc")
        return ^uint32(syscall.EFAULT)+1
    }
    desc := (*UserDesc)(unsafe.Pointer(addr))
    if desc.Flags & UDESC_SEG_NOT_PRESENT != 0{
        text_mode_print_errorln("fixme: not handling updating entries")
        return  ^uint32(syscall.ENOSYS)+1
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
            return ^uint32(syscall.ESRCH)+1
        }
        desc.EntryNumber = slot
    }
    SetTlsSegment(slot, desc, currentThread.tlsSegments[:])

    return 0
}

func linuxOpenSyscall(path uint32, flags uint32) uint32 {
    _, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(flags))
    if !ok {
        return ^uint32(syscall.EFAULT)+1
    }
    //s := cstring(addr)
    //text_mode_println(s)
    return ^uint32(syscall.ENOSYS)+1
    //printRegisters(currentInfo, regs)

}

func linuxWriteVSyscall(fd uint32, arr uint32, count uint32) uint32 {
    if fd < 1 || fd > 2 {
        return ^uint32(syscall.EBADF)+1
    }

    addr, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(arr))
    if !ok {
        text_mode_print_errorln("Could not look up string addr")
        return ^uint32(syscall.EFAULT)+1
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
    return uint32(printed) //TODO: Return number of bytes written
}

func linuxWriteSyscall(fd uint32, text uint32, length uint32) uint32{
    if fd < 1 || fd > 2 {
        return ^uint32(syscall.EBADF)+1
    }

    addr, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(text))
    if !ok {
        text_mode_print_errorln("Could not look up string addr")
        return ^uint32(syscall.EFAULT)+1
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
    return uint32(len(s)) //TODO: nr of bytes written
}

func linuxReadSyscall(fd uint32, buf uint32, count uint32) uint32 {
    addr, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(buf))
    if !ok {
        text_mode_print_errorln("Could not look up read addr")
        return ^uint32(syscall.EFAULT)+1
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
    return num
}

func linuxFutexSyscall(uaddr uint32, futex_op uint32, val uint32, timeout uint32, uaddr2 uint32, val_3 uint32) uint32 {
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
        return ^uint32(syscall.ENOSYS)+1
    }

    addr, ok := currentThread.domain.MemorySpace.getPhysicalAddress(uintptr(uaddr))
    if !ok {
        text_mode_print_errorln("Could not look up read addr")
        return ^uint32(syscall.EFAULT)+1
    }
    futexAddr := (*uint32)(unsafe.Pointer(addr))
    switch (futex_op & 0xf) {
        case FUTEX_WAIT:
            if timeout != 0 {
                text_mode_print_error("Timeouts are not supported yet")
                return ^uint32(syscall.ENOSYS)+1
            }
            // This should be atomically
            if val != *futexAddr {
                return ^uint32(syscall.EAGAIN)+1
            }

            currentThread.isBlocked = true
            currentThread.waitAddress = futexAddr
            return 0
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
            return woken
        default:
            text_mode_print_error("Unsupported futex op")
            text_mode_print_hex32(futex_op)
            text_mode_println("")
            return ^uint32(syscall.ENOSYS)+1
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

func InitSyscall() {
    SetInterruptHandler(0x80, linuxSyscallHandler, KCS_SELECTOR, PRIV_USER)
}
