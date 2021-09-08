#include "textflag.h"

TEXT ·backupFpRegs(SB),NOSPLIT,$0
    MOVL ·buffer+0(FP), AX
    FXSAVE (AX)
    RET

TEXT ·restoreFpRegs(SB),NOSPLIT,$0
    MOVL ·buffer+0(FP), AX
    FXRSTOR (AX)
    RET
