#include "textflag.h"

TEXT ·kernelPanic(SB),NOSPLIT,$0
    CALL ·do_kernelPanic(SB)
    RET
