package kernel

type thread struct {
	next *thread
	prev *thread

	domain      *domain
	tid         uint32
	userStack   stack
	kernelStack stack
	isRemoved   bool
	// Currently ignored '^^ I don't have to do it thanks to spurious wakeups
	isBlocked   bool
	waitAddress *uint32

	// flag that shows that this thread would be handled as a new process in linux
	isFork bool

	// Infos to stall a thread when switching
	info InterruptInfo
	regs RegisterState

	kernelInfo InterruptInfo
	kernelRegs RegisterState

	isKernelInterrupt    bool
	interruptedKernelEIP uintptr
	interruptedKernelESP uint32

	// TLS
	tlsSegments [GDT_ENTRIES]GdtEntry

	// fxsave and fxrstor need 512 bytes but 16 byte aligned
	// need space to force alignment if array not aligned
	fpOffset uintptr // should be between 0-15
	fpState  [512 + 16]byte
}

type threadList struct {
	thread *thread
}

func (l *threadList) Next() *thread {
	if l.thread == nil {
		return nil
	}
	l.thread = l.thread.next
	return l.thread
}

func (l *threadList) Enqueue(t *thread) {
	if l.thread == nil {
		l.thread = t
		t.next = t
		t.prev = t
	} else {
		t.next = l.thread
		t.prev = l.thread.prev
		l.thread.prev.next = t
		l.thread.prev = t
	}
}

func (l *threadList) Dequeue(t *thread) {
	if t == l.thread && t.next == t {
		l.thread = nil
	} else if t == l.thread {
		l.thread = t.next
	}
	t.prev.next = t.next
	t.next.prev = t.prev
	t.next = nil
	t.prev = nil
}
