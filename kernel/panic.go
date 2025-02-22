package kernel

import (
	"github.com/sanserogames/letsgo-os/kernel/log"
)

// wrapper for do_kernelPanic that gets the return address
// and pushers it on the stack and then calls do_kernelPanic
func kernelPanic(msg string)

//go:nosplit
func do_kernelPanic(caller uintptr, msg string) {
	log.KErrorLn("\n", msg, " - kernel panic :(")
	log.KPrint("Called from function: ")
	printFuncName(caller - 4) // account for the fact that caller points to the instruction after the call
	if CurrentThread != nil {
		panicHelper(CurrentThread)
	} else {
		log.KErrorLn("Cannot print registers. 'currentThread' is nil")
	}
	DisableInterrupts()
	Hlt()
	// does not return
}

func panicHelper(thread *Thread) {
	log.KPrintLn("Domain ID: ", thread.domain.pid, ", Thread ID: ", thread.tid)
	log.KPrintLn("Program name: ", thread.domain.programName)
	if thread.isKernelInterrupt {
		log.KPrint("In kernel function: ")
		printFuncName(thread.kernelInfo.EIP)
	} else {
		log.KPrintLn("In user function: ", thread.info.EIP)
	}
	printThreadRegisters(thread)
	DisableInterrupts()
	Hlt()
}

func printThreadRegisters(t *Thread) {
	log.KPrint("User regs:          Kernel regs:\n")
	f := runtimeFindFunc(uintptr(t.kernelInfo.EIP))
	log.KPrint("EIP: ", t.info.EIP, "      ", "EIP: ", t.kernelInfo.EIP, " ", f._Func().Name(), "\n")
	//rintRegisterLineInfo("EIP: ", t.info.EIP, t.kernelInfo.EIP, f._Func().Name())
	printRegisterLine(20, "ESP: ", t.info.ESP, t.kernelInfo.ESP)
	printRegisterLine(20, "EBP: ", t.regs.EBP, t.kernelRegs.EBP)
	printRegisterLine(20, "EAX: ", t.regs.EAX, t.kernelRegs.EAX)
	printRegisterLine(20, "EBX: ", t.regs.EBX, t.kernelRegs.EBX)
	printRegisterLine(20, "ECX: ", t.regs.ECX, t.kernelRegs.ECX)
	printRegisterLine(20, "EDX: ", t.regs.EDX, t.kernelRegs.EDX)
	printRegisterLine(20, "ESI: ", t.regs.ESI, t.kernelRegs.ESI)
	printRegisterLine(20, "EDI: ", t.regs.EDI, t.kernelRegs.EDI)
	printRegisterLine(20, "EFLAGS: ", t.info.EFLAGS, t.kernelInfo.EFLAGS)
	printRegisterLine(20, "Exception: ", t.info.ExceptionCode, t.kernelInfo.ExceptionCode)
	printRegisterLine(20, "Interrupt: ", t.info.InterruptNumber, t.kernelInfo.InterruptNumber)
	printRegisterLine(20, "Krn ESP: ", t.regs.KernelESP, t.kernelRegs.KernelESP)
}

func printRegisterLine(tabLength int, label string, userReg, kernelReg uint32) {
	firstLength := len(label)
	log.KPrint(label, uintptr(userReg))
	// pad number
	firstLength += 3 // account for the hexadecimal 0x#
	for i, n := firstLength, userReg>>4; i < tabLength; i, n = i+1, n>>4 {
		if n == 0 {
			log.KPrint(" ")
		}
	}
	log.KPrint(label, uintptr(kernelReg), "\n")
}

// func printRegisters(info *InterruptInfo, regs *RegisterState) {
// 	log.KPrintLn("Interrupt: ", uintptr(info.InterruptNumber))
// 	log.KPrintLn("Exception: ", uintptr(info.ExceptionCode))
// 	log.KPrintLn("EIP: ", uintptr(info.EIP))
// 	log.KPrintLn("CS: ", uintptr(info.CS))
// 	log.KPrintLn("EFLAGS: ", uintptr(info.EFLAGS))
// 	log.KPrintLn("ESP: ", uintptr(info.ESP))
// 	log.KPrintLn("SS: ", uintptr(info.SS))
// 	log.KPrintLn("-----------")
// 	log.KPrintLn("GS: ", uintptr(regs.GS))
// 	log.KPrintLn("FS: ", uintptr(regs.FS))
// 	log.KPrintLn("ES: ", uintptr(regs.ES))
// 	log.KPrintLn("DS: ", uintptr(regs.DS))
// 	log.KPrintLn("EBP: ", uintptr(regs.EBP))
// 	log.KPrintLn("EAX: ", uintptr(regs.EAX))
// 	log.KPrintLn("EBX: ", uintptr(regs.EBX))
// 	log.KPrintLn("ECX: ", uintptr(regs.ECX))
// 	log.KPrintLn("EDX: ", uintptr(regs.EDX))
// 	log.KPrintLn("ESI: ", uintptr(regs.ESI))
// 	log.KPrintLn("EDI: ", uintptr(regs.EDI))
// 	log.KPrintLn("KernelESP", uintptr(regs.KernelESP))
// }
