set timeout=0
set default=0

menuentry "letsgoos" {
    multiboot2 /boot/kernel.bin
 {MODULES} 
    boot
}
