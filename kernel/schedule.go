package kernel

import (
	"unsafe"

	"github.com/sanserogames/letsgo-os/kernel/log"
	"github.com/sanserogames/letsgo-os/kernel/mm"
	"github.com/sanserogames/letsgo-os/kernel/utils"
)

type stack struct {
	hi uintptr
	lo uintptr
}

type taskswitchbuf struct {
	sp uintptr
}

var (
	CurrentThread  *Thread    = &scheduleThread
	CurrentDomain  *Domain    = nil
	allDomains     domainList = domainList{head: nil, tail: nil}
	largestPid     uint32     = 0x0
	kernelHlt      bool       = false
	scheduleThread Thread     = Thread{}
)

func backupFpRegs(buffer uintptr)
func restoreFpRegs(buffer uintptr)

func AddDomain(d *Domain) {
	allDomains.Append(d)
	if ENABLE_DEBUG {
		log.KDebugLn("Added new domain with pid ", d.Pid)
	}
	if CurrentDomain == nil || CurrentThread == nil {
		CurrentDomain = allDomains.head
		CurrentThread = CurrentDomain.runningThreads.thread
	}
}

func ExitDomain(d *Domain) {
	allDomains.Remove(d)

	if allDomains.head == nil {
		CurrentDomain = nil
	}

	// clean up memory
	scheduleStackArg(func(dom uintptr) {
		doma := utils.UIntToPointer[Domain](dom)
		cleanUpDomain(doma)
	}, (uintptr)(unsafe.Pointer(d)))
}

// Execute on scheduleStack
func cleanUpDomain(d *Domain) {
	// Clean up threads
	for cur := d.runningThreads.thread; d.runningThreads.thread != nil; cur = d.runningThreads.thread {
		//log.KDebugln("Clean up thread ", cur.tid)
		//log.KDebugln("t:", (uintptr)(unsafe.Pointer(cur)), " t.n:", (uintptr)(unsafe.Pointer(cur.next)), " t.p:", (uintptr)(unsafe.Pointer(cur.prev)))
		d.runningThreads.Dequeue(cur)
		cleanUpThread(cur)
	}
	for cur := d.blockedThreads.thread; d.blockedThreads.thread != nil; cur = d.blockedThreads.thread {
		cleanUpThread(cur)
		d.blockedThreads.Dequeue(cur)
	}
	// Clean allocated memory
	d.MemorySpace.FreeAllPages()

	// Clean up kernel resources
	Schedule()
	mm.FreePage(unsafe.Pointer(d))
}

// Execute on scheduleStack
func cleanUpThread(t *Thread) {
	// TODO; Adjust when thread control block is no longer a single page
	threadPtr := (uintptr)(unsafe.Pointer(t))
	threadDomain := t.Domain
	threadDomain.MemorySpace.UnmapPage(t.kernelStack.lo)
	threadDomain.MemorySpace.UnmapPage(threadPtr)
	if CurrentThread == t {
		CurrentThread = nil
	}
}

func ExitThread(t *Thread) {
	if t.Domain.numThreads <= 1 {
		// we're last thread
		ExitDomain(t.Domain) // does not return
	}
	t.Domain.RemoveThread(t)
	if ENABLE_DEBUG {
		log.KDebugLn("Removing thread ", t.Tid, " from domain ", t.Domain.Pid)
	}
	scheduleStackArg(func(threadPtr uintptr) {
		thread := utils.UIntToPointer[Thread](threadPtr)
		cleanUpThread(thread)
		Schedule()
	}, (uintptr)(unsafe.Pointer(t)))
}

func waitForInterrupt()

/*
 * Yield execution but allow thread to be scheduled again.
 */
func Yield() {
	waitForInterrupt()
}

/*
 * Yield execution and prevent thread from being scheduled
 */
func Block() {
	CurrentThread.IsBlocked = true
	for CurrentThread.IsBlocked {
		waitForInterrupt()
	}
}

func ResumeThread(t *Thread) {
	t.IsBlocked = false
	t.Domain.blockedThreads.Dequeue(t)
	t.Domain.runningThreads.Enqueue(t)
}

func getNextUnblockedThread(domain *Domain) *Thread {
	newThread := domain.runningThreads.Next()
	if newThread == nil || !newThread.IsBlocked {
		return newThread
	}
	for cur := domain.runningThreads.Next(); cur != newThread; cur = domain.runningThreads.Next() {
		if !cur.IsBlocked {
			return cur
		}
	}
	return nil
}

func Schedule() {
	if CurrentDomain == nil {
		log.KErrorLn("No Domains to schedule")
		Shutdown()
		// DisableInterrupts()
		// Hlt()
	}
	//log.KDebug("Scheduling in ")
	//printTid(defaultLogWriter, currentThread)
	nextDomain := CurrentDomain.next
	newThread := getNextUnblockedThread(nextDomain)

	if newThread == nil {
		for newDomain := nextDomain.next; newDomain != nextDomain; newDomain = newDomain.next {
			newThread = getNextUnblockedThread(newDomain)
			if newThread != nil {
				break
			}
		}
	}
	//if currentThread.next == currentThread && currentThread.tid != 0{
	//    kernelPanic("Why no next?")
	//}
	//if newThread.tid == currentThread.tid && currentThread.tid != 0 {
	//    kernelPanic("Why only one thread?")
	//}
	//log.KPrintln("next domain: ", nextDomain.pid)
	//log.KPrintln("next thread: ", newThread.tid)
	//log.KPrintln("domain threads: ", nextDomain.numThreads)

	if newThread == nil {
		if kernelHlt {
			// We are already stalling the kernel
			return
		}
		// All threads blocked or no threads exist anymore.
		kernelHlt = true
		PerformSchedule = false
		//currentThread = nil
		kernelPanic("test")
		return
	}

	kernelHlt = false
	CurrentDomain = nextDomain
	if newThread == CurrentThread {
		return
	}

	switchToThread(newThread)

	//log.KDebug("Now executing: ")
	//printTid(defaultLogWriter, currentThread)
}

func switchToThread(t *Thread) {
	// Save state of current thread
	if CurrentThread != nil {
		addr := uintptr(unsafe.Pointer(&(CurrentThread.fpState)))
		offset := 16 - (addr % 16)
		CurrentThread.fpOffset = offset

		backupFpRegs(addr + offset)
	}

	// Load next thread
	//log.KDebugln("Switching to domain pid", currentDomain.pid, " and thread ", t.tid)
	CurrentThread = t

	addr := uintptr(unsafe.Pointer(&(CurrentThread.fpState)))
	offset := CurrentThread.fpOffset
	if offset != 0xffffffff {
		if (addr+offset)%16 != 0 {
			log.KPrintLn(addr, " ", offset)
			kernelPanic("Cannot restore FP state. Not aligned. Did array move?")
		}
		restoreFpRegs(addr + offset)
	}

	// Load TLS
	FlushTlsTable(t.tlsSegments[:])
}

func InitScheduling() {

}
