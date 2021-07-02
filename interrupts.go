package main

import (
    "unsafe"
    "reflect"
)

type IdtEntry struct {
    offsetLow   uint16
    selector    uint16
    flags       uint16
    offsetHigh  uint16
}

type IdtDescriptor struct {
    IdtSize    uint16
    IdtAddressLow uint16
    IdtAddressHigh uint16
}

type InterruptInfo struct {
     InterruptNumber uint32
     ExceptionCode uint32
     RIP uint32
     CS uint32
     EFLAGS uint32
     ESP uint32
     SS uint32
}

// Reverse of stack pushing
type RegisterState struct {
     GS uint32
     FS uint32
     ES uint32
     DS uint32

     EDI uint32
     ESI uint32
     EBP uint32
     ESP uint32
     EBX uint32
     EDX uint32
     ECX uint32
     EAX uint32
}


var idtTable = [256]IdtEntry{}

var idtDescriptor IdtDescriptor = IdtDescriptor{}

var handlers [256]func(intNum uint8)

func isrVector()

// Actually an array of functions disguised as a function
func isrEntryList()

func installIDT(descriptor *IdtDescriptor)

func EnableInterrupts()
func DisableInterrupts()

func do_isr(regs RegisterState, info InterruptInfo){
    //text_mode_print("Interrupt")
    //text_mode_print_hex(uint8(info.SS))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(info.ESP))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(info.EFLAGS))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(info.CS))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(info.RIP))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(info.ExceptionCode))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(info.InterruptNumber))
    //text_mode_print_char(0x20)

    //text_mode_print_hex(uint8(regs.EAX))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(regs.ECX))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(regs.EDX))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(regs.EBX))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(regs.ESP))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(regs.EBP))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(regs.ESI))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(regs.EDI))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(regs.DS))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(regs.ES))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(regs.FS))
    //text_mode_print_char(0x20)
    //text_mode_print_hex(uint8(regs.GS))
    //text_mode_print_char(0x20)

    handlers[info.InterruptNumber](uint8(info.InterruptNumber))
}

func SetInterruptHandler(irq uint8, f func (intNum uint8)){
    handlers[irq] = f
}

func defaultHandler(intNum uint8){
    text_mode_print("Default handler")
}

func InitInterrupts(){
    isrBaseAddr := reflect.ValueOf(isrEntryList).Pointer()
    for i := range idtTable {
        if(i == 2 || i == 15) {continue;}
        isrAddr := isrBaseAddr + uintptr(i*23)
        low := uint16(isrAddr)
        high := uint16(uint32(isrAddr)>>16)
        idtTable[i].offsetLow = low
        idtTable[i].selector = 0x08
        idtTable[i].flags = 0x8E00
        idtTable[i].offsetHigh = high
        SetInterruptHandler(uint8(i), defaultHandler)
    }
    text_mode_print("Hi from interrupts")
    idtDescriptor.IdtSize = uint16(uintptr(len(idtTable))*unsafe.Sizeof(idtTable[0]))-1
    idtAddr := uint32(uintptr(unsafe.Pointer(&idtTable)))
    idtDescriptor.IdtAddressLow = uint16(idtAddr)
    idtDescriptor.IdtAddressHigh = uint16(idtAddr >> 16)
    installIDT(&idtDescriptor)
}


