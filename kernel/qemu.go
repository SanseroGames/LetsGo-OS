package kernel

import (
	"github.com/sanserogames/letsgo-os/kernel/log"
)

func Shutdown() {
	log.KDebugLn("Shutting down...")
	Outw(0x604, 0x2000)
	kernelPanic("Qemu shutdown did not work :(")
}
