package kernel

import (
	"syscall"
)

const (
	COM1_PORT uint16 = 0x3F8
	COM1_IRQ  uint8  = 0x4
)

var (
	serialDevice UARTSerialDevice
)

type UARTSerialDevice struct {
	BasePort    uint16
	initialized bool
}

func (d UARTSerialDevice) isInitialized() bool {
	return d.initialized
}

func (d UARTSerialDevice) is_transmit_empty() bool {
	return d.initialized && Inb(d.BasePort+5)&0x20 != 0
}

func (d UARTSerialDevice) WriteChar(arg byte) {
	if d.isInitialized() == false {
		return
	}
	for d.is_transmit_empty() == false {
	}
	Outb(d.BasePort, arg)
}

func (d UARTSerialDevice) Write(arr []byte) (int, error) {
	if d.isInitialized() == false {
		kernelPanic("DEBUG: Serial Device not initialized")
	}
	for _, v := range arr {
		if v == '\n' {
			// Make serial look correct
			d.WriteChar('\r')
		}
		d.WriteChar(v)
	}
	return len(arr), nil
}

func (d UARTSerialDevice) HasReceivedData() bool {
	return Inb(d.BasePort+5)&1 == 1
}

func (d UARTSerialDevice) Read() uint8 {
	return Inb(d.BasePort)
}

func (d *UARTSerialDevice) Initialize() syscall.Errno {
	if d.BasePort == 0 {
		return syscall.ENOSYS
	}
	Outb(d.BasePort+1, 0x00) // Disable all interrupts
	Outb(d.BasePort+3, 0x80) // Enable DLAB (set baud rate divisor)
	Outb(d.BasePort+0, 0x03) // Set divisor to 3 (lo byte) 38400 baud
	Outb(d.BasePort+1, 0x00) //                  (hi byte)
	Outb(d.BasePort+3, 0x03) // 8 bits, no parity, one stop bit, disable DLAB
	Outb(d.BasePort+2, 0x07) // Enable FIFO, clear them, with 1-byte threshold
	Outb(d.BasePort+4, 0x1E) // Set in loopback mode, test the serial chip
	Outb(d.BasePort+0, 0xAE) // Test serial chip (send byte 0xAE and check if serial returns same byte)

	// Check if serial is faulty (i.e: not same byte as sent)
	if Inb(d.BasePort+0) != 0xAE {
		return syscall.ENOSYS
	}

	// If serial is not faulty set it in normal operation mode
	// (not-loopback with IRQs enabled and OUT#1 and OUT#2 bits enabled)
	Outb(d.BasePort+4, 0x00)
	// Outb(d.BasePort+1, 0x01) // IRQs enabled, RTS/DSR set

	d.initialized = true
	return ESUCCESS
}

func InitSerialDevice() {
	serialDevice = UARTSerialDevice{BasePort: COM1_PORT}
	err := serialDevice.Initialize()
	if err != ESUCCESS {
		kernelPanic("Could not initialize serial device")
	}
}

func InitSerialDeviceInterrupt() {
	// RegisterPICHandler(COM1_IRQ, testInterrupt4)
	// EnableIRQ(COM1_IRQ)
}
