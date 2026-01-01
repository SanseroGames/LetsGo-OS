package multiboot

import "unsafe"

const (
	MEM_MAP_AVAILABLE = 1
)

type MultibootInfo struct {
	TotalSize uint32
	reserved  uint32
}

type multibootType uint32

type MultibootTag struct {
	Type uint32
	Size uint32
}

type MultibootMemoryMap struct {
	MultibootTag // Type is 6
	EntrySize    uint32
	EntryVersion uint32
	Entries      MemoryMap // Take pointer of it and use it as slice with
}

// A module represents a module to be loaded along with the kernel.
type MultibootModule struct {
	MultibootTag
	// Start is the inclusive start of the Module memory location
	Start uint32

	// End is the exclusive end of the Module memory location.
	End uint32

	// Cmdline is a pointer to a null-terminated ASCII string.
	cmdline [100]byte
}

func (m *MultibootModule) Cmdline() string {
	if m.Size <= 17 {
		return ""
	}
	return unsafe.String(&m.cmdline[0], max(m.Size-16-1, 1))
}

type Module struct {
	Start   uint32
	End     uint32
	Cmdline string
}

type MemoryMap struct {
	BaseAddr uint64

	Length uint64

	Type uint32

	reserved uint32
}
