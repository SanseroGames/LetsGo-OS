package main

import (
    "unsafe"
)

const (
    // Reserve memory below 50 MB for kernel image
    KERNEL_START = 1 << 20
    KERNEL_RESERVED = 50 << 20
    PAGE_SIZE = 4 << 10
    ENTRIES_PER_TABLE = PAGE_SIZE/4

    PAGE_PRESENT = 1 << 0
    PAGE_RW = 1 << 1
    PAGE_PERM_USER = 1 << 2
    PAGE_PERM_KERNEL = 0 << 2
    PAGE_WRITETHROUGH = 1 << 3
    PAGE_DISABLE_CACHE = 1 << 4
)

var (
    freePagesList *page
    kernelMemSpace MemSpace = MemSpace{}
)

type PageTable [ENTRIES_PER_TABLE]pageTableEntry

type pageTableEntry uintptr

type page struct {
    next* page
}

type MemSpace struct {
    PageDirectory *PageTable
    VmTop uintptr
    Brk uintptr
}

func enablePaging()

func switchPageDir(dir *PageTable)

func getCurrentPageDir() *PageTable

func getPageFaultAddr() uint32

func pageFaultHandler() {
    text_mode_print_char(0xa)
    text_mode_print_errorln("Page Fault! Disabling Interrupt and halting!")
    text_mode_print_char(0xa)
    text_mode_print("Domain ID:")
    text_mode_print_hex32(currentThread.domain.pid)
    text_mode_println("")
    text_mode_print("Thread ID:")
    text_mode_print_hex(uint8(currentThread.tid))
    text_mode_print_char(0xa)
    text_mode_print("Exception code: ")
    text_mode_print_hex32(currentThread.info.ExceptionCode)
    text_mode_print_char(0xa)
    text_mode_print("EIP: ")
    text_mode_print_hex32(currentThread.info.EIP)
    text_mode_println("")
    text_mode_print("Causing Address: ")
    text_mode_print_hex32(getPageFaultAddr())
    DisableInterrupts()
    Hlt()
}

func Memclr(p uintptr, n int) {
	s := (*(*[1 << 30]byte)(unsafe.Pointer(p)))[:n]
	// the compiler will emit runtime.memclrNoHeapPointers
	for i := range s {
		s[i] = 0
	}
}

func freePage(addr uintptr) {
    if addr % PAGE_SIZE != 0 {
        return
    }
    p := (*page)(unsafe.Pointer(addr))
    p.next = freePagesList
    freePagesList = p
}

func allocPage() uintptr {
    if freePagesList == nil {
        kernelPanic("[PAGE] Out of pages to allocate")
    }
    p := freePagesList
    freePagesList = p.next
    return uintptr(unsafe.Pointer(p))
}

func createNewPageDirectory() MemSpace {
    var ret MemSpace
    ret.PageDirectory = (*PageTable)(unsafe.Pointer(allocPage()))
    Memclr(uintptr(unsafe.Pointer(ret.PageDirectory)), PAGE_SIZE)
    for i:=uint64(KERNEL_START); i < KERNEL_RESERVED; i+=PAGE_SIZE {
        ret.tryMapPage(uintptr(i), uintptr(i), PAGE_RW | PAGE_PERM_KERNEL)
    }
    return ret
}

// Would be better to return addres when multitasking
func (m *MemSpace) getPageTable(page uintptr) pageTableEntry {
    directoryIndex := page >> 22
    e := m.PageDirectory[directoryIndex]
    if e & PAGE_PRESENT == 1{
        return e
    }
    // No pagetable present
    addr := allocPage()
    Memclr(addr, PAGE_SIZE)
    e = pageTableEntry(addr | PAGE_PRESENT | PAGE_RW | PAGE_PERM_USER)
    m.PageDirectory[directoryIndex] = e
    return e
}

func (m *MemSpace) tryMapPage(page uintptr, virtAddr uintptr, flags uint8) bool {
    pt := m.getPageTable(virtAddr)
    pta := (*PageTable)(unsafe.Pointer(pt &^ ((1 << 12) - 1)))
    pageIndex := virtAddr >> 12 & ((1 << 10) - 1)
    e := pta[pageIndex]
    if e & PAGE_PRESENT == 1 {
        return false
    }
    e = pageTableEntry(page | uintptr(flags) | PAGE_PRESENT)
    pta[pageIndex] = e
    if virtAddr >= m.VmTop && virtAddr < 0x80000000 {
        m.VmTop = virtAddr+PAGE_SIZE
    }
    return true
}

func (m *MemSpace) mapPage(page uintptr, virtAddr uintptr, flags uint8) {
        //text_mode_print("Mapping ")
        //text_mode_print_hex32(uint32(page))
        //text_mode_print(" -> ")
        //text_mode_print_hex32(uint32(virtAddr))
        //text_mode_println("")
    if !m.tryMapPage(page, virtAddr, flags) {
        text_mode_print_errorln("Page already present")
        text_mode_print_hex32(uint32(page))
        text_mode_print(" -> ")
        text_mode_print_hex32(uint32(virtAddr))
        text_mode_println("")
        kernelPanic("Tried to remap a page")
    }
}

func (m *MemSpace) unMapPage(virtAddr uintptr) {
    pt := m.getPageTable(virtAddr)
    pta := (*PageTable)(unsafe.Pointer(pt &^ ((1 << 12) - 1)))
    pageIndex := virtAddr >> 12 & ((1 << 10) - 1)
    e := pta[pageIndex]
    if e & PAGE_PRESENT == 1 {
        phAddr := uintptr(e &^ (PAGE_SIZE -1)) | (virtAddr & (PAGE_SIZE -1))
        pta[pageIndex] = e &^ PAGE_PRESENT
        freePage(phAddr)
    }
}

func (m *MemSpace) getPhysicalAddress(virtAddr uintptr) (uintptr, bool) {
    pageAddr := virtAddr &^ (PAGE_SIZE -1)
    pt := m.getPageTable(pageAddr)
    pta := (*PageTable)(unsafe.Pointer(pt &^ ((1 << 12) - 1)))
    pageIndex := pageAddr >> 12 & ((1 << 10) - 1)
    e := pta[pageIndex]
    if e & PAGE_PRESENT == 0 {
        return 0, false
    } else {
        return uintptr(e &^ (PAGE_SIZE -1)) | (virtAddr & (PAGE_SIZE -1)),true
    }
}

func InitPaging() {
    var topAddr uint64 = 0
    for _,p := range memoryMaps {
        if p.BaseAddr + p.Length > topAddr && p.BaseAddr + p.Length < 0x10000000 {
            topAddr = p.BaseAddr + p.Length
        }
        if p.Type != MEM_MAP_AVAILABLE {
            continue
        }
        if p.Length < PAGE_SIZE {
            continue
        }

        if p.BaseAddr + p.Length < KERNEL_RESERVED {
            continue
        }

        startAddr := uintptr(p.BaseAddr)
        if startAddr % PAGE_SIZE != 0 {
            startAddr = (startAddr + PAGE_SIZE - 1) &^ (PAGE_SIZE - 1)
        }
        if startAddr < KERNEL_RESERVED {
            startAddr = KERNEL_RESERVED
        }
        for i := startAddr; i < uintptr(p.BaseAddr + p.Length); i += PAGE_SIZE {
            freePage(i)
        }
    }

    kernelPageDirectory := (*PageTable)(unsafe.Pointer(allocPage()))
    Memclr(uintptr(unsafe.Pointer(kernelPageDirectory)), PAGE_SIZE)
    kernelMemSpace.PageDirectory = kernelPageDirectory
    //for _,p := range memoryMaps {
    //    startAddr := p.BaseAddr &^ (PAGE_SIZE - 1)
    //    lastAddr := p.BaseAddr + p.Length - 1
    //    for i:= startAddr; i < lastAddr; i += PAGE_SIZE {
    //        tryMapPage(uintptr(i), uintptr(i), PAGE_RW | PAGE_PERM_KERNEL , kernelPageDirectory)
    //    }
    //}
    //text_mode_print_hex32(uint32(topAddr))
    //text_mode_println("")
    //text_mode_print_hex32(uint32(uintptr(unsafe.Pointer(kernelPageDirectory))))
    //text_mode_println("")
    for i:=uint64(0); i < topAddr; i+=PAGE_SIZE {
        flag := uint8(PAGE_PERM_KERNEL)
        // TODO: Get address somewhere
        if i < 0x100000 || i >= 0x199000 {
            flag |= PAGE_RW
        }
        kernelMemSpace.tryMapPage(uintptr(i), uintptr(i), flag)
    }
    switchPageDir(kernelPageDirectory)
    SetInterruptHandler(0xE, pageFaultHandler, KCS_SELECTOR, PRIV_USER)
    enablePaging()
    //printPageTable((*PageTable)(unsafe.Pointer(kernelPageDirectory[0] &^ ((1 << 12) - 1))))
}

func printPageTable(table *PageTable, start, length int) {
    // length of 192 fills the screen
    for i, n := range table[start:start+length] {
        text_mode_print_hex32(uint32(n))
        if i % 8 == 7 {
            text_mode_println("")
        } else {
            text_mode_print(" ")
        }
    }
}
