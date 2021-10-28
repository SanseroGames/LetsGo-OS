package main

const (
    PIT_PORT_DATA = 0x40
    PIT_PORT_COMMAND = 0x43
)

func handlePit() {
}

func InitPit() {
    Outb(PIT_PORT_DATA, 0x80);		// Low byte
	Outb(PIT_PORT_DATA, 0x00);	// High byte
    RegisterPICHandler(0, handlePit)
    EnableIRQ(0)
}
