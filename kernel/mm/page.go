package mm

import (
	"unsafe"

	"github.com/sanserogames/letsgo-os/kernel/log"
	"github.com/sanserogames/letsgo-os/kernel/panic"
	"github.com/sanserogames/letsgo-os/kernel/utils"
)

type page struct {
	next *page
}

type Page []byte

func (page Page) Clear() {
	clear(page)
}

func (page Page) Address() uintptr {
	return (uintptr)(page.Pointer())
}

func (page Page) Pointer() unsafe.Pointer {
	return unsafe.Pointer(unsafe.SliceData(page))
}

var freePagesList *page
var AllocatedPages int = 0

func FreePage(addr uintptr) {
	if addr%PAGE_SIZE != 0 {
		log.KDebugLn("[PAGE] WARNING: freeing Page but is not page aligned: ", addr)
		panic.KernelPanic("[Page] non-aligned page")
	}
	// Just to check for immediate double free
	// If I were to check for double freeing correctly I would have to traverse the list
	// every time completely but that would make freeing O(n)
	if addr == uintptr(unsafe.Pointer(freePagesList)) {
		log.KDebugLn("[Page] immediate double freeing page ", addr)
		panic.KernelPanic("[Page] double freeing page")
	}
	p := utils.UIntToPointer[page](addr)
	p.next = freePagesList
	freePagesList = p
	AllocatedPages--
}

func AllocPage() Page {
	if freePagesList == nil {
		panic.KernelPanic("[PAGE] Out of pages to allocate")
	}
	p := freePagesList
	freePagesList = p.next
	AllocatedPages++
	if PAGE_DEBUG {
		log.KDebugLn("[PAGE]: Allocated ", unsafe.Pointer(p))
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(p)), PAGE_SIZE)
}

func Memclr(p uintptr, n int) {
	// s := (*(*[1 << 30]byte)(unsafe.Pointer(p)))[:n]
	s := utils.UIntToSlice[byte](p, n)
	clear(s)
	// for i := range s {
	// 	s[i] = 0
	// }
}
