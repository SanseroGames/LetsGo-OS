package kernel

import (
	"unsafe"

	"github.com/sanserogames/letsgo-os/kernel/log"
)

type stack struct {
	hi uintptr
	lo uintptr
}

type taskswitchbuf struct {
	sp uintptr
}

var (
	currentThread  *thread    = &scheduleThread
	currentDomain  *domain    = nil
	allDomains     domainList = domainList{head: nil, tail: nil}
	largestPid     uint32     = 0x0
	kernelHlt      bool       = false
	scheduleThread thread     = thread{}
)

func backupFpRegs(buffer uintptr)
func restoreFpRegs(buffer uintptr)

func AddDomain(d *domain) {
	allDomains.Append(d)
	if ENABLE_DEBUG {
		log.KDebugLn("Added new domain with pid ", d.pid)
	}
	if currentDomain == nil || currentThread == nil {
		currentDomain = allDomains.head
		currentThread = currentDomain.runningThreads.thread
	}
}

func ExitDomain(d *domain) {
	allDomains.Remove(d)

	if allDomains.head == nil {
		currentDomain = nil
	}

	// clean up memory
	scheduleStackArg(func(dom uintptr) {
		doma := (*domain)(unsafe.Pointer(dom))
		cleanUpDomain(doma)
	}, (uintptr)(unsafe.Pointer(d)))
}

// Execute on scheduleStack
func cleanUpDomain(d *domain) {
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
	if ENABLE_DEBUG {
		log.KDebugLn("Allocated pages ", allocatedPages, " (out of", maxPages, ")")
	}
	Schedule()
	FreePage((uintptr)(unsafe.Pointer(d)))
}

// Execute on scheduleStack
func cleanUpThread(t *thread) {
	// TODO; Adjust when thread control block is no longer a single page
	threadPtr := (uintptr)(unsafe.Pointer(t))
	threadDomain := t.domain
	threadDomain.MemorySpace.UnmapPage(t.kernelStack.lo)
	threadDomain.MemorySpace.UnmapPage(threadPtr)
	if currentThread == t {
		currentThread = nil
	}
}

func ExitThread(t *thread) {
	if t.domain.numThreads <= 1 {
		// we're last thread
		ExitDomain(t.domain) // does not return
	}
	t.domain.RemoveThread(t)
	if ENABLE_DEBUG {
		log.KDebugLn("Removing thread ", t.tid, " from domain ", t.domain.pid)
	}
	scheduleStackArg(func(threadPtr uintptr) {
		thread := (*thread)(unsafe.Pointer(threadPtr))
		cleanUpThread(thread)
		Schedule()
	}, (uintptr)(unsafe.Pointer(t)))
}

func BlockThread(t *thread) {
	t.isBlocked = true
	t.domain.runningThreads.Dequeue(t)
	t.domain.blockedThreads.Enqueue(t)
	PerformSchedule = true
}

func getESP() uintptr
func waitForInterrupt()

func Block() {
	waitForInterrupt()
}

func ResumeThread(t *thread) {
	t.isBlocked = false
	t.domain.blockedThreads.Dequeue(t)
	t.domain.runningThreads.Enqueue(t)
}

func Schedule() {
	if currentDomain == nil {
		log.KErrorLn("No Domains to schedule")
		Shutdown()
		// DisableInterrupts()
		// Hlt()
	}
	//log.KDebug("Scheduling in ")
	//printTid(defaultLogWriter, currentThread)
	nextDomain := currentDomain.next
	newThread := nextDomain.runningThreads.Next()
	if newThread == nil {
		for newDomain := nextDomain.next; newDomain != nextDomain; newDomain = newDomain.next {
			newThread = newDomain.runningThreads.Next()
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
	currentDomain = nextDomain
	if newThread == currentThread {
		return
	}

	switchToThread(newThread)

	//log.KDebug("Now executing: ")
	//printTid(defaultLogWriter, currentThread)
}

func switchToThread(t *thread) {
	// Save state of current thread
	if currentThread != nil {
		addr := uintptr(unsafe.Pointer(&(currentThread.fpState)))
		offset := 16 - (addr % 16)
		currentThread.fpOffset = offset

		backupFpRegs(addr + offset)
	}

	// Load next thread
	//log.KDebugln("Switching to domain pid", currentDomain.pid, " and thread ", t.tid)
	currentThread = t

	addr := uintptr(unsafe.Pointer(&(currentThread.fpState)))
	offset := currentThread.fpOffset
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
