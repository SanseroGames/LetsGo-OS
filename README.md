# Let's Go OS

This is a little experiment I'm working on where I try to create an OS kernel in the Go programming language.

## Go-al

The goal is not to create a "fully" functional operating system (whatever this means) but learn more about operating systems and how they are made.

Originally I planned to implement a FAT32-filesystem. I implemented one once (in C) but I
cannot continue to work on project due to missing harware. Of course you could easily implement
a filesystem in Linux, but that's kinda boring.

So instead I started to write my own OS for this project. It is also a way to start learning Go. Well, at least somewhat.

## Why Go?

When I got the idea to write an operating system, I knew that I don't want to use C, as I already had some experience writing an OS in C. I deliberated
whether to use Rust or Go to write my OS. The deciding factor in the end was, that Go would be the bigger challenge.

Now, you might think that it is not that good of an idea to use Go to write a kernel, and you would be right. The Go-runtime is designed to be used in
userspace of an operating system. There are a lot of assumptions that this implies, like the ability to invoke syscalls and the use of dynamic memory
allocation. In a kernel you cannot rely on an OS to handle this for you, as you are the OS. However, with carefull programming and giving up of certain
Go-features you can do it, so why not? It is a challenge.

## Design

One goal of this project is to achieve stuff fast, so I did not spent time designing the OS. To speed up the whole process, I decided to use the syscall
interface of Linux. This way I would not have to write my own userspace libraries and could use any programming language I wanted in my userspace. It
has quite a few drawbacks though. While I do the syscalls in the same way as Linux, this requires me to use all the assumptions of Linux as well. For
example, starting a process requries me to set up it's memory space the same way as Linux would, the syscalls need to return values that are expected on
a Linux kernel and the kernel design is assumed to be monolithical.

I'm not yet sure if I stick with this design. Another idea I had is to write my own ABI but also write a Linux-syscall wrapper library that would
intercept Linux syscalls and translate it to my OS.


## Requirements
- golang (duh)
- nasm (used to build bootstrap x86 assembly)
- gcc-multilib (OS targets 32bit so this lib is needed when compiled on 64bit platforms for user programs written in C)
- g++-multilib (same as above but for C++ programs)
- Rust (for user programs written in Rust)
- qemu-system-i386 (Emulator used in the makefile. You can use a different x86 emulator if you want)
- grub2 (Used to generate the boot media)
- xorriso (used as well to generate the boot media)

## Inspirations used for this projects

- [bare-metal-gophers](https://github.com/achilleasa/bare-metal-gophers) I used this project as a starting point for my project. The project was a demo project for a talk at GolangUK2017.
- [gopher-os](https://github.com/gopher-os/gopher-os)
- [eggos](https://github.com/icexin/eggos)
- and of course the [OSDev-Wiki](https://wiki.osdev.org)
