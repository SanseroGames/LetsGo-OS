package main

func handleKeyboard(){
    keycode := Inb(0x60) // TODO: constant Where to get this?
    // Better would be to create a buffer and then a consumer for the key codes to print...
    if(keycode & 0x80 == 0x80) {return}
    text_mode_print_char(keycode)
    text_mode_print_char(0x20)
    text_mode_print_hex(keycode)
    text_mode_print_char(0x0a)
}

func InitKeyboard(){
    EnableIRQ(1)
    RegisterPICHandler(1, handleKeyboard)
}
