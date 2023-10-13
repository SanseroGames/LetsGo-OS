package main

import (
	"debug/elf"
	"reflect"
	"unsafe"
)

type auxVecEntry struct {
	Type  uint32
	Value uint32
}

const (
	AT_NULL     = 0  /* end of vector */
	AT_IGNORE   = 1  /* entry should be ignored */
	AT_EXECFD   = 2  /* file descriptor of program */
	AT_PHDR     = 3  /* program headers for program */
	AT_PHENT    = 4  /* size of program header entry */
	AT_PHNUM    = 5  /* number of program headers */
	AT_PAGESZ   = 6  /* system page size */
	AT_BASE     = 7  /* base address of interpreter */
	AT_FLAGS    = 8  /* flags */
	AT_ENTRY    = 9  /* entry point of program */
	AT_NOTELF   = 10 /* program is not ELF */
	AT_UID      = 11 /* real uid */
	AT_EUID     = 12 /* effective uid */
	AT_GID      = 13 /* real gid */
	AT_EGID     = 14 /* effective gid */
	AT_PLATFORM = 15 /* string identifying CPU for optimizations */
	AT_HWCAP    = 16 /* arch dependent hints at CPU capabilities */
	AT_CLKTCK   = 17 /* frequency at which times() increments */
	/* values 18 through 22 are reserved */
	AT_SECURE        = 23 /* secure mode boolean */
	AT_BASE_PLATFORM = 24 /* string identifying real platform, may
	 * differ from AT_PLATFORM. */
	AT_RANDOM = 25 /* address of 16 random bytes */
	AT_HWCAP2 = 26 /* extension of AT_HWCAP */

	AT_EXECFN = 31 /* filename of program */
)

func LoadAuxVector(buf []auxVecEntry, elfHdr *elf.Header32, loadAddr uintptr) int {
	//NEW_AUX_ENT(AT_HWCAP, ELF_HWCAP);
	start := 0
	buf[start].Type = AT_PAGESZ
	buf[start].Value = PAGE_SIZE
	start++
	buf[start].Type = AT_UID
	buf[start].Value = 0
	start++
	buf[start].Type = AT_GID
	buf[start].Value = 0
	start++
	buf[start].Type = AT_EUID
	buf[start].Value = 0
	start++
	buf[start].Type = AT_EGID
	buf[start].Value = 0
	start++
	buf[start].Type = AT_ENTRY
	buf[start].Value = elfHdr.Entry
	start++
	buf[start].Type = AT_PHDR
	buf[start].Value = uint32(loadAddr) + elfHdr.Phoff
	start++
	buf[start].Type = AT_PHENT
	buf[start].Value = uint32(elfHdr.Phentsize)
	start++
	buf[start].Type = AT_PHNUM
	buf[start].Value = uint32(elfHdr.Phnum)
	start++
	buf[start].Type = AT_RANDOM
	buf[start].Value = uint32(loadAddr) // Currently just pointing somewhere. don't care about security
	start++
	buf[start].Type = AT_NULL
	buf[start].Value = 0
	start++

	//NEW_AUX_ENT(AT_CLKTCK, CLOCKS_PER_SEC);
	//NEW_AUX_ENT(AT_PHDR, load_addr + exec->e_phoff);
	//NEW_AUX_ENT(AT_PHENT, sizeof(struct elf_phdr));
	//NEW_AUX_ENT(AT_PHNUM, exec->e_phnum);
	//NEW_AUX_ENT(AT_BASE, interp_load_addr);
	//if (bprm->interp_flags & BINPRM_FLAGS_PRESERVE_ARGV0)
	//flags |= AT_FLAGS_PRESERVE_ARGV0;
	//NEW_AUX_ENT(AT_FLAGS, flags);
	//NEW_AUX_ENT(AT_ENTRY, e_entry);
	//NEW_AUX_ENT(AT_UID, from_kuid_munged(cred->user_ns, cred->uid));
	//NEW_AUX_ENT(AT_EUID, from_kuid_munged(cred->user_ns, cred->euid));
	//NEW_AUX_ENT(AT_GID, from_kgid_munged(cred->user_ns, cred->gid));
	//NEW_AUX_ENT(AT_EGID, from_kgid_munged(cred->user_ns, cred->egid));
	//NEW_AUX_ENT(AT_SECURE, bprm->secureexec);
	//NEW_AUX_ENT(AT_RANDOM, (elf_addr_t)(unsigned long)u_rand_bytes);
	//#ifdef ELF_HWCAP2
	//NEW_AUX_ENT(AT_HWCAP2, ELF_HWCAP2);
	//#endif
	//NEW_AUX_ENT(AT_EXECFN, bprm->exec);
	//if (k_platform) {
	//	NEW_AUX_ENT(AT_PLATFORM,
	//		    (elf_addr_t)(unsigned long)u_platform);
	//}
	//if (k_base_platform) {
	//	NEW_AUX_ENT(AT_BASE_PLATFORM,
	//		    (elf_addr_t)(unsigned long)u_base_platform);
	//}
	//if (bprm->have_execfd) {
	//	NEW_AUX_ENT(AT_EXECFD, bprm->execfd);
	//}
	return start
}

func LoadElfFile(multibootModule string, space *MemSpace) (*elf.Header32, uintptr, uintptr) {
	var module MultibootModule
	for _, n := range loadedModules {
		if n.Cmdline == multibootModule {
			module = n
			break
		}
	}

	if module.Cmdline != multibootModule {
		kerrorln("[ELF] Unknown module: ", multibootModule, " ", module.Cmdline)
		return nil, 0, 0
	}
	moduleLen := int(module.End - module.Start)
	// catch weird things...
	if moduleLen < 4 {
		return nil, 0, 0
	}
	elfData := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Len:  moduleLen,
		Cap:  moduleLen,
		Data: uintptr(module.Start),
	}))

	// Test if really elf file
	if elfData[0] != 0x7f || elfData[1] != 'E' || elfData[2] != 'L' || elfData[3] != 'F' {
		kerrorln("[ELF] '", multibootModule, "' is not a ELF file")
		return nil, 0, 0
	}

	//elfTarget := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
	//    Len:  moduleLen,
	//    Cap:  moduleLen,
	//    Data: baseAddr,
	//}))
	elfHeader := (*elf.Header32)(unsafe.Pointer(uintptr(module.Start)))

	progHeaders := *(*[]elf.Prog32)(unsafe.Pointer(&reflect.SliceHeader{
		Len:  int(elfHeader.Phnum),
		Cap:  int(elfHeader.Phnum),
		Data: uintptr(module.Start + elfHeader.Phoff),
	}))

	baseAddr := uintptr(0xffffffff)
	topAddr := uint32(0)
	for _, n := range progHeaders {
		if n.Type == uint32(elf.PT_LOAD) {
			if uintptr(n.Vaddr) < baseAddr {
				baseAddr = uintptr(n.Vaddr)
			}
			localTop := uint32(0)
			contents := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
				Len:  int(n.Filesz),
				Cap:  int(n.Filesz),
				Data: uintptr(module.Start + n.Off),
			}))
			offset := n.Vaddr & (PAGE_SIZE - 1)
			start := uint32(0)
			if offset != 0 {
				p := allocPage()
				Memclr(p, PAGE_SIZE)
				target := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
					Len:  PAGE_SIZE,
					Cap:  PAGE_SIZE,
					Data: p,
				}))
				end := int(PAGE_SIZE - offset)
				if end > len(contents) {
					end = len(contents)
				}
				copy(target[offset:PAGE_SIZE], contents[0:end])
				space.mapPage(p, uintptr(n.Vaddr&^(PAGE_SIZE-1)), PAGE_RW|PAGE_PERM_USER)
				localTop = n.Vaddr&^(PAGE_SIZE-1) + PAGE_SIZE
				start = PAGE_SIZE - offset
			}
			for i := start; i < n.Filesz; i += PAGE_SIZE {
				p := allocPage()
				Memclr(p, PAGE_SIZE)
				target := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
					Len:  PAGE_SIZE,
					Cap:  PAGE_SIZE,
					Data: p,
				}))
				end := int(i + PAGE_SIZE)
				if end > len(contents) {
					end = len(contents)
				}
				copy(target, contents[i:end])
				// Currently don't care about write protecton of code
				space.mapPage(p, uintptr(n.Vaddr+i), PAGE_RW|PAGE_PERM_USER)
				localTop = n.Vaddr + i + PAGE_SIZE
				//delay(1000)
			}
			if n.Filesz < n.Memsz {
				for i := localTop; i < n.Vaddr+n.Memsz; i += PAGE_SIZE {
					p := allocPage()
					Memclr(p, PAGE_SIZE)
					//text_mode_print_hex32(i)
					space.mapPage(p, uintptr(i), PAGE_RW|PAGE_PERM_USER)
					localTop = i + PAGE_SIZE
				}
			}
			//text_mode_print_hex32(n.Vaddr)
			//text_mode_print(" ")
			//text_mode_print_hex32(n.Memsz)
			//text_mode_print(" ")
			//text_mode_print_hex32(module.Start + n.Off)
			//text_mode_print(" ")
			//text_mode_print_hex32(n.Filesz)
			//text_mode_print(" ")
			//text_mode_print_hex32(n.Type)
			//text_mode_print(" ")
			//text_mode_print_hex32(offset)
			//text_mode_println("")
			if localTop > topAddr {
				topAddr = localTop
			}

		}
	}
	return elfHeader, baseAddr, uintptr(topAddr)

}
