package main

import (
    "unsafe"
    "reflect"
    "syscall"
)

const (

)

func linuxSyscallHandler(info *InterruptInfo, regs *RegisterState){
    switch (regs.EAX) {
    case syscall.SYS_WRITE: {
        // Linux write syscall
        linuxWriteSyscall(info, regs)
    }
    case syscall.SYS_SET_THREAD_AREA: {
        linuxSetThreadAreaSyscall(info, regs)
    }
    case syscall.SYS_OPEN: {
        linuxOpenSyscall(info, regs)
    }
    case syscall.SYS_CLOSE: {
        text_mode_println("close syscall")
    }
    case syscall.SYS_READ: {
        linuxReadSyscall(info, regs)
    }
    case syscall.SYS_SCHED_GETAFFINITY: {
        text_mode_println("sched getaffinity syscall")
    }
    case syscall.SYS_NANOSLEEP: {
        //text_mode_println("nanosleep syscall")
    }
    case syscall.SYS_EXIT_GROUP: {
        text_mode_println("exit group syscall")
    }
    case syscall.SYS_BRK:
    case syscall.SYS_MMAP2: {
        linuxMmap2Syscall(info, regs)
    }
    case syscall.SYS_MINCORE: {
        linuxMincoreSyscall(info, regs)
    }
    case syscall.SYS_MUNMAP: {
        linuxMunmapSyscall(info, regs)
    }
    case syscall.SYS_CLOCK_GETTIME: {
        regs.EAX = 0
    }
    case syscall.SYS_RT_SIGPROCMASK: {
        regs.EAX = 0
    }
    case syscall.SYS_SIGALTSTACK: {
        regs.EAX = 0
    }
    case syscall.SYS_RT_SIGACTION: {
        regs.EAX = 0
    }
    case syscall.SYS_GETTID: {
        regs.EAX = 0
    }
    case syscall.SYS_CLONE: {
        regs.EAX = ^uint32(syscall.EAGAIN)+1
    }
    case syscall.SYS_FUTEX: {
        regs.EAX = 0xffffffff
    }
    default: unsupportedSyscall(info, regs)
    }
}

func linuxMincoreSyscall(info *InterruptInfo, regs *RegisterState) {
    text_mode_print("mincore syscall")
    text_mode_print(" addr: ")
    text_mode_print_hex32(regs.EBX)
    text_mode_print(" length: ")
    text_mode_print_hex32(regs.ECX)
    text_mode_print(" vec: ")
    text_mode_print_hex32(regs.EDX)
    text_mode_println("")
    //addr := regs.EBX
    length := regs.ECX
    vec, ok := currentTask.getPhysicalAddress(uintptr(regs.EDX))
    if !ok {
        text_mode_print_errorln("Could not look up vec array")
        regs.EAX = ^uint32(syscall.EFAULT)+1
        return
    }

    arr := (*[30 << 1]byte)(unsafe.Pointer(vec))[:(length+PAGE_SIZE-1) / PAGE_SIZE]
    for i := range arr {
        arr[i] = 1
    }
}

func linuxMunmapSyscall(info *InterruptInfo, regs *RegisterState) {
    text_mode_println("munmap syscall")
    for i:= uint32(0); i < regs.ECX; i += PAGE_SIZE {
        currentTask.unMapPage(uintptr(regs.EBX + i))
    }
}

func linuxMmap2Syscall(info *InterruptInfo, regs *RegisterState) {
    text_mode_println("mmap2 syscall")
    target := regs.EBX
    size := regs.ECX
    prot := regs.EDX
    if target == 0 {
        target = uint32(currentTask.VmTop)
    }
    if prot == 0 {
        regs.EAX = uint32(currentTask.VmTop)
        return
    }
    text_mode_print("vmtop: ")
    text_mode_print_hex32(uint32(currentTask.VmTop))
    text_mode_print(" target: ")
    text_mode_print_hex32(regs.EBX)
    text_mode_print(" size: ")
    text_mode_print_hex32(size)
    text_mode_print(" prot: ")
    text_mode_print_hex32(regs.EDX)
    text_mode_print(" flags: ")
    text_mode_print_hex32(regs.ESI)
    text_mode_println("")
    for i:=uint32(0); i < size; i+=PAGE_SIZE {
        p := allocPage()
        Memclr(p, PAGE_SIZE)
        flags := uint8(PAGE_PERM_USER | PAGE_RW)
        currentTask.mapPage(p, uintptr(target+i), flags)
    }
    regs.EAX = target
}

func linuxSetThreadAreaSyscall(info *InterruptInfo, regs *RegisterState) {
    text_mode_println("set thread area")
    addr, ok := currentTask.getPhysicalAddress(uintptr(regs.EBX))
    if !ok {
        text_mode_print_errorln("Could not look up user desc")
        regs.EAX = ^uint32(syscall.EFAULT)+1
        return
    }
    desc := (*UserDesc)(unsafe.Pointer(addr))
    if desc.Flags & UDESC_SEG_NOT_PRESENT != 0{
        text_mode_print_errorln("fixme: not handling updating entries")
        return
    }

    if desc.EntryNumber != 0xffffffff {
        text_mode_print_errorln("fixme: not handling set entry number at the moment")
        return
    }
    flags := uint8(SEG_GRAN_BYTE)
    if desc.Flags & UDESC_LIMIT_IN_PAGES != 0 {
        flags = SEG_GRAN_4K_PAGE
    }
    access := uint8(PRIV_USER | SEG_NORMAL)
    if desc.Flags & UDESC_RX_ONLY != 0{
        access |= SEG_EXEC
    } else {
        access |= SEG_W
    }

    slot := AddSegment(uintptr(desc.BaseAddr), uintptr(desc.Limit), access,flags)
    desc.EntryNumber = uint32(slot)
    UpdateGdt()

    regs.EAX = 0
}

func linuxOpenSyscall(info *InterruptInfo, regs *RegisterState) {
    text_mode_println("open syscall")
    addr, ok := currentTask.getPhysicalAddress(uintptr(regs.EBX))
    if !ok {
        return
    }
    s := cstring(addr)
    text_mode_println(s)
    if s == "/sys/kernel/mm/transparent_hugepage/hpage_pmd_size" && false {
        text_mode_println("reading page size")
        regs.EAX = 0x42
        //kernelPanic()
    } else {
        regs.EAX = ^uint32(syscall.ENOSYS)+1
    }
    //printRegisters(info, regs)

}

var stop int = 0

func linuxWriteSyscall(info *InterruptInfo, regs *RegisterState){
    // Not safe
    addr, ok := currentTask.getPhysicalAddress(uintptr(regs.ECX))
    if !ok {
        text_mode_print_errorln("Could not look up string addr")
        return
    }
    var s string
    hdr := (*reflect.StringHeader)(unsafe.Pointer(&s)) // case 1
    hdr.Data = uintptr(unsafe.Pointer(uintptr(addr))) // case 6 (this case)
    hdr.Len = int(regs.EDX)
    // Test if it is a go panic and print infos to debug:

    //prevent stack trace
    if stop > 30 {
        return
    }
    stop++


    if len(s) >= 6 && s[0:6] == "panic:"{
        text_mode_print_errorln("panic:")
        printRegisters(info, regs)
        text_mode_println("")
    } else {
        text_mode_print(s)
    }
}

func linuxReadSyscall(info *InterruptInfo, regs *RegisterState) {
    text_mode_println("read syscall")
    addr, ok := currentTask.getPhysicalAddress(uintptr(regs.ECX))
    if !ok {
        text_mode_print_errorln("Could not look up read addr")
        return
    }
    if regs.EBX == 0x42 {
        arr := (*[1 <<30]byte)(unsafe.Pointer(addr))[:regs.EDX]
        arr[0] = '4'
        arr[1] = '0'
        arr[2] = '9'
        arr[3] = '6'
        regs.EAX = 4
    }
}

func unsupportedSyscall(info *InterruptInfo, regs *RegisterState){
    text_mode_print_char(0xa)
    text_mode_print_errorln("Unsupported Linux Syscall received! Disabling Interrupts and halting")
    panicHelper(info, regs)
}

func InitSyscall() {
    SetInterruptHandler(0x80, linuxSyscallHandler, KCS_SELECTOR, PRIV_USER)
}
