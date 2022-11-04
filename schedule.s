#include "textflag.h"

TEXT ·backupFpRegs(SB),NOSPLIT,$0
    MOVL ·buffer+0(FP), AX
    FXSAVE (AX)
    RET

TEXT ·restoreFpRegs(SB),NOSPLIT,$0
    MOVL ·buffer+0(FP), AX
    FXRSTOR (AX)
    RET

TEXT ·getESP(SB),NOSPLIT,$4
    MOVL SP, ret+0(FP)
    RET

TEXT ·waitForInterrupt(SB),NOSPLIT,$0
    // When an interrunt in kernel space happens it does not push stack info on the stack
    // We do this here so it is found in the interrupt
    MOVL SP, AX
    PUSHL SS
    PUSHL AX
    STI
    HLT
    CLI
    // As with handling the interrupt, iret does not clear the stack data of the stack when
    // staying in kernel mode. So we clean up ourselves
    POPL AX
    POPL AX
    RET
