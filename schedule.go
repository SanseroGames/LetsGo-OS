package main

import (
    "unsafe"
)

type domain struct {
    next *domain
    prev *domain

    Segments SegmentList
    MemorySpace MemSpace
    CurThread *thread
}

func (d *domain) EnqueueThread(t *thread) {
    t.domain = d
    if d.CurThread == nil {
        d.CurThread = t
        t.next = t
        t.prev = t
    } else {
        t.next = d.CurThread.next
        t.prev = d.CurThread
        d.CurThread.next.prev = t
        d.CurThread.next = t
    }
}

func (d *domain) DequeueCurrentThread() {
    if d.CurThread == nil {
        DequeueDomain(d)
        return
    }
    if d.CurThread == d.CurThread.next {
        // We should be the last thread alive
        DequeueDomain(d)
        return
    }
    d.CurThread.prev.next = d.CurThread.next
    d.CurThread.next.prev = d.CurThread.prev
    // TODO: Free memory from thread
    d.CurThread = d.CurThread.next
}

type thread struct {
    next *thread
    prev *thread

    domain *domain
    tid uint32
    StackStart uintptr

    // Infos to stall a thread when switching
    info InterruptInfo
    regs RegisterState

    // fxsave and fxrstor need 512 bytes but 16 byte aligned
    // need space to force alignment if array not aligned
    fpState [528]byte
    fpOffset uintptr // should be between 0-15
}

var (
    currentDomain *domain
)

func backupFpRegs(buffer uintptr)
func restoreFpRegs(buffer uintptr)

func EnqueueDomain(domain *domain) {
    if currentDomain == nil {
        currentDomain = domain
        domain.next = domain
        domain.prev = domain
    } else {
        domain.next = currentDomain.next
        domain.prev = currentDomain
        currentDomain.next.prev = domain
        currentDomain.next = domain
    }
}

func DequeueDomain(d *domain) {
    if currentDomain == nil {
        return
    }
    if d == currentDomain {
        // special case
        return
    }
    cur := currentDomain
    if d == cur {
        cur.prev.next = cur.next
        cur.next.prev = cur.prev
        return
    }
}

func Schedule(info *InterruptInfo, regs *RegisterState) {
    if currentDomain == nil {
        kernelPanic("No domain to schedule. :(")
        // does not return
    }
    if currentDomain.next == currentDomain && currentDomain.CurThread.next == currentDomain.CurThread {
        return
    }
    currentDomain.CurThread.info = *info
    currentDomain.CurThread.regs = *regs

    addr := uintptr(unsafe.Pointer(&(currentDomain.CurThread.fpState)))
    offset := 16 - addr % 16
    currentDomain.CurThread.fpOffset = offset

    backupFpRegs(addr + offset)

    currentDomain.CurThread = currentDomain.CurThread.next

    currentDomain = currentDomain.next
    *info = currentDomain.CurThread.info
    *regs = currentDomain.CurThread.regs
    addr = uintptr(unsafe.Pointer(&(currentDomain.CurThread.fpState)))
    offset = currentDomain.CurThread.fpOffset
    if offset != 0xffffffff {
        if (addr + offset) % 16 != 0 {
            text_mode_print_hex32(uint32(addr))
            text_mode_print(" ")
            text_mode_print_hex32(uint32(offset))
            kernelPanic("Cannot restore FP state. Not aligned. Did array move?")
        }
        restoreFpRegs(addr + offset)
    }

}

func InitScheduling() {

}
