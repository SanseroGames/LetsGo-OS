package kernel

import "github.com/sanserogames/letsgo-os/kernel/mm"

type Domain struct {
	Next *Domain

	Pid        uint32
	numThreads uint32
	nextTid    uint32

	Segments    SegmentList
	MemorySpace mm.MemSpace

	runningThreads threadList
	blockedThreads threadList

	ProgramName string
}

func (d *Domain) AddThread(t *Thread) {
	t.Domain = d
	t.Next = nil
	t.prev = nil
	d.runningThreads.Enqueue(t)
	t.Tid = d.nextTid
	d.nextTid++
	d.numThreads++
}

func (d *Domain) RemoveThread(t *Thread) {
	if t.Domain != d {
		return
	}
	if t.IsBlocked {
		d.blockedThreads.Dequeue(t)
	} else {
		d.runningThreads.Dequeue(t)
	}
	t.isRemoved = true
	d.numThreads--
}

type domainList struct {
	Head *Domain
	tail *Domain
}

func (l *domainList) Append(domain *Domain) {
	if domain == nil {
		return
	}
	if l.Head == nil {
		l.Head = domain
		l.tail = l.Head
		l.Head.Next = l.tail
	} else {
		domain.Next = l.Head
		l.tail.Next = domain
		l.tail = domain
	}
	domain.Pid = largestPid
	largestPid++
}

func (l *domainList) Remove(d *Domain) {
	if d == l.Head {
		if d == l.tail {
			l.Head = nil
			l.tail = nil
		} else {
			l.Head = d.Next
			l.tail.Next = l.Head
		}
		return
	}
	for e := l.Head; e.Next != l.Head; e = e.Next {
		if e.Next == d {
			e.Next = d.Next
			if d == l.tail {
				l.tail = e
			}
			break
		}
	}
}
