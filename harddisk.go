package main

import (
)

type AtaDrive struct {
    IOBase uint16
    ControlBase uint16
    Initialized bool
    IsSlave bool
    IdentifyData [512]byte
    IdentifyStruct AtaIdentify
}

type AtaIdentify struct {
    generalCfg [2]uint8
    numCylinders [2]uint8
    specialConfiguration [2]uint8
    numHeads [2]uint8
    retired1 [4]uint8 // Not filled
    NumSectorsPerTrack [2]uint8
    vendorUnique1 [6]uint8
    serialNumber [20]uint8
    retired2 [2]uint8 // Not filled
    obsolete1 [2]uint8 // Not filled
    firmwareRevision [8]uint8
    modelNumber [40]uint8
    maximumBlockTransfer uint8
    vendorUnique2 uint8
    trustedComputing [2]uint8
    capabilities1 [2] uint8
    capabilities2 [2] uint8
    obsoleteWords51 [4]uint8 // Not filled
    translationFieldsValid [2]uint8
}

func (i* AtaIdentify) InitFromBytes(data []byte){
    copy(i.generalCfg[:], data[0:2])
    copy(i.numCylinders[:], data[2:4])
    copy(i.specialConfiguration[:], data[4:6])
    copy(i.numHeads[:], data[6:8])
    copy(i.NumSectorsPerTrack[:], data[12:14])
    copy(i.vendorUnique1[:], data[14:20])
    copy(i.serialNumber[:], data[20:40])
    copy(i.firmwareRevision[:], data[40:48])
    copy(i.modelNumber[:], data[48:88])
    i.maximumBlockTransfer = data[88]
    i.vendorUnique2 = data[89]
    copy(i.trustedComputing[:], data[90:92])
    copy(i.capabilities1[:], data[92:94])
    copy(i.capabilities2[:], data[94:96])
    copy(i.translationFieldsValid[:], data[96:98])
}

func (i* AtaIdentify) printInfos() {
    printInfoHelper("generalCfg", i.generalCfg[:])
    printInfoHelper("numCylinders", i.numCylinders[:])
    printInfoHelper("specialConfiguration", i.specialConfiguration[:])
    printInfoHelper("numHeads", i.numHeads[:])
    printInfoHelper("NumSectorsPerTrack", i.NumSectorsPerTrack[:])
    printInfoHelper("vendorUnique1", i.vendorUnique1[:])
    printInfoHelper("serialNumber", i.serialNumber[:])
    printInfoHelper("firmwareRevision", i.firmwareRevision[:])
    printInfoHelper("modelNumber", i.modelNumber[:])
    printInfoHelper("maximumBlockTransfer", []uint8 {i.maximumBlockTransfer})
    printInfoHelper("vendorUnique2", []uint8 {i.vendorUnique2})
    printInfoHelper("trustedComputing", i.trustedComputing[:])
    printInfoHelper("capabilities1", i.capabilities1[:])
    printInfoHelper("capabilities2", i.capabilities2[:])
    printInfoHelper("translationFieldsValid", i.translationFieldsValid[:])

}

func printInfoHelper(name string, data []uint8){
    text_mode_print(name)
    text_mode_print(": ")
    for c,i := range data {
        text_mode_print_hex(i)
        text_mode_print(" ")
        if(c % 22 == 21 && c > 0){
            text_mode_println("")
        }
    }
    text_mode_println("")
}

const (
    ataTimeout = 5000
)

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
    ataStatusBusy = 0x80
    ataStatusReady = 0x40
    ataStatusDiskFail = 0x20
    ataStatusSrv = 0x10
    ataStatusDRQ = 0x08
    ataStatusCorr = 0x04
    ataStatusIdx = 0x02
    ataStatusError = 0x01
)

const (
    ataResetCommand = 4
    ataNoInterruptsCommand = 1
    ataReadCommand = 0x20
    ataWriteCommand = 0x30
    ataFlushCacheCommand = 0xE7
    ataIdentifyCommand = 0xEC
)

func (d *AtaDrive) Initialize() {
    if !d.Reset() { return }

    driveSelect := 0xE0
    if d.IsSlave {
        driveSelect = 0xF0
    }

    a := Inb(d.IOBase + ataLbaMid)
    b := Inb(d.IOBase + ataLbaHi)
    if (a == 0x14 && b == 0xeb) {
        /* This is a magic identifier for ATAPI devices! */
        return
    } else if (a == 0x69 && b == 0x96) {
        /* This is a magic identifier for SATA-ATAPI devices! */
        return
    } else if (a == 0x3c && b == 0xc3) {
        /* This is a magic identifier for SATA devices! */
    } else if (a == 0 && b == 0) {
        /* Plain old ATA disk */
    } else if (a == 0xff && b == 0xff) {
        /* Nothing there */
        return;
    } else {
        text_mode_print_errorln("Unknown device type")
        return
    }

    Outb(d.IOBase + ataDriveAndHead, uint8(driveSelect))
    d.delay()
    Outb(d.IOBase + ataCommandRegister, ataIdentifyCommand)
    d.delay()
    for i:=0; i < ataTimeout; i++ {
       Inb(d.IOBase + ataStatusRegister)
    }
    status := Inb(d.IOBase + ataStatusRegister)
    if status == 0 {
       return
    }
    for c := 0; c < 256; c++ {
        w := Inw(d.IOBase + ataDataRegister)
        d.IdentifyData[c*2] = uint8(w)
        d.IdentifyData[c*2+1] = uint8(w >> 8)
    }
    d.IdentifyStruct.InitFromBytes(d.IdentifyData[:])
    d.Initialized = true
}

func (d *AtaDrive) delay() {
    ASR := d.ControlBase + ataAlternativeStatusRegister
    Inb(ASR)
    Inb(ASR)
    Inb(ASR)
    Inb(ASR)
}

func (d *AtaDrive) Reset() bool {
    driveSelect := 0xE0
    if d.IsSlave {
        driveSelect = 0xF0
    }
    Outb(d.IOBase + ataDriveAndHead, uint8(0xa0))
    d.delay()
    DCR := d.ControlBase + ataDeviceControlRegister
    ASR := d.ControlBase + ataAlternativeStatusRegister
    // Do a software reset
    Outb(DCR, ataResetCommand | ataNoInterruptsCommand)
    for i:=0; i < 10000; i++ {
        Inb(ASR)
    }
    // Clear it again
    Outb(DCR, 0)
    for i:=0; i < 10000; i++ {
        Inb(ASR)
    }
    Inb(d.IOBase + ataErrorRegister)
    Outb(d.IOBase + ataDriveAndHead, uint8(driveSelect))
    d.delay()
    t := 50000
    for ; t > 0; t-- {
        status := Inb(d.IOBase + ataStatusRegister)
        if (status & ataStatusBusy) == 0 {
            break
        }
        d.delay()
    }
    if t == 0 {
        text_mode_print_errorln("Timeout resetting drive")
        return false
    }
    return true
}

// TODO: Explicit length?
func (d *AtaDrive) WriteSectors(address int, buffer[]byte) {
    text_mode_print_errorln("Write does not work :(")
    return
    if !d.Initialized {
        return
    }
    // TODO: Padd with zeros
    if len(buffer) < 512 {
        return
    }
    count := len(buffer) / 512
    // TODO: This is dangerous
    text_mode_print("Writing 0x")
    text_mode_print_hex(uint8(count))
    text_mode_println(" sectors")
    for i:=0; i<count; i++ {
        d.workSectors(address+i, 1, buffer, true)
    }
}

func (d *AtaDrive) ReadSectors(address int, count uint8, buffer[]byte) {
    if !d.Initialized {
        return
    }
    if int(count)*512 > len(buffer) {
        return
    }
    for i:=0; i < int(count); i++ {
        d.workSectors(address, 1, buffer[i*512:(i+1)*512], false)
    }
}

// Assumes disk is initialized
func (d *AtaDrive) workSectors(address int, count uint8, buffer[]byte, write bool) {
    driveSelect := 0xE0
    if d.IsSlave {
        driveSelect = 0xF0
    }

    Outb(d.IOBase + ataSectorCount, count)
    Outb(d.IOBase + ataLbaLow, uint8(address))
    Outb(d.IOBase + ataLbaMid, uint8(address >> 8))
    Outb(d.IOBase + ataLbaHi, uint8(address >> 16))
    Outb(d.IOBase + ataDriveAndHead, uint8(driveSelect | ((address >> 24) & 0x0F)))
    if write {
        Outb(d.IOBase + ataCommandRegister, ataWriteCommand)
    } else {
        Outb(d.IOBase + ataCommandRegister, ataReadCommand)
    }

    i := 0
    hasError := false
    for {
        s := Inb(d.IOBase + ataStatusRegister);
        if(i > 1000) {
            text_mode_print_errorln("Timeout trying to read disk")
            text_mode_print_hex(s)
            text_mode_println("")
            d.Reset()
            return
        }
        if i > 4 && s & 0x21 != 0 {
            hasError = true
            break
        }
        if(s & 0x88 == 0x08) {break}
        i++
    }
    if hasError {
        text_mode_print_errorln("Error while trying to execute disk command")
        return
    }
    offset := 0
    for n := 0; n < int(count); n++ {
        for c := 0; c < 256; c++ {
            if write {
                w := uint16(buffer[offset]) | (uint16(buffer[offset+1]) << 8)
                Outw(d.IOBase + ataDataRegister, w)
            } else {
                w := Inw(d.IOBase + ataDataRegister)
                buffer[offset] = uint8(w)
                buffer[offset+1] = uint8(w >> 8)
            }
            offset+=2
        }
        for s := Inb(d.IOBase + ataStatusRegister); s & 0x80 == 0x80; {}
    }
    if write {
        // Flush cache
        Outb(d.IOBase + ataCommandRegister, ataFlushCacheCommand)
        // Wait for operation to complete
        //i := 0
        //Inb(d.IOBase + ataStatusRegister)
        //Inb(d.IOBase + ataStatusRegister)
        //Inb(d.IOBase + ataStatusRegister)
        //Inb(d.IOBase + ataStatusRegister)
        //for s := Inb(d.IOBase + ataStatusRegister); s & 0x80 == 0x80; {
        //    if i > 1000000 {
        //        text_mode_print_errorln("Timeout while flushing disk cache")
        //        return
        //    }
        //    i++
        //}
    }
    //if t & 0x80 == 0x80 {
        // TODO: AAAA I don't know why this is happening!!!!
        //d.Reset()
    //}
}

var firstDrive AtaDrive = AtaDrive {
    IOBase: 0x1f0,
    ControlBase: 0x3F6,
    IsSlave: false,
}

func InitATA(){
    firstDrive.Initialize()
    firstDrive.IdentifyStruct.printInfos()
    //for i:=0; i < 8; i++ {
    //    testReadAndWrite()
    //    delay(2000)
    //}
    //for c,i := range firstDrive.IdentifyData {
    //    text_mode_print_hex(i)
    //    text_mode_print(" ")
    //    if(c % 22 == 21 && c > 0){
    //        text_mode_println("")
    //    }
    //}
}

var hdBuf [1024]byte
var hdBuf2 [1024]byte

func testReadAndWrite(){
    for i := range hdBuf {
        hdBuf[i] = byte(i%256)
    }
    copy(hdBuf2[:], hdBuf[:])
    text_mode_println("Writing Data")
    //firstDrive.WriteSectors(0, hdBuf[:512])
    //firstDrive.WriteSectors(1, hdBuf[512:])
    for i := range hdBuf {
        hdBuf[i] = 0x42
    }
    text_mode_println("Reading Data")
    firstDrive.ReadSectors(0, 1, hdBuf[:512])
    firstDrive.ReadSectors(1, 1, hdBuf[512:])
    //firstDrive.ReadSectors(0, 2, hdBuf[:])
    for c := range hdBuf {
        text_mode_print_hex(hdBuf[c])
        text_mode_print(" ")
        if(c % 512 == 511 && c > 0){
            text_mode_println("")
        }
    }
    text_mode_println("")
    for c := range hdBuf2 {
        text_mode_print_hex(hdBuf2[c])
        text_mode_print(" ")
        if(c % 512 == 511 && c > 0){
            text_mode_println("")
        }
    }
    text_mode_println("")
    noMatch := false
    for i := range hdBuf2 {
        if(hdBuf[i] != hdBuf2[i]){
            text_mode_print_errorln("Buffers don't match!!!")
            text_mode_print_hex16(uint16(i))
            text_mode_println("")
            noMatch = true
            break
        }
    }
    if(!noMatch){
        text_mode_println("Buffers match")
    }

}
