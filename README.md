# MyFiNEmu
MyFiNEmu stands for "My First NES Emulator".  This is my pet project
to help teach myself Go and learn about console emulation.

# Implementation plan
1. Core CPU support
   1. Basic (documented) CPU instructions
   2. CPU registers
   3. Basic memory access
   4. Basic support for running a main loop
   5. Test harness to do run instructions and check 
      resulting registers and/or memory locations
2. ROM image support
   1. Read uncompressed ROM images in iNES format.
   2. Read data from virtual game cartridge.
   3. Implement logging/debugging so that we can inspect 
      system state and verify that games are running.
3. Video support
4. Audio support

# NES Information
* [Writing a NES Emulator in Rust](https://bugzmanov.github.io/nes_ebook/)
* [NES Memory Map](https://www.nesdev.org/wiki/CPU_memory_map)
* [6502 Guide](https://www.nesdev.org/obelisk-6502-guide/)
* [Easy 6502](https://skilldrick.github.io/easy6502/)
