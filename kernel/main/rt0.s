#include "go_asm.h"
#include "textflag.h"

// According to the "ELF handling for TLS" document section 4.3.2
// (https://www.akkadia.org/drepper/tls.pdf) for the GNU variant of the IA-32 ABI, 
// gs:0x00 contains a pointer to the TCB. Variables in the TLS are stored 
//  before the TCB and are accessed using negative offsets from the TCB address.
DATA g0_ptr(SB)/4, $0
GLOBL g0_ptr(SB), NOPTR, $4
DATA tcb_ptr(SB)/4, $0
GLOBL tcb_ptr(SB), NOPTR, $4

DATA tmp_multiboot_info(SB)/4, $0
GLOBL tmp_multiboot_info(SB), NOPTR, $4

#define G_STACK_LO 0x0
#define G_STACK_HI 0x4
#define G_STACKGUARD0 0x8

// Kernel entry point
TEXT _rt0_entry(SB), NOSPLIT+TOPFRAME+NOFRAME, $0
	CMPL AX, $(const_multibootMagic)
	JNE unsupported_bootloader

	MOVL BX, tmp_multiboot_info(SB)

	LEAL ·stack(SB),AX

	// Enforce 16-byte boundary alignment
	ADDL $0xf, AX 
	ANDL $(-1 ^ 0xf), AX

	// Initalize our stack by pointing ESP to the BSS-allocated stack. In x86,
	// stack grows downwards so we need to point ESP to stack_top
	MOVL AX, BX
	ADDL $(const_stackSize), AX
	MOVL AX, SP

	MOVL BX, runtime·g0+G_STACK_LO(SB)
	MOVL AX, runtime·g0+G_STACK_HI(SB)
	MOVL BX, runtime·g0+G_STACKGUARD0(SB)
	MOVL $runtime·g0(SB), g0_ptr(SB)


	// Enable SSE/AVX
	CALL _rt0_enable_sse(SB)

	// Load initial GDT
 	CALL _rt0_load_gdt(SB)

	// jump into the go code
	PUSHL BX
	PUSHL AX
	MOVL tmp_multiboot_info(SB), BX
	PUSHL BX

	CALL ·main(SB)

	// Main should never return; halt the cpu
halt:
	CLI
	HLT

unsupported_bootloader:
	// First entry of string type in Go is the pointer to the string data
	MOVL ·unsupportedBootloaderMsg(SB), DI
	CALL write_string(SB)
	JMP halt

//------------------------------------------------------------------------------
// Write the NULL-terminated string contained in edi to the screen using white
// text on red background.  Assumes that text-mode is enabled and that its
// physical address is 0xb8000.
//------------------------------------------------------------------------------
TEXT write_string(SB), NOSPLIT, $0
	PUSHL AX
	PUSHL BX

	MOVL $0xb8000, BX
	MOVB $0x4F, AH       // white text on red background
	
next_char:
	MOVB 0(DI), AL
	TESTB AL, AL
	JZ done
	
	// Write character and attribute to video memory
	MOVW AX, 0(BX)
	ADDL $2, BX
	ADDL $1, DI
	JMP next_char
	
done:
	POPL BX
	POPL AX
	RET

//------------------------------------------------------------------------------
// Enable SSE support. Code taken from:
// http://wiki.osdev.org/SSE#Checking_for_SSE
//------------------------------------------------------------------------------
TEXT _rt0_enable_sse(SB), NOSPLIT, $0
	PUSHL AX
	PUSHL BX

	// check for SSE
	MOVL $1, AX
	CPUID
	TESTL $(1<<25), DX
	JZ no_sse

	// Enable SSE
	MOVL CR0, AX
	ANDL $0xFFFB, AX        // Clear CR0.EM (coprocessor emulation)
	ORL $0x2, AX            // Set CR0.MP (coprocessor monitoring)
	MOVL AX, CR0
	
	MOVL CR4, AX
	ORL $(3 << 9), AX          // Set CR4.OSFXSR and CR4.OSXMMEXCPT (bits 9-10)
	MOVL AX, CR4

	POPL BX
	POPL AX
	RET
no_sse:
	CLI
	HLT


// rt0_load_gdt loads the GDT and sets up segment registers
TEXT _rt0_load_gdt(SB), NOSPLIT, $0
	PUSHL AX
	PUSHL BX

	// Store the address to the TCB in tcb_ptr
	// and set up gs base address to it: todo fix
	MOVL $tcb_ptr(SB), AX
	MOVL AX, tcb_ptr(SB)
	// TODO: Use struct sizes?
	MOVL $·initialGdt+const_kGS_SELECTOR(SB), BX
	MOVB AL, 2(BX)
	MOVB AH, 3(BX)
	SHRL $16, AX
	MOVB AL, 4(BX)

	// Load GDT
	MOVL $·initialGdtDescriptor(SB), AX
	MOVL $·initialGdt(SB), BX
	MOVL BX, (gdtDescriptor_GdtAddressLow)(AX)
	
	LGDT (AX)
	
	// GDT has been loaded but the CPU still has the previous GDT data in cache.
	// We need to manually update the descriptors and use a JMP command to set
	// the CS segment descriptor
	// since go does not support far jumps with setting the CS register
	// we use a workaround using the RETF instruction
	CALL _rt0_update_cs_register(SB)
	
update_descriptors:
	MOVL $(const_kDS_SELECTOR), AX
	MOVW AX, DS
	MOVW AX, ES
	MOVW AX, FS
	MOVW AX, SS
	
	MOVL $(const_kGS_SELECTOR), AX
	MOVW AX, GS
	
	POPL BX
	POPL AX
	RET

TEXT _rt0_update_cs_register(SB), NOSPLIT+NOFRAME, $0
	POPL AX
	PUSHL $const_kCS_SELECTOR
	PUSHL AX
	BYTE $0xCB // RETF, used to also load the CS register
