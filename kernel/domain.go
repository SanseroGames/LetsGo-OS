package kernel

import "github.com/sanserogames/letsgo-os/kernel/mm"

type Domain struct {
	next *Domain

	pid        uint32
	numThreads uint32
	nextTid    uint32

	Segments    SegmentList
	MemorySpace mm.MemSpace

	runningThreads threadList
	blockedThreads threadList

	programName string
}

func (d *Domain) AddThread(t *Thread) {
	t.domain = d
	t.next = nil
	t.prev = nil
	d.runningThreads.Enqueue(t)
	t.tid = d.nextTid
	d.nextTid++
	d.numThreads++
}

func (d *Domain) RemoveThread(t *Thread) {
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
	head *Domain
	tail *Domain
}

func (l *domainList) Append(domain *Domain) {
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

func (l *domainList) Remove(d *Domain) {
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
