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

type InterruptHandler func()

var (
    idtTable = [256]IdtEntry{}
    idtDescriptor IdtDescriptor = IdtDescriptor{}
    handlers [256]InterruptHandler
    PerformSchedule = false
    oldThread *thread
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

// TODO: Do I keep this? Debugging
type stack struct {
	lo uintptr
	hi uintptr
}

type g struct {
	// Stack parameters.
	// stack describes the actual stack memory: [stack.lo, stack.hi).
	// stackguard0 is the stack pointer compared in the Go stack growth prologue.
	// It is stack.lo+StackGuard normally, but can be StackPreempt to trigger a preemption.
	// stackguard1 is the stack pointer compared in the C stack growth prologue.
	// It is stack.lo+StackGuard on g0 and gsignal stacks.
	// It is ~0 on other goroutine stacks, to trigger a call to morestackc (and crash).
	stack       stack   // offset known to runtime/cgo
	stackguard0 uintptr // offset known to liblink
	stackguard1 uintptr // offset known to liblink
}

//go:nosplit
func do_isr(regs RegisterState, info InterruptInfo){
    setGS(KGS_SELECTOR)
    setDS(KDS_SELECTOR)

    switchPageDir(kernelMemSpace.PageDirectory)

    //tss.esp0 = uint32(kernelStack)
    if tss.esp0 != uint32(kernelThread.StackStart) {
        kernelPanic("AAAA")
    }
    if regs.KernelESP < uint32(kernelThread.StackEnd) {
        DisableInterrupts()
        Hlt()
        text_mode_print_hex32(info.ESP)
        text_mode_print_hex32(uint32(kernelThread.StackEnd))
        kernelPanic("Stack over or underflow")
    }

    if info.CS == KCS_SELECTOR {
        oldThread = currentThread
        currentThread = &kernelThread
    }
    currentThread.info = info
    currentThread.regs = regs
    handlers[info.InterruptNumber]()
    if info.CS == KCS_SELECTOR {
        currentThread = oldThread
    }
    if PerformSchedule {
        Schedule()
        PerformSchedule = false
    }

    if info.CS != KCS_SELECTOR {
        info = currentThread.info
        regs = currentThread.regs
    }
    if info.CS != KCS_SELECTOR {
        switchPageDir(currentThread.domain.MemorySpace.PageDirectory)
    }
}

func SetInterruptHandler(irq uint8, f InterruptHandler, selector int, priv uint8){
    handlers[irq] = f
    idtTable[irq].selector = uint16(selector)
    idtTable[irq].flags = uint16(priv | 0xE | PRESENT)<<8

}

func defaultHandler(){
    text_mode_print_char(0xa)
    text_mode_print_errorln("Unhandled interrupt! Disabling Interrupt and halting!")
    text_mode_print("Interrupt number: ")
    text_mode_print_hex(uint8(currentThread.info.InterruptNumber))
    text_mode_print_char(0xa)
    text_mode_print("Exception code: ")
    text_mode_print_hex32(currentThread.info.ExceptionCode)
    text_mode_print_char(0xa)
    text_mode_print("EIP: ")
    text_mode_print_hex32(currentThread.info.EIP)
    DisableInterrupts()
    Hlt()
}

func InitInterrupts(){
    isrBaseAddr := reflect.ValueOf(isrEntryList).Pointer()
    for i := range idtTable {
        if(i == 2 || i == 15) {continue}
        // Bytes are counted based on assembly
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
