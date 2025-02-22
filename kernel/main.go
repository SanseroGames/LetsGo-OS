package kernel

import (
	"path"
	"unsafe"

	"github.com/sanserogames/letsgo-os/kernel/log"
)

const ENABLE_DEBUG = false

func printFuncName(pc uintptr) {
	f := runtimeFindFunc(pc)
	if !f.valid() {
		log.KPrintLn("func: ", pc)
		return
	}
	s := f._Func().Name()
	file, line := f._Func().FileLine(pc)
	_, filename := path.Split(file)
	log.KPrintLn(s, " (", filename, ":", line, ")")
}

func printTid(t *Thread) {
	log.KPrintLn("Pid: ", t.domain.pid, ", Tid: ", t.tid)
}

func cstring(ptr uintptr) string {
	var n int
	for p := ptr; *(*byte)(unsafe.Pointer(p)) != 0; p++ {
		n++
	}
	return unsafe.String((*byte)(unsafe.Pointer(ptr)), n)
}
