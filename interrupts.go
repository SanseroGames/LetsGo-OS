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
     EIP uint32
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
    KernelESP uint32
    EBX uint32
    EDX uint32
    ECX uint32
    EAX uint32
}

const (
    INTERRUPT_GATE = 0xE
)

var (
    idtTable = [256]IdtEntry{}
    idtDescriptor IdtDescriptor = IdtDescriptor{}
    handlers [256]func(info *InterruptInfo, regs *RegisterState)
    infoEscapePreventer InterruptInfo
    regsEscapePreventer RegisterState
)

func isrVector()

// Actually an array of functions disguised as a function
func isrEntryList()

func installIDT(descriptor *IdtDescriptor)

func getIDT() *IdtDescriptor

func EnableInterrupts()
func DisableInterrupts()

func setDS(ds_segment uint32)
func setGS(gs_segment uint32)

//go:nosplit
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
    setGS(KGS_SELECTOR)
    setDS(KDS_SELECTOR)

    switchPageDir(kernelMemSpace.PageDirectory)

    infoEscapePreventer = info
    regsEscapePreventer = regs
    handlers[info.InterruptNumber](&infoEscapePreventer, &regsEscapePreventer)
    Schedule(&infoEscapePreventer, &regsEscapePreventer)
    info = infoEscapePreventer
    regs = regsEscapePreventer

    switchPageDir(currentDomain.MemorySpace.PageDirectory)
}

func SetInterruptHandler(irq uint8, f func (info *InterruptInfo, regs *RegisterState), selector int, priv uint8){
    handlers[irq] = f
    idtTable[irq].selector = uint16(selector)
    idtTable[irq].flags = uint16(priv | 0xE | PRESENT)<<8

}

func defaultHandler(info *InterruptInfo, regs *RegisterState){
    text_mode_print_char(0xa)
    text_mode_print_errorln("Unhandled interrupt! Disabling Interrupt and halting!")
    text_mode_print("Interrupt number: ")
    text_mode_print_hex(uint8(info.InterruptNumber))
    text_mode_print_char(0xa)
    text_mode_print("Exception code: ")
    text_mode_print_hex32(info.ExceptionCode)
    text_mode_print_char(0xa)
    text_mode_print("EIP: ")
    text_mode_print_hex32(info.EIP)
    DisableInterrupts()
    Hlt()
}

func InitInterrupts(){
    isrBaseAddr := reflect.ValueOf(isrEntryList).Pointer()
    for i := range idtTable {
        if(i == 2 || i == 15) {continue;}
        isrAddr := isrBaseAddr + uintptr(i*23)
        low := uint16(isrAddr)
        high := uint16(uint32(isrAddr)>>16)
        idtTable[i].offsetLow = low
        idtTable[i].offsetHigh = high
        SetInterruptHandler(uint8(i), defaultHandler, KCS_SELECTOR, PRIV_KERNEL)
    }
    idtDescriptor.IdtSize = uint16(uintptr(len(idtTable))*unsafe.Sizeof(idtTable[0]))-1
    idtAddr := uint32(uintptr(unsafe.Pointer(&idtTable)))
    idtDescriptor.IdtAddressLow = uint16(idtAddr)
    idtDescriptor.IdtAddressHigh = uint16(idtAddr >> 16)
    installIDT(&idtDescriptor)
}

func printIdt(idt []IdtEntry){
    for _,n := range idt {
        text_mode_print_hex16(n.offsetLow)
        text_mode_print(" ")
        text_mode_print_hex16(n.selector)
        text_mode_print(" ")
        text_mode_print_hex16(n.flags)
        text_mode_print(" ")
        text_mode_print_hex16(n.offsetHigh)
        text_mode_println("")
    }
}
