#include "textflag.h"

TEXT ·Outb(SB),NOSPLIT,$0
    MOVW port+0(FP), DX
    MOVB value+2(FP), AX
    OUTB
    RET

TEXT ·Inb(SB),NOSPLIT,$0-5
    MOVW port+0(FP), DX
    INB
    MOVB AX, ret+4(FP)
    RET

TEXT ·Hlt(SB),NOSPLIT,$0
    HLT
    RET
