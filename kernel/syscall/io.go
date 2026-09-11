package syscall

import (
	"syscall"
	"unsafe"

	"github.com/sanserogames/letsgo-os/kernel"
	"github.com/sanserogames/letsgo-os/kernel/log"
	"github.com/sanserogames/letsgo-os/kernel/mm"
	"github.com/sanserogames/letsgo-os/kernel/utils"
)

type ioVec struct {
	iovBase uintptr /* Starting address */
	iovLen  uint32  /* Number of bytes to transfer */
}

type directoryEntryHead struct {
	inodeNr      uint64
	fsOffset     uint64
	direntLength uint16
	fileType     byte
	// append filename as null terminated C string and update direntLength
	fileNameStart byte
}

func linuxOpenSyscall(args syscallArgs) (uint32, syscall.Errno) {
	//path := args.arg1
	flags := args.arg2
	addr, ok := kernel.CurrentThread.Domain.MemorySpace.GetPhysicalAddress(uintptr(flags))
	if !ok {
		return 0, syscall.EFAULT
	}
	s := utils.CString(addr)
	if PRINT_SYSCALL {
		log.KDebugLn("[SYS-OPEN] ", s)
	}
	//text_mode_println(s)
	return 0, syscall.ENOSYS
	//printRegisters(currentInfo, Regs)

}

// pid: 2 tid: 3 :: open at syscall (295 hex(0x127))
// |- arg1:0xffffff9c, arg2:0x8476030, arg3:0x8000, arg4:0x0, arg5:0x0, arg6:0x0
func linuxOpenAtSyscall(args syscallArgs) (uint32, syscall.Errno) {
	fd := args.arg1
	path := args.arg2
	flags := args.arg3
	pathaddr, ok := kernel.CurrentThread.Domain.MemorySpace.GetPhysicalAddress(uintptr(path))
	if !ok {
		return 0, syscall.EFAULT
	}
	s1 := utils.CString(pathaddr)
	if PRINT_SYSCALL {
		log.KDebugLn("[SYS-OPENAT] fd:", fd)
		log.KDebugLn("[SYS-OPENAT] path:", s1)
		log.KDebugLn("[SYS-OPENAT] flags:", flags)
	}

	if s1 == "/dev/null" {
		return 42, ESUCCESS
	}

	if flags&syscall.O_DIRECTORY != 0 {
		// pretend to show the modules direcotry
		return 67, ESUCCESS
	}

	//text_mode_println(s)
	return 0, syscall.ENOSYS
	//printRegisters(currentInfo, Regs)

}

func linuxGetdents64Syscall(args syscallArgs) (uint32, syscall.Errno) {
	fd := args.arg1
	buffer := uintptr(args.arg2)
	// count := args.arg3

	if kernel.CurrentThread.Domain.TmpReadDirectory {
		kernel.CurrentThread.Domain.TmpReadDirectory = false
		return 0, ESUCCESS
	}

	if fd != 67 {
		return 0, syscall.ENOTDIR
	}

	writtenBytes := 0

	for idx, module := range kernel.LoadedModules {

		if module.Type != 3 {
			continue
		}

		entry := directoryEntryHead{
			inodeNr:      uint64(idx),
			fsOffset:     0xdeadbeef,
			direntLength: 0,
			fileType:     0,
		}

		filename := module.Cmdline()
		filenameSlice := unsafe.Slice(unsafe.StringData(filename), len(filename))
		entrySize := 19 //unsafe.Sizeof(entry)

		entry.direntLength = uint16(entrySize + len(filenameSlice))

		err := mm.WriteTypeToUserSpace(&entry, buffer+uintptr(writtenBytes), &kernel.CurrentThread.Domain.MemorySpace)
		if err != 0 {
			return 0, err
		}

		err = kernel.CurrentThread.Domain.MemorySpace.WriteBytesToUserSpace(buffer+uintptr(writtenBytes+entrySize), filenameSlice)
		if err != 0 {
			return 0, err
		}

		writtenBytes += int(entry.direntLength)
	}

	kernel.CurrentThread.Domain.TmpReadDirectory = true

	return uint32(writtenBytes), ESUCCESS
}

func linuxWriteVSyscall(args syscallArgs) (uint32, syscall.Errno) {
	fd := args.arg1
	arr := uintptr(args.arg2)
	count := args.arg3
	if fd < 1 || fd > 2 {
		return 0, syscall.EBADF
	}

	if count <= 0 {
		return 0, syscall.EINVAL
	}

	processed := 0
	printed := uint32(0)
	for item, err := range mm.IterateUserSpaceType[ioVec](arr, &kernel.CurrentDomain.MemorySpace) {
		if err != ESUCCESS {
			return 0, err
		}
		written, err := writeSyscallWriteData(fd, item.iovBase, item.iovLen)
		if err != ESUCCESS {
			return 0, err
		}
		printed += written
		processed++
		if processed == int(count) {
			break
		}
	}
	return printed, ESUCCESS
}

func linuxWriteSyscall(args syscallArgs) (uint32, syscall.Errno) {
	fd := args.arg1
	text := uintptr(args.arg2)
	length := args.arg3
	if PRINT_SYSCALL {
		log.KDebugLn("FD: ", fd, " text: ", text, " length: ", length)
	}
	if fd < 1 || fd > 2 {
		return 0, syscall.EBADF
	}
	return writeSyscallWriteData(fd, text, length)
}

func writeSyscallWriteData(fd uint32, text uintptr, length uint32) (uint32, syscall.Errno) {
	var buf [100]byte

	s := unsafe.String(&buf[0], len(buf))

	index := 0
	writeLen := uint32(0)
	for value, err := range kernel.CurrentDomain.MemorySpace.IterateUserSpace(text) {
		if err != ESUCCESS {
			return 0, err
		}
		buf[index] = value
		index++
		if index == len(buf) {
			if fd == 2 {
				log.KError(s)
			} else {
				log.KPrint(s)
			}
			index = 0
		}
		writeLen++
		if writeLen == length {
			break
		}
	}
	rest := unsafe.String(&buf[0], index)
	if fd == 2 {
		log.KError(rest)
	} else {
		log.KPrint(rest)
	}
	return uint32(writeLen), ESUCCESS
}

func linuxReadSyscall(args syscallArgs) (uint32, syscall.Errno) {
	fd := args.arg1
	buf := args.arg2
	count := args.arg3

	var num uint32 = 0

	var charBuf [1]byte
	char := charBuf[:]

	if fd == 0 {
		for num == 0 {
			for !kernel.SerialDevice.HasReceivedData() {
				kernel.Yield()
			}

			for kernel.SerialDevice.HasReceivedData() && num < count {
				char[0] = kernel.SerialDevice.Read()
				// arr[num] = char
				kernel.CurrentThread.Domain.MemorySpace.WriteBytesToUserSpace(uintptr(buf), char)
				num++
			}
		}
	}
	// I don't use keyboard anymore, since I have serial.
	// When I implmentf device drivers, I might add this again
	// else if fd == 42 {
	// 	for num == 0 {
	// 		for buffer.Len() == 0 {
	// 			kernel.Block()
	// 		}

	// 		for buffer.Len() > 0 && num < count {
	// 			raw_key := buffer.Pop().Keycode
	// 			pressed := raw_key&0x80 == 0
	// 			key := raw_key & 0x7f
	// 			if pressed {
	// 				arr[num] = translateKeycode(key)
	// 				num++
	// 			}
	// 		}
	// 	}
	// }
	return num, ESUCCESS
}
