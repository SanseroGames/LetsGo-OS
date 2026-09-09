package main

import (
	"unsafe"

	"github.com/sanserogames/letsgo-os/kernel"
)

// Multiboot magic number
const multibootMagic = 0x36d76289

// 16KiB
const stackSize = 16 * 1024

// Reserve 16K for the stack + 16 to enforce alignment to 16 byte boundaries
var stack [stackSize + 16]byte

var unsupportedBootloaderMsg = "[rt0] kernel not loaded by multiboot-compliant bootloader\x00"

const kCS_SELECTOR = kernel.KCS_SELECTOR
const kDS_SELECTOR = kernel.KDS_SELECTOR
const kGS_SELECTOR = kernel.KGS_SELECTOR

// Type aliases so types can be used in assembly
type gdtDescriptor kernel.GdtDescriptor

var initialGdt = [4]kernel.GdtEntry{
	// nullSegment. See https://wiki.osdev.org/GDT_Tutorial#What_to_Put_In_a_GDT why the null descriptor is necessary
	{LimitLow: 0, BaseLow: 0, BaseMid: 0, Access: 0, LimitHighAndFlags: 0, BaseHigh: 0},
	// CS Segment at kernel.KCS_index
	{LimitLow: uint16(0xFFFF), BaseLow: 0, BaseMid: 0, Access: kernel.SEG_EXEC | kernel.SEG_R | kernel.SEG_PRESENT | kernel.SEG_NORMAL, LimitHighAndFlags: 0xF | kernel.SEG_GRAN_4K_PAGE | kernel.SEG_BIG_MODE, BaseHigh: 0},
	// DS Segment at kernel.KCS_index
	{LimitLow: uint16(0xFFFF), BaseLow: 0, BaseMid: 0, Access: kernel.SEG_NOEXEC | kernel.SEG_W | kernel.SEG_PRESENT | kernel.SEG_NORMAL, LimitHighAndFlags: 0xF | kernel.SEG_GRAN_4K_PAGE | kernel.SEG_BIG_MODE, BaseHigh: 0},
	// GS Segment at kernel.KCS_index
	// Address to TCB will be set in assembly
	{LimitLow: uint16(0xFFFF), BaseLow: 0, BaseMid: 0, Access: kernel.SEG_NOEXEC | kernel.SEG_W | kernel.SEG_PRESENT | kernel.SEG_NORMAL, LimitHighAndFlags: 0xF | kernel.SEG_GRAN_BYTE | kernel.SEG_BIG_MODE, BaseHigh: 0},
}

var initialGdtDescriptor = kernel.GdtDescriptor{
	GdtSize:        uint16(unsafe.Sizeof(initialGdt)),
	GdtAddressLow:  0, // uint16(uintptr(unsafe.Pointer(&initialGdt))), // Setting the address like this does not work
	GdtAddressHigh: 0, // uint16(uintptr(unsafe.Pointer(&initialGdt)) >> 16), // Setting the address like this does not work
}
