package kernel

func Inb(port uint16) uint8
func Inw(port uint16) uint16

func Outb(port uint16, value uint8)
func Outw(port uint16, value uint16)

func Hlt()
