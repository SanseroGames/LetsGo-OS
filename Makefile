define newline


endef

BUILD_DIR := build
USR_BUILD_DIR := $(BUILD_DIR)/usr

LD := ld
CC := gcc
CXX := g++
AS := nasm

GOOS := linux
GOARCH := 386
GOROOT := $(shell go env GOROOT)
ARCH := x86

RUST_TARGET := i686-unknown-linux-musl

LD_FLAGS := -n -melf_i386 -T arch/$(ARCH)/script/linker.ld -static --no-ld-generated-unwind-info
AS_FLAGS := -g -f elf32 -F dwarf -I arch/$(ARCH)/asm/
CC_FLAGS := -static -fno-inline -O0
CXX_FLAGS := -static -fno-inline -O0

kernel_target :=$(BUILD_DIR)/kernel-$(ARCH).bin
iso_target := $(BUILD_DIR)/kernel-$(ARCH).iso

disk_image := disk/file.img

asm_src_files := $(wildcard arch/$(ARCH)/asm/*.s)
asm_obj_files := $(patsubst arch/$(ARCH)/asm/%.s, $(BUILD_DIR)/arch/$(ARCH)/asm/%.o, $(asm_src_files))

usr_go_apps_src := $(wildcard usr/*/main.go)
usr_go_apps_obj := $(patsubst usr/%/main.go, $(USR_BUILD_DIR)/%.o, $(usr_go_apps_src))
usr_c_apps_src := $(wildcard usr/*/main.c)
usr_c_apps_obj := $(patsubst usr/%/main.c, $(USR_BUILD_DIR)/%.o, $(usr_c_apps_src))
usr_cpp_apps_src := $(wildcard usr/*/main.cpp)
usr_cpp_apps_obj := $(patsubst usr/%/main.cpp, $(USR_BUILD_DIR)/%.o, $(usr_cpp_apps_src))
usr_rust_apps_src := $(wildcard usr/*/Cargo.toml)
usr_rust_apps_obj := $(patsubst usr/%/Cargo.toml, $(USR_BUILD_DIR)/%.o, $(usr_rust_apps_src))

.PHONY: kernel usr iso

kernel: $(kernel_target)

$(kernel_target): $(asm_obj_files) go.o
	@echo "[$(LD)] linking kernel-$(ARCH).bin"
	@$(LD) $(LD_FLAGS) -o $(kernel_target) $(asm_obj_files) $(BUILD_DIR)/go.o

$(USR_BUILD_DIR)/%.o: usr/%/main.go
	@echo "[go] Building $@ (from $<)"
	@GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o $(USR_BUILD_DIR)/$(basename $(notdir $@)) $<

$(USR_BUILD_DIR)/%.o: usr/%/Cargo.toml
	@echo "[Rust] Building $@ (from $<)"
	@cargo build --target=${RUST_TARGET} --manifest-path $<
	@cp $(dir $<)/target/${RUST_TARGET}/debug/$(basename $(notdir $@)) $(USR_BUILD_DIR)/$(basename $(notdir $@))

$(USR_BUILD_DIR)/%.o: usr/%/main.c
	@echo "[CC] Building $@ (from $<)"
	@$(CC) -m32 $(CC_FLAGS) -o $(USR_BUILD_DIR)/$(basename $(notdir $@)) $<

$(USR_BUILD_DIR)/%.o: usr/%/main.cpp
	@echo "[C++] Building $@ (from $<)"
	@$(CXX) -m32 $(CXX_FLAGS) -o $(USR_BUILD_DIR)/$(basename $(notdir $@)) $<


usr: $(usr_go_apps_obj) $(usr_c_apps_obj) $(usr_rust_apps_obj) $(usr_cpp_apps_obj)

go.o:
	@mkdir -p $(BUILD_DIR)

	@echo "[go] compiling go kernel sources into a standalone .o file"
	@# build/go.o is a elf32 object file but all Go symbols are unexported. Our
	@# asm entrypoint code needs to know the address to 'main.main' and 'runtime.g0'
	@# so we use objcopy to globalize them
	@GOARCH=$(GOARCH) GOOS=$(GOOS) go build -ldflags='-buildmode=c-archive' -o $(BUILD_DIR)/go.o
	@echo "[objcopy] globalizing symbols {runtime.g0, main.main} in go.o"
	@objcopy \
                --globalize-symbol runtime.g0 \
                --globalize-symbol main.main \
                $(BUILD_DIR)/go.o $(BUILD_DIR)/go.o

$(BUILD_DIR)/arch/$(ARCH)/asm/%.o: arch/$(ARCH)/asm/%.s
	@mkdir -p $(shell dirname $@)
	@echo "[$(AS)] $<"
	@$(AS) $(AS_FLAGS) $< -o $@

iso: $(iso_target)

$(iso_target): $(kernel_target) usr
	@echo "[grub] building ISO kernel-$(ARCH).iso"

	@mkdir -p $(BUILD_DIR)/isofiles/boot/grub
	@mkdir -p $(BUILD_DIR)/isofiles/usr/
	@cp $(kernel_target) $(BUILD_DIR)/isofiles/boot/kernel.bin

	@cp $(wildcard $(USR_BUILD_DIR)/*) $(BUILD_DIR)/isofiles/usr/
	@a="$$(./dup.sh $(patsubst $(USR_BUILD_DIR)/%,/usr/%, $(wildcard $(USR_BUILD_DIR)/*)))"; sed -e "s#{MODULES}#$$a#g" arch/x86/script/grub.cfg.tpl | tr '@' '\n' > arch/x86/script/grub.cfg
	@cp arch/$(ARCH)/script/grub.cfg $(BUILD_DIR)/isofiles/boot/grub

	@grub-mkrescue -o $(iso_target) $(BUILD_DIR)/isofiles 2>&1 | sed -e "s/^/  | /g"
	@rm -r $(BUILD_DIR)/isofiles

run: iso
	qemu-system-i386 -d int,cpu_reset -no-reboot -cdrom $(iso_target) \
		-hda $(disk_image) -boot order=dc

# When building gdb target disable optimizations (-N) and inlining (l) of Go code
gdb: GC_FLAGS += -N -l
gdb: iso
	qemu-system-i386 -d int,cpu_reset -s -S -cdrom $(iso_target) \
		-hda $(disk_image) -boot order=dc &
	sleep 1
	echo $(GOROOT)
	gdb \
	    -ex 'add-auto-load-safe-path $(pwd)' \
		-ex 'set arch i386:intel' \
	    -ex 'file $(kernel_target)' \
	    -ex 'target remote localhost:1234' \
	    -ex 'set arch i386:intel' \
		-ex 'source $(GOROOT)/src/runtime/runtime-gdb.py' \
	@killall qemu-system-i386 || true
