# Good Boy

Good Boy is a Game Boy (DMG) emulator written in Go.
Mostly a work in progress, but good enough to play some popular games (Tetris, Super Mario Land, Donkey Kong, ...).

![goodboy screenshot](https://github.com/user-attachments/assets/21f6b4f8-83fb-45f7-be59-1bb50466a015)

## Usage

From source:
```code 
go run . [rom file]
```

Using a pre-built binary:
```code
goodboy [rom file]
```

Play with <kbd>&larr;</kbd>, <kbd>&uarr;</kbd>, <kbd>&darr;</kbd>, <kbd>&rarr;</kbd>, <kbd>A</kbd>, <kbd>S</kbd>, <kbd>Enter</kbd>, <kbd>R Shift</kbd>.

The emulator has a built-in textual debugger and tracer (use `-debug` and `-trace`).

## Features & TODOs

- [x] CPU, timer, interrupt, graphics, joypad, sound
- [x] Cartridge types: ROM-only, MBC1
- [x] Built-in debugger
- [x] Pass Blargg's cpu_instrs, instr_timing, mem_timing, mem_timing-2
- [x] Pass [dmg-acid2](https://github.com/mattcurrie/dmg-acid2) test
- [ ] PPU timing (currently rendering any line takes 172 dots)
- [ ] Support more cartridge types (MBC3, MBC5, ...)
- [ ] Pass more Blargg tests, Mooneye, etc
- [ ] Game save/restore
- [ ] Gameboy color
