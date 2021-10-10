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
    //MOVL ·segments_ds+12(FP), AX
    //ORW $3, AX
    //MOVW AX, DS
    //MOVL ·segments_es+0(FP), AX
    //ORW $3, AX
    //MOVW AX, ES
    //MOVL ·segments_fs+16(FP), AX
    //ORW $3, AX
    //MOVW AX, FS
    //MOVL ·segments_gs+20(FP), AX
    //ORW $3, AX
    //MOVW AX, GS

    ////Set up Stackframe iret expects 
    //MOVL ·segments_ss+8(FP), AX
    //ORW $3, AX
    //PUSHL AX
    //MOVL ·stackAddr+28(FP), AX
    //PUSHL AX
    //PUSHFL
    //MOVL ·segments_cs+4(FP), AX
    //ORW $3, AX
    //PUSHL AX
    //MOVL ·funcAddr+24(FP), AX
    //PUSHL AX
    //
    //IRETL