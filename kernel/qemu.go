package kernel

func Shutdown() {
	kdebugln("Shutting down...")
	Outw(0x604, 0x2000)
	kernelPanic("Qemu shutdown did not work :(")
}
