#include "textflag.h"

TEXT ·flushTss(SB),NOSPLIT,$0
    MOVW ·segmentIndex(FP), AX
    LTR AX
	RET

TEXT ·hackyGetFuncAddr(SB),NOSPLIT,$0
    MOVL ·funcAddr+0(FP), AX
    MOVL (AX), AX
    MOVL AX, ·ret+4(FP)
    RET

TEXT ·JumpUserMode(SB),NOSPLIT,$0
    MOVL ·segments_ds+12(FP), AX
    ORW $3, AX
    MOVW AX, DS
    MOVL ·segments_es+0(FP), AX
    ORW $3, AX
    MOVW AX, ES
    MOVL ·segments_fs+16(FP), AX
    ORW $3, AX
    MOVW AX, FS
    MOVL ·segments_gs+20(FP), AX
    ORW $3, AX
    MOVW AX, GS

    //Set up Stackframe iret expects 
    MOVL ·segments_ss+8(FP), AX
    ORW $3, AX
    PUSHL AX
    MOVL ·stackStart+28(FP), AX
    PUSHL AX
    PUSHFL
    MOVL ·segments_cs+4(FP), AX
    ORW $3, AX
    PUSHL AX
    // Possible that this will fail in the future
    MOVL ·funcAddr+24(FP), AX
    PUSHL AX
    
    IRETL
