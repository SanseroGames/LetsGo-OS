package kernel

type Thread struct {
	Next *Thread
	prev *Thread

	Domain      *Domain
	Tid         uint32
	userStack   stack
	kernelStack stack
	isRemoved   bool
	// Currently ignored '^^ I don't have to do it thanks to spurious wakeups
	IsBlocked   bool
	WaitAddress *uint32

	// flag that shows that this thread would be handled as a new process in linux
	isFork bool

	// Infos to stall a thread when switching
	Info InterruptInfo
	Regs RegisterState

	kernelInfo InterruptInfo
	KernelRegs RegisterState

	IsKernelInterrupt    bool
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
	thread *Thread
}

func (l *threadList) Next() *Thread {
	if l.thread == nil {
		return nil
	}
	l.thread = l.thread.Next
	return l.thread
}

func (l *threadList) Enqueue(t *Thread) {
	if l.thread == nil {
		l.thread = t
		t.Next = t
		t.prev = t
	} else {
		t.Next = l.thread
		t.prev = l.thread.prev
		l.thread.prev.Next = t
		l.thread.prev = t
	}
}

func (l *threadList) Dequeue(t *Thread) {
	if t == l.thread && t.Next == t {
		l.thread = nil
	} else if t == l.thread {
		l.thread = t.Next
	}
	t.prev.Next = t.Next
	t.Next.prev = t.prev
	t.Next = nil
	t.prev = nil
}
