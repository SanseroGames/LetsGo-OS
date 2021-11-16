#include "textflag.h"

TEXT ·Inb(SB),NOSPLIT,$0-5
    MOVW port+0(FP), DX
    INB
    MOVB AX, ret+4(FP)
    RET

TEXT ·Inw(SB),NOSPLIT,$0-5
    MOVW port+0(FP), DX
    INW
    MOVW AX, ret+4(FP)
    RET

TEXT ·Outb(SB),NOSPLIT,$0
    MOVW port+0(FP), DX
    MOVB value+2(FP), AX
    OUTB
    RET

TEXT ·Outw(SB),NOSPLIT,$0
    MOVW port+0(FP), DX
    MOVW value+2(FP), AX
    OUTW
    RET

TEXT ·Hlt(SB),NOSPLIT,$0
    HLT
    RET
