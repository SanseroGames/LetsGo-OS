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

var freePagesList *page
var AllocatedPages int = 0

func FreePage(addr uintptr) {
	if addr%PAGE_SIZE != 0 {
		log.KDebugLn("[PAGE] WARNING: freeingPage but is not page aligned: ", addr)
		return
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

func AllocPage() uintptr {
	if freePagesList == nil {
		panic.KernelPanic("[PAGE] Out of pages to allocate")
	}
	p := freePagesList
	freePagesList = p.next
	AllocatedPages++
	if PAGE_DEBUG {
		log.KDebugLn("[PAGE]: Allocated ", unsafe.Pointer(p))
	}
	return uintptr(unsafe.Pointer(p))
}

func Memclr(p uintptr, n int) {
	// s := (*(*[1 << 30]byte)(unsafe.Pointer(p)))[:n]
	s := utils.UIntToSlice[byte](p,n)
	for i := range s {
		s[i] = 0
	}
}
