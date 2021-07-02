package main

func Inb(port uint16) uint8

func Outb(port uint16, value uint8)

func Hlt()

func test(i uint32){
    debug_print_flags(uint8(i))
}
