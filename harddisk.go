package main

type AtaDrive struct {
    IOBase uint16
    ControlBase uint16
    Initialized bool
}

const (
    hdFirstATABus = 0x1f0

    ataDataRegister = 0
    ataErrorRegister = 1
    ataSectorCount = 2
    ataLbaLow = 3
    ataLbaMid = 4
    ataLbaHi = 5
    ataDriveAndHead = 6
    ataStatusRegister = 7
    ataCommandRegister = 7

    ataAlternativeStatusRegister = 0
    ataDeviceControlRegister = 0
    ataDriveAddressRegister = 1
)

const (
    ataResetCommand = 4
    ataReadCommand = 0x20
)

func (d *AtaDrive) Reset(){
    DCR := d.ControlBase + ataDeviceControlRegister
    // Do a software reset
    Outb(DCR, ataResetCommand)
    // Clear it again
    Outb(DCR, 0)
    // Do a 400ns delay
    for i:=0; i < 4; i++ {
        Inb(DCR)
    }
    i := 0
    for status := Inb(DCR); (status & 0xc0) != 0x40; {
        if(i > 100){
            text_mode_print_error("Timeout initializing disk")
            return
        }
        i++
    }
    d.Initialized = true
}

func (d *AtaDrive) ReadSectors(address int, count uint8, buffer[]byte) {
    //if(!d.Initialized) {return}
    // Pretent only use master
    Outb(d.IOBase + ataSectorCount, count)
    Outb(d.IOBase + ataLbaLow, uint8(address))
    Outb(d.IOBase + ataLbaMid, uint8(address >> 8))
    Outb(d.IOBase + ataLbaHi, uint8(address >> 16))
    Outb(d.IOBase + ataDriveAndHead, uint8(0xE0 | ((address >> 24) & 0x0F)))
    Outb(d.IOBase + ataCommandRegister, ataReadCommand)

    i := 0
    for {
        if(i > 100) {
            text_mode_print_error("Timeout when trying to read disk")
            return
        }
        s := Inb(d.IOBase + ataStatusRegister);
        if(i > 4){
            // Test Error flag
        }
        if(s & 0x80 == 0) {break}
        i++
    }
    for c := 0; c < 256; c++ {
        w := Inw(d.IOBase + ataDataRegister)
        buffer[c] = uint8(w)
        buffer[c+1] = uint8(w >> 8)
    }
}

var hdBuf [512]byte

var firstDrive AtaDrive = AtaDrive {
    IOBase: 0x1f0,
    ControlBase: 0x3F6,
}

func InitATA(){
    firstDrive.Reset()
    firstDrive.ReadSectors(0, 1, hdBuf[:])
    for c:=0; c < 512; c++ {
        text_mode_print_hex(hdBuf[c])
        text_mode_print(" ")
        if(c % 24 == 23 && c > 0){
            text_mode_println("")
        }
    }
}
