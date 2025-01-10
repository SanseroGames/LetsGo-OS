package panic

import _ "unsafe"

// Link to kernel panic function of kernel package. This is done using a linkname
// to prevent cyclic import for packages that are used by the kernel package but
// also would want to kernelPanic
//
//go:linkname KernelPanic github.com/sanserogames/letsgo-os/kernel.kernelPanic
func KernelPanic(msg string)
