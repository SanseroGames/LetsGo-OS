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
     EIP uintptr
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
    PRINT_INTERRUPT_DEBUG = false
)

type InterruptHandler func()

var (
    idtTable = [256]IdtEntry{}
    idtDescriptor IdtDescriptor = IdtDescriptor{}
    handlers [256]InterruptHandler
    PerformSchedule = false
    oldThread *thread
)

func interrupt_debug(args ...interface{}) {
    if PRINT_INTERRUPT_DEBUG {
        kdebugln(args)
    }
}

func isrVector()

// Actually an array of functions disguised as a function
func isrEntryList()

func installIDT(descriptor *IdtDescriptor)

func getIDT() *IdtDescriptor

func EnableInterrupts()
func DisableInterrupts()

func scheduleStack(fn func())

// Pass an argument to function on schedule stack. If more complex strucutres need to be passed, pass pointer and cast it accordingly
func scheduleStackArg(fn func(arg uintptr), arg uintptr)


//go:nosplit
func scheduleStackFail(caller uintptr){
    kprint("Called from ")
    printFuncName(caller-4)
    kernelPanic("Called schedulestack on schedule stack")
}

//go:nosplit
func stackFail(esp, failingAddress uintptr) {
    setGS(KGS_SELECTOR)
    setDS(KDS_SELECTOR)
    kerrorln("Return from schedule stack to a location not in the kernel")
    kprintln("Offending Address: ", failingAddress, ", ESP: ", esp)
    kernelPanic("Fix pls")
}

func doubleStackReturn(){
    kernelPanic("Returing from scheduleStack twice")
}

func setDS(ds_segment uint32)
func setGS(gs_segment uint32)

//go:nosplit
func do_isr(regs RegisterState, info InterruptInfo){
    setGS(KGS_SELECTOR)
    setDS(KDS_SELECTOR)

    switchPageDir(kernelMemSpace.PageDirectory)
    SetInterruptStack(scheduleThread.kernelStack.hi)

    //printRegisters(defaultLogWriter, &info, &regs)

    if regs.KernelESP < uint32(scheduleThread.kernelStack.lo) {
        // Test for scheduler stack underflow
        DisableInterrupts()
        Hlt()
        text_mode_print_hex32(info.ESP)
        text_mode_print_hex32(uint32(scheduleThread.kernelStack.lo))
        kernelPanic("Stack underflow")
    }

    if info.CS == KCS_SELECTOR {
        // We were interrupting the kernel
        currentThread.kernelInfo = info
        currentThread.kernelRegs = regs
        currentThread.isKernelInterrupt = true
        currentThread.interruptedKernelEIP = info.EIP
        currentThread.interruptedKernelESP = info.ESP
        // Disable interrupts for interrupted kernel threads
        currentThread.kernelInfo.EFLAGS &= 0xffffffff ^ EFLAGS_IF

    } else {
        if regs.KernelESP > uint32(currentThread.kernelStack.hi) ||
            regs.KernelESP < uint32(currentThread.kernelStack.lo) {
            text_mode_print_errorln("kernel stack for process is out of range")
            text_mode_print("KernelESP: ")
            text_mode_print_hex32(regs.KernelESP)
            text_mode_print(" Stack Hi: ")
            text_mode_print_hex32(uint32(currentThread.kernelStack.hi))
            text_mode_print(" Stack Low: ")
            text_mode_print_hex32(uint32(currentThread.kernelStack.lo))
            text_mode_print("\n TSS ESP0: ")
            text_mode_print_hex32(tss.esp0)
            text_mode_println("")
            kernelPanic("Fix pls")
        }
        currentThread.info = info
        currentThread.regs = regs
        currentThread.isKernelInterrupt = false
    }

    interrupt_debug("[INTERRUPT-IN] Debug infos")
    if PRINT_INTERRUPT_DEBUG {
        printTid(defaultLogWriter, currentThread)
        printThreadRegisters(currentThread, defaultLogWriter)
    }

    handlers[info.InterruptNumber]()

    if PerformSchedule {
        // If not already on schedule stack
        if regs.KernelESP < uint32(scheduleThread.kernelStack.hi) && regs.KernelESP > uint32(scheduleThread.kernelStack.lo) {
            // We're already on scheduler stack
            interrupt_debug("Scheduling on schedule kernel thread ")
            Schedule()
        } else {
            // We're on threads kernel stack
            interrupt_debug("Scheduling on user kernel thread")
            scheduleStack(func(){
                Schedule()
            })
            interrupt_debug("Performed schedule")
        }
        PerformSchedule = false
    }

    interrupt_debug("[INTERRUPT-OUT] Debug infos")
    if PRINT_INTERRUPT_DEBUG {
        printTid(defaultLogWriter, currentThread)
        printThreadRegisters(currentThread, defaultLogWriter)
    }

    if currentThread.isKernelInterrupt {
        // We were interrupting the kernel
        info = currentThread.kernelInfo
        regs = currentThread.kernelRegs
        info.EIP = currentThread.interruptedKernelEIP
        info.ESP = currentThread.interruptedKernelESP
        currentThread.isKernelInterrupt = false
        interrupt_debug("Kernel return")
    } else {
        info = currentThread.info
        regs = currentThread.regs
        currentThread.isKernelInterrupt = false
        SetInterruptStack(currentThread.kernelStack.hi)
        switchPageDir(currentThread.domain.MemorySpace.PageDirectory)
        interrupt_debug("User return")
    }
    interrupt_debug("[INTERRUPT-OUT] Returning to ", info.EIP, " with stack ", uintptr(info.ESP))


}

func SetInterruptHandler(irq uint8, f InterruptHandler, selector int, priv uint8){
    handlers[irq] = f
    idtTable[irq].selector = uint16(selector)
    idtTable[irq].flags = uint16(priv | 0xE | PRESENT)<<8

}

func defaultHandler(){
    kerrorln("\nUnhandled interrupt! Disabling Interrupt and halting!")
    kprintln("Interrupt number: ", currentThread.info.InterruptNumber)
    kprintln("Exception code: ", uintptr(currentThread.info.ExceptionCode))
    kernelPanic("fix pls")
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
