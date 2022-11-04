package main

import (
    "syscall"
)

const (
    COM1_PORT uint16 = 0x3F8
)

var (
    serialDevice UARTSerialDevice;
)

type UARTSerialDevice struct {
    BasePort uint16;
    initialized bool;
}

func (d UARTSerialDevice) isInitialized() bool {
    return d.initialized;
}

func (d UARTSerialDevice) is_transmit_empty() bool {
    return d.initialized && Inb(d.BasePort + 5) & 0x20 != 0;
}

func (d UARTSerialDevice) WriteChar(arg byte) {
    if d.initialized == false { return; }
    for (d.is_transmit_empty() == false){}
    Outb(d.BasePort,arg);
}

func (d UARTSerialDevice) Write(arr []byte) (int,error) {
    if d.initialized == false { kernelPanic("DEBUG: Serial Device not initialized"); }
    for _, v := range arr {
        if v == '\n'{
            // Make serial look correct
            d.WriteChar('\r')
        }
        d.WriteChar(v)
    }
    return len(arr), nil
}

func (d *UARTSerialDevice) Initialize() syscall.Errno {
    if d.BasePort == 0 { return syscall.ENOSYS; }
    Outb(d.BasePort + 1, 0x00);    // Disable all interrupts
    Outb(d.BasePort + 3, 0x80);    // Enable DLAB (set baud rate divisor)
    Outb(d.BasePort + 0, 0x03);    // Set divisor to 3 (lo byte) 38400 baud
    Outb(d.BasePort + 1, 0x00);    //                  (hi byte)
    Outb(d.BasePort + 3, 0x03);    // 8 bits, no parity, one stop bit
    Outb(d.BasePort + 2, 0xC7);    // Enable FIFO, clear them, with 14-byte threshold
    Outb(d.BasePort + 4, 0x0B);    // IRQs enabled, RTS/DSR set
    Outb(d.BasePort + 4, 0x1E);    // Set in loopback mode, test the serial chip
    Outb(d.BasePort + 0, 0xAE);    // Test serial chip (send byte 0xAE and check if serial returns same byte)

    // Check if serial is faulty (i.e: not same byte as sent)
    if(Inb(d.BasePort + 0) != 0xAE) {
       return syscall.ENOSYS;
    }

    // If serial is not faulty set it in normal operation mode
    // (not-loopback with IRQs enabled and OUT#1 and OUT#2 bits enabled)
    Outb(d.BasePort + 4, 0x0F);
    d.initialized = true;
    return ESUCCESS;

}

func InitSerialDevice(){
    serialDevice = UARTSerialDevice{BasePort: COM1_PORT}
    err := serialDevice.Initialize()
    if err != ESUCCESS {
        kernelPanic("Could not initialize serial device");
    }
}
