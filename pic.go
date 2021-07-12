package main

const (
    PIC1Port uint16 = 0x20
    PIC1Data uint16 = PIC1Port+1
    PIC2Port uint16 = 0xA0
    PIC2Data uint16 = PIC2Port+1

    PIC1Offset byte = 0x20
    PIC2Offset byte = 0x28

    PIC_ICW1_ICW4   byte = 0x01 // ICW4 will be sent
    PIC_ICW1_Init   byte = 0x10
    PIC_ICW4_8086   byte = 0x01 // Set 8086/88 mode. unset MCS-80/85 mode

    PIC_EOI     byte = 0x20 // End of interrupt
    PIC_ReadIRR byte = 0xa
    PIC_ReadISR byte = 0xb
)

var picHandlers [16]func()

func PICInterruptHandler(info *InterruptInfo, regs *RegisterState){
    irq := info.InterruptNumber - uint32(PIC1Offset)
    if(irq == 7){
        Outb(PIC1Port, PIC_ReadISR)
        res := Inb(PIC1Port)
        if(res & (1 << 7) == 0){
            // Spurious IRQ
            return
        }
    }
    if(irq == 15){
        Outb(PIC1Port, PIC_ReadISR)
        res := Inb(PIC1Port)
        if(res & (1 << 7) == 0){
            // Spurious IRQ
            // PIC1 does not know it is spurious
            Outb(PIC1Port, PIC_EOI)
            return
        }
    }

    picHandlers[irq]()

    if(irq >= 8){
        Outb(PIC2Port, PIC_EOI)
    }
    Outb(PIC1Port, PIC_EOI)
}

func defaultPicHandler(){

}

func InitPIC(){

    Outb(PIC1Port, PIC_ICW1_Init | PIC_ICW1_ICW4)
    Outb(PIC2Port, PIC_ICW1_Init | PIC_ICW1_ICW4)

    Outb(PIC1Data, PIC1Offset)
    Outb(PIC2Data, PIC2Offset)

    Outb(PIC1Data, 0b0100) // Tells PIC1 that there is slave PIC at IRQ2
    Outb(PIC2Data, 2) // Tells slave PIC what IRQ it is on master PIC (binary form)
    Outb(PIC1Data, PIC_ICW4_8086)
    Outb(PIC2Data, PIC_ICW4_8086)

    Outb(PIC1Data, 0xff)
    Outb(PIC2Data, 0xff)

    for i:=0; i < 8; i++{
        SetInterruptHandler(PIC1Offset+byte(i), PICInterruptHandler)
    }
    for i:=0; i < 8; i++{
        SetInterruptHandler(PIC2Offset+byte(i), PICInterruptHandler)
    }
    for i := range picHandlers {
        RegisterPICHandler(uint8(i), defaultPicHandler)
    }
}

func DisableIRQ(irq uint8){
    port := PIC1Data
    if(irq > 7){
        port = PIC2Data
        irq -= 8
    }
    value := Inb(port) | (1 << irq)
    Outb(port, value)
}

func EnableIRQ(irq uint8){
    port := PIC1Data
    if(irq > 7){
        port = PIC2Data
        irq -= 8
    }
    value := Inb(port) &^ (1 << irq)
    Outb(port, value)
}

func RegisterPICHandler(irq uint8, f func()){
    picHandlers[irq] = f
}
