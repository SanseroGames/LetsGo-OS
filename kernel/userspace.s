#include "textflag.h"

TEXT ·flushTss(SB),NOSPLIT,$0
    MOVW ·segmentIndex(FP), AX
    LTR AX
	RET

// Possible that this will fail in the future
TEXT ·hackyGetFuncAddr(SB),NOSPLIT,$0
    MOVL ·funcAddr+0(FP), AX
    MOVL (AX), AX
    MOVL AX, ·ret+4(FP)
    RET

TEXT ·JumpUserMode(SB),NOSPLIT,$0
    POPL AX
    POPL GS
    POPL FS
    POPL ES
    POPL DS
    POPAL
    ADDL $8, SP
    IRETL
    
