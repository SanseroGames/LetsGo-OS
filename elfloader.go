package main

import (
    "debug/elf"
    "unsafe"
    "reflect"
)


func LoadElfFile(multibootModule string, space *MemSpace) uint32 {
    var module MultibootModule
    for _,n := range loadedModules{
        if (n.Cmdline == multibootModule) {
            module = n
            break
        }
    }

    if module.Cmdline != multibootModule {
        text_mode_print_error("[ELF] Unknown module: ")
        text_mode_print_errorln(multibootModule)
        return 0
    }
    moduleLen := int(module.End - module.Start)
    // catch weird things...
    if moduleLen < 4 {return 0}
    elfData := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
        Len:  moduleLen,
	    Cap:  moduleLen,
	    Data: uintptr(module.Start),
    }))

    // Test if really elf file
    if elfData[0] != 0x7f  || elfData[1] != 'E' || elfData[2] != 'L' || elfData[3] != 'F' {
        text_mode_print_error("[ELF] '")
        text_mode_print_error(multibootModule)
        text_mode_print_errorln("' is not a ELF file")
        return 0
    }

    //elfTarget := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
    //    Len:  moduleLen,
	//    Cap:  moduleLen,
	//    Data: baseAddr,
    //}))
    elfHeader := *(*elf.Header32)(unsafe.Pointer(uintptr(module.Start)))

    progHeaders := *(*[]elf.Prog32)(unsafe.Pointer(&reflect.SliceHeader{
        Len:  int(elfHeader.Phnum),
	    Cap:  int(elfHeader.Phnum),
	    Data: uintptr(module.Start + elfHeader.Phoff),
    }))

    for _,n := range progHeaders {
        if n.Type == uint32(elf.PT_LOAD) {
            contents := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
                Len:  int(n.Filesz),
	            Cap:  int(n.Filesz),
	            Data: uintptr(module.Start + n.Off),
            }))
            for i := uint32(0); i < n.Filesz; i+= PAGE_SIZE {
                p := allocPage()
                Memclr(p, PAGE_SIZE)
                target := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
                    Len:  PAGE_SIZE,
	                Cap:  PAGE_SIZE,
	                Data: p,
                }))
                end := int(i+PAGE_SIZE)
                if end > len(contents) {
                    end = len(contents)
                }
                copy(target, contents[i:end])
                // Currently don't care about write protecton of code
                space.mapPage(p, uintptr(n.Vaddr+i), PAGE_RW | PAGE_PERM_USER)
            }
            if n.Filesz < n.Memsz {
                for i:= (n.Filesz + PAGE_SIZE - 1) &^ (PAGE_SIZE - 1); i < n.Memsz; i += PAGE_SIZE {
                    p := allocPage()
                    Memclr(p, PAGE_SIZE)
                    space.mapPage(p, uintptr(n.Vaddr+i), PAGE_RW | PAGE_PERM_USER)
                }
            }
            //text_mode_print_hex32(n.Vaddr)
            //text_mode_print(" ")
            //text_mode_print_hex32(n.Memsz)
            //text_mode_print(" ")
            //text_mode_print_hex32(uint32(len(target)))
            //text_mode_print(" ")
            //text_mode_print_hex(uint8(target[0]))
            //target[0] = 0x42
            //text_mode_print(" ")
            //text_mode_print_hex32(module.Start + n.Off)
            //text_mode_print(" ")
            //text_mode_print_hex32(n.Filesz)
            //text_mode_print(" ")
            //text_mode_print_hex32(uint32(len(contents)))
            //text_mode_print(" ")
            //text_mode_print_hex(uint8(contents[0]))
            //text_mode_println("")

            //text_mode_print_hex(uint8(target[0]))
        }
    }
    return elfHeader.Entry

}
