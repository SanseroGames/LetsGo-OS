# Let's Go OS

This is a little experiment I did for myself where I tried to make an OS in the 
Go programming language. 

## Go-al

The goal is not to create a "fully" functional operating system (whatever this means). 
I only want to implement a FAT32-filesystem and use that. I did one once (in C) but I
cannot continue working on that one and implementing a filesystem for Linux is boring 
so I wrote my own OS for this. Also for a long time I wanted to start learning Go.

## Why Go?

Go is not really designed to be used to write an operating system with it. But it is 
still possoble, so why not do it? I just did not want to use C. The alternative would have been Rust but Rust is designed to be also used as a programming language for operating systems so Go is more a challange and thus more interesting. Still I plan to use Rust in my userspace.

## Inspirations used for this projects

- [bare-metal-gophers](https://github.com/achilleasa/bare-metal-gophers) I used this project as a starting point for my project. The project was a demo project for a talk at GolangUK2017.
- [gopher-os](https://github.com/gopher-os/gopher-os)
- [eggos](https://github.com/icexin/eggos)
- and of course the [OSDev-Wiki](https://wiki.osdev.org)
