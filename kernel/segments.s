#include "textflag.h"

TEXT ·installGDT(SB),NOSPLIT,$0
    MOVL ·descriptor(FP), AX
    LGDT (AX)
	RET

TEXT ·getGDT(SB),NOSPLIT,$0
    SGDT (AX)
    MOVL AX, ret+0(FP)
    RET
