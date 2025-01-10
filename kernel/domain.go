package kernel

type domain struct {
	next *domain

	pid        uint32
	numThreads uint32
	nextTid    uint32

	Segments    SegmentList
	MemorySpace MemSpace

	runningThreads threadList
	blockedThreads threadList

	programName string
}

func (d *domain) AddThread(t *thread) {
	t.domain = d
	t.next = nil
	t.prev = nil
	d.runningThreads.Enqueue(t)
	t.tid = d.nextTid
	d.nextTid++
	d.numThreads++
}

func (d *domain) RemoveThread(t *thread) {
	if t.domain != d {
		return
	}
	if t.isBlocked {
		d.blockedThreads.Dequeue(t)
	} else {
		d.runningThreads.Dequeue(t)
	}
	t.isRemoved = true
	d.numThreads--
}

type domainList struct {
	head *domain
	tail *domain
}

func (l *domainList) Append(domain *domain) {
	if domain == nil {
		return
	}
	if l.head == nil {
		l.head = domain
		l.tail = l.head
		l.head.next = l.tail
	} else {
		domain.next = l.head
		l.tail.next = domain
		l.tail = domain
	}
	domain.pid = largestPid
	largestPid++
}

func (l *domainList) Remove(d *domain) {
	if d == l.head {
		if d == l.tail {
			l.head = nil
			l.tail = nil
		} else {
			l.head = d.next
			l.tail.next = l.head
		}
		return
	}
	for e := l.head; e.next != l.head; e = e.next {
		if e.next == d {
			e.next = d.next
			if d == l.tail {
				l.tail = e
			}
			break
		}
	}
}
