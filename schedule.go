package main

import (
    "unsafe"
)

type domain struct {
    next *domain
    prev *domain

    pid uint32
    numThreads uint32
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
        t.next = d.CurThread
        t.prev = d.CurThread.prev
        d.CurThread.prev.next = t
        d.CurThread.prev = t
    }
    t.tid = d.numThreads
    d.numThreads++
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

    // Currently ignored '^^ I don't have to do it thanks to spurious wakeups
    isBlocked bool
    waitAddress *uint32

    // Infos to stall a thread when switching
    info InterruptInfo
    regs RegisterState

    // TLS
    tlsSegments [GDT_ENTRIES]GdtEntry

    // fxsave and fxrstor need 512 bytes but 16 byte aligned
    // need space to force alignment if array not aligned
    fpOffset uintptr // should be between 0-15
    fpState [512+16]byte
}

var (
    currentDomain *domain
    largestPid uint32 = 0x0
)

func backupFpRegs(buffer uintptr)
func restoreFpRegs(buffer uintptr)

func EnqueueDomain(domain *domain) {
    if currentDomain == nil {
        currentDomain = domain
        domain.next = domain
        domain.prev = domain
    } else {
        domain.next = currentDomain
        domain.prev = currentDomain.prev
        currentDomain.prev.next = domain
        currentDomain.prev = domain
    }
    domain.pid = largestPid
    largestPid++
}

func DequeueDomain(d *domain) {
    if currentDomain == nil {
        return
    }
    if d == currentDomain {
        // special case
        if currentDomain == currentDomain.next {
            kernelPanic("Exiting last domain")
        }
        d.prev.next = d.next
        d.next.prev = d.prev
        switchToThread(currentDomain.next.CurThread.next)
        return
    }
    // Assume all domains form a circle
    for cur :=currentDomain.next; cur != currentDomain; cur = cur.next {
        if d == cur {
            cur.prev.next = cur.next
            cur.next.prev = cur.prev
            //TODO: Free memory
            return
        }
    }
}

func Schedule() {
    if currentDomain == nil {
        kernelPanic("No domain to schedule. :(")
        // does not return
    }

    if currentDomain.next == currentDomain && currentDomain.CurThread.next == currentDomain.CurThread {
        return
    }
    switchToThread(currentDomain.next.CurThread.next)

}

func switchToThread(t *thread) {
    // Save state of current thread
    addr := uintptr(unsafe.Pointer(&(currentDomain.CurThread.fpState)))
    offset := 16 - (addr % 16)
    currentDomain.CurThread.fpOffset = offset

    backupFpRegs(addr + offset)

    // Switch domain
    d := t.domain
    currentDomain = d

    // Load next thread

    d.CurThread = t

    currentInfo = &(d.CurThread.info)
    currentRegs = &(d.CurThread.regs)

    addr = uintptr(unsafe.Pointer(&(d.CurThread.fpState)))
    offset = d.CurThread.fpOffset
    if offset != 0xffffffff {
        if (addr + offset) % 16 != 0 {
            text_mode_print_hex32(uint32(addr))
            text_mode_print(" ")
            text_mode_print_hex32(uint32(offset))
            kernelPanic("Cannot restore FP state. Not aligned. Did array move?")
        }
        restoreFpRegs(addr + offset)
    }

    // Load TLS
    FlushTlsTable(t.tlsSegments[:])
}

func InitScheduling() {

}
