package mm

import (
	"iter"
	"syscall"
	"unsafe"

	"github.com/sanserogames/letsgo-os/kernel/log"
	"github.com/sanserogames/letsgo-os/kernel/panic"
)

const (

	// Reserve memory below 50 MB for kernel image
	KERNEL_START    = 1 << 20
	KERNEL_RESERVED = 50 << 20

	MAX_ALLOC_VIRT_ADDR = 0xf0000000
	MIN_ALLOC_VIRT_ADDR = 0x8000000
)

type MemSpace struct {
	PageDirectory *PageTable
	VmTop         uintptr
	Brk           uintptr
}

func (m *MemSpace) getPageTable(virtAddr uintptr) *PageTable {
	e := m.PageDirectory.GetEntry(virtAddr >> 10)

	if !e.IsPresent() {
		// No page table present
		addr := AllocPage()
		addr.Clear()
		// User perm here to allow all entries in page table to be possibly accessed by
		// user. User does not have access to page table though.
		m.PageDirectory.SetEntry(virtAddr>>10, addr.Address(), PAGE_PRESENT|PAGE_RW|PAGE_PERM_USER)

	}

	return e.AsPageTable()
}

func (m *MemSpace) TryMapPage(page uintptr, virtAddr uintptr, flags uint8) bool {

	pt := m.getPageTable(virtAddr)
	e := pt.GetEntry(virtAddr)
	if e.IsPresent() {
		return false
	}
	pt.SetEntry(virtAddr, page, uintptr(flags)|PAGE_PRESENT)
	if virtAddr >= m.VmTop && virtAddr < 0x8000000 {
		m.VmTop = virtAddr + PAGE_SIZE
	}
	return true
}

func (m *MemSpace) MapPage(page uintptr, virtAddr uintptr, flags uint8) {
	if PAGE_DEBUG {
		log.KDebugLn("[PAGE] Mapping page ", page, " to virt addr ", virtAddr)
	}
	if !m.TryMapPage(page, virtAddr, flags) {
		log.KErrorLn("Page already present")
		log.KPrintLn(page, " -> ", virtAddr)
		panic.KernelPanic("Tried to remap a page")
	}
}

func (m *MemSpace) UnmapPage(virtAddr uintptr) {
	if PAGE_DEBUG {
		log.KDebug("[PAGE] Unmapping page ", virtAddr)
	}
	pt := m.getPageTable(virtAddr)
	e := pt.GetEntry(virtAddr)
	if e.IsPresent() {
		e.UnsetPresent()
		FreePage(e.GetPhysicalAddress())
		if PAGE_DEBUG {
			log.KDebugLn("(phys-addr: ", e.GetPhysicalAddress(), ")")
		}
	} else {
		if PAGE_DEBUG {
			log.KDebugLn("\n[PAGE] WARNING: Page was already unmapped")
		}
	}
}

func (m *MemSpace) GetPageTableEntry(virtAddr uintptr) *PageTableEntry {
	pt := m.getPageTable(virtAddr)
	return pt.GetEntry(virtAddr)
}

func (m *MemSpace) FindSpaceFor(startAddr uintptr, length uintptr) uintptr {
	if PAGE_DEBUG {
		log.KDebugLn("[PAGE] Find space for ", startAddr, " with size ", length)
	}
	if startAddr < MIN_ALLOC_VIRT_ADDR {
		if PAGE_DEBUG {
			log.KDebugLn("[PAGE] startAddr was below MIN_ALLOC_VIRT_ADDR")
		}
		startAddr = MIN_ALLOC_VIRT_ADDR
	}
	for startAddr < MAX_ALLOC_VIRT_ADDR {
		// TODO: Check if page table is not allocated and count it as free instead of getting table entry and causing the table to be allocated
		for ; m.GetPageTableEntry(startAddr).IsPresent() && startAddr < MAX_ALLOC_VIRT_ADDR; startAddr += PAGE_SIZE {
		}
		endAddr := startAddr + length
		if PAGE_DEBUG {
			log.KDebugLn("[PAGE][FIND] Trying ", startAddr, " with endaddr ", endAddr)
		}
		if endAddr > MAX_ALLOC_VIRT_ADDR {
			break
		}
		isRangeFree := true
		for i := startAddr; i < endAddr && isRangeFree; i += PAGE_SIZE {
			entry := m.GetPageTableEntry(i)
			isRangeFree = isRangeFree && !entry.IsPresent()
			if !isRangeFree {
				startAddr = i
			}
		}
		if isRangeFree {
			if PAGE_DEBUG {
				log.KDebugLn("[PAGE][FIND] Found ", startAddr)
			}
			return startAddr
		} else {
			if PAGE_DEBUG {
				log.KDebugLn("[PAGE][FIND] position did not work")
			}
		}
	}
	if PAGE_DEBUG {
		log.KDebugLn("[PAGE][FIND] Did not find suitable location ", startAddr)
	}
	return 0
}

func (m *MemSpace) GetPhysicalAddress(virtAddr uintptr) (uintptr, bool) {
	if !m.IsAddressAccessible(virtAddr) {
		return 0, false
	} else {
		if PAGE_DEBUG {
			// log.KDebugln("[PAGING] Translated address: ", virtAddr, "->", uintptr(e&^(PAGE_SIZE-1)))
		}
		e := m.GetPageTableEntry(virtAddr)
		return uintptr(e.GetPhysicalAddress()) | (virtAddr & (PAGE_SIZE - 1)), true
	}
}

func (m *MemSpace) IsAddressAccessible(virtAddr uintptr) bool {
	e := m.GetPageTableEntry(virtAddr)
	return e.IsPresent() && e.IsUserAccessible()
}

func (m *MemSpace) IsRangeAccessible(startAddr uintptr, endAddr uintptr) bool {
	for pageAddr := startAddr &^ PAGE_SIZE; pageAddr < endAddr; pageAddr += PAGE_SIZE {
		if !m.IsAddressAccessible(pageAddr) {
			return false
		}
	}
	return true
}

func (m *MemSpace) FreeAllPages() {
	for tableIdx, tableEntry := range m.PageDirectory {
		if !tableEntry.IsPresent() {
			// tables that are not present don't need to be freed
			continue
		}
		pta := tableEntry.AsPageTable()
		if tableIdx > (KERNEL_RESERVED >> 22) {
			for i, entry := range pta {
				virtAddr := uintptr((tableIdx << 22) + (i << 12))
				if !entry.IsPresent() || virtAddr <= KERNEL_RESERVED {
					// Ignore kernel reserved mappings and tables that are not present
					continue
				}
				m.UnmapPage(virtAddr)
			}
		}
		tableEntry.UnsetPresent()
		FreePage(tableEntry.GetPhysicalAddress())
	}
	FreePage(unsafe.Pointer(m.PageDirectory))
}

func (m *MemSpace) ReadBytesFromUserSpace(startAddr uintptr, buffer []byte) syscall.Errno {

	count := 0
	for value, err := range m.IterateUserSpace(startAddr) {
		if count >= len(buffer) {
			break
		}
		if err != 0 {
			return err
		}
		buffer[count] = value
		count++
	}
	return 0
}

func (m *MemSpace) IterateUserSpace(startAt uintptr) iter.Seq2[byte, syscall.Errno] {
	return func(yield func(byte, syscall.Errno) bool) {
		if !m.IsAddressAccessible(startAt) {
			yield(0, syscall.EFAULT)
			return
		}

		startOffset := (startAt & (PAGE_SIZE - 1))

		for currentPage := startAt & (^uintptr(0) ^ (PAGE_SIZE - 1)); currentPage != 0; currentPage += PAGE_SIZE {
			physicalPage, accessible := m.GetPhysicalAddress(currentPage)
			if !(accessible) {
				yield(0, syscall.EFAULT)
				return
			}
			for currentOffset := startOffset; currentOffset < PAGE_SIZE; currentOffset++ {
				value := *(*byte)(unsafe.Pointer(physicalPage | currentOffset))
				if !yield(value, 0) {
					return
				}
			}
			startOffset = 0
		}

		yield(0, syscall.EFAULT)
	}
}

func IterateUserSpaceType[T any](startAt uintptr, memSpace *MemSpace) iter.Seq2[T, syscall.Errno] {
	return func(yield func(T, syscall.Errno) bool) {
		var size T
		var buf [100]byte
		valueSize := unsafe.Sizeof(size)
		if valueSize > uintptr(len(buf)) {
			panic.KernelPanic("IterateUserSpaceType: Used too big data structure to read from userspace (bigger as buf)")
		}
		index := 0
		for value, err := range memSpace.IterateUserSpace(startAt) {
			if err != 0 {
				yield(size, err)
				return
			}
			buf[index] = value
			index++
			if index == int(valueSize) {
				result := *(*T)(unsafe.Pointer(&buf))
				if !yield(result, 0) {
					return
				}
				index = 0
			}
		}

	}
}

func NewUserSpaceSplice[T any](memSpace *MemSpace, startAddr uintptr, len int) UserSpaceSlice[T] {
	return UserSpaceSlice[T]{
		memSpace:  memSpace,
		startAddr: startAddr,
		len:       len,
	}
}

type UserSpaceSlice[T any] struct {
	memSpace  *MemSpace
	startAddr uintptr
	len       int
}

func (s *UserSpaceSlice[T]) At(index int) (T, syscall.Errno) {
	var size T
	var buf [100]byte
	if index < 0 || index >= s.len {
		// TODO: Return error instead?
		log.KErrorLn("Index out of range: ", index, " ", s.len)
		panic.KernelPanic("UserSpaceSlice: Index out of range")
	}
	valueSize := unsafe.Sizeof(size)
	if valueSize > uintptr(len(buf)) {
		panic.KernelPanic("UserSlice: Used too big data structure to read from userspace (bigger as buf)")
	}

	s.memSpace.ReadBytesFromUserSpace(s.startAddr+uintptr(index*int(valueSize)), buf[:valueSize])
	// count := 0
	// for value, err := range s.memSpace.IterateUserSpace(s.startAddr + uintptr(index*int(valueSize))) {
	// 	if count >= int(valueSize) {
	// 		break
	// 	}
	// 	if err != 0 {
	// 		return size, err
	// 	}
	// 	buf[count] = value
	// 	count++
	// }

	// for v := range buf {
	// 	log.KDebug(v)
	// }
	// log.KDebugLn("")

	result := *(*T)(unsafe.Pointer(&buf))

	return result, 0

}

func (s *UserSpaceSlice[T]) Iterate() iter.Seq2[T, syscall.Errno] {
	return func(yield func(T, syscall.Errno) bool) {
		for i := 0; i < s.len; i++ {
			value, err := s.At(i)
			if !yield(value, err) || err != 0 {
				return
			}
		}
	}
}

// func (s *UserSpaceIterator[T]) Iterate() iter.Seq2[T, syscall.Errno] {
// 	return func(yield func(T, syscall.Errno) bool) {
// 		for i := 0; i < s.len; i++ {
// 			value, err := s.At(i)
// 			if !yield(value, err) || err != 0 {
// 				return
// 			}
// 		}
// 	}
// }
