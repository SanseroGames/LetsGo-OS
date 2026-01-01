package kernel

const (
	PAGE_DEBUG = false

	// Reserve memory below 50 MB for kernel image
	KERNEL_START      = 1 << 20
	KERNEL_RESERVED   = 50 << 20
	PAGE_SIZE         = 4 << 10
	ENTRIES_PER_TABLE = PAGE_SIZE / 4

	PAGE_PRESENT       = 1 << 0
	PAGE_RW            = 1 << 1
	PAGE_PERM_USER     = 1 << 2
	PAGE_PERM_KERNEL   = 0 << 2
	PAGE_WRITETHROUGH  = 1 << 3
	PAGE_DISABLE_CACHE = 1 << 4

	PAGE_FAULT_PRESENT           = 1 << 0
	PAGE_FAULT_WRITE             = 1 << 1
	PAGE_FAULT_USER              = 1 << 2
	PAGE_FAULT_INSTRUCTION_FETCH = 1 << 4

	MAX_ALLOC_VIRT_ADDR = 0xf0000000
	MIN_ALLOC_VIRT_ADDR = 0x8000000
)
