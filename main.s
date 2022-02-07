#include "textflag.h"

TEXT ·kernelPanic(SB),NOSPLIT,$0
    MOVL 0(SP), AX
    PUSHL AX
    JMP ·do_kernelPanic(SB)
