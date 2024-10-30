package main

import (
	"time"
)

const clockFreq = 1_048_576

// A Ticker is a system that advances every time Tick is called
type Ticker interface {
	Tick()
}

// PixelSetter allows to Set the color of a pixel at a given coordinate
type PixelSetter interface {
	SetPixel(r int, c int, color byte)
}

// PressedKeys encapsulate the status of all the keys used in GB
type PressedKeys struct {
	up, down, left, right, aBtn, bBtn, startBtn, selectBtn bool
}

// KeysListener listens for key changes and calls SetPressedKeys when a key is pressed or released
type KeysListener interface {
	SetPressedKeys(keys PressedKeys)
}

// Emulator represents the core of the emulator with all its subsystems
type Emulator struct {
	mcu       *Mcu
	cpu       Ticker
	ppu       *Ppu
	ppuMemory *PpuMemory
	dma       *OamDma
	timer     *Timer
	joypad    *JoyPad
	apu       *Apu
}

// MakeEmulator creates a new instance of Emulator
func MakeEmulator(bootRom []byte, rom []byte, debug bool, trace bool, pixelSetter PixelSetter) Emulator {
	interrupts := Interrupts{}
	apu := Apu{}
	joypad := JoyPad{interrupts: &interrupts}
	ppuMemory := PpuMemory{}
	timer := Timer{interrupts: &interrupts, apu: &apu}

	mcu := CreateMemory([]IoHandler{&ppuMemory, &joypad, &apu, &interrupts, &timer})
	cpu := CreateCpu(&mcu, &interrupts, trace)
	ppu := MakePpu(&mcu, &ppuMemory, &interrupts, pixelSetter)
	dma := OamDma{ppuMem: &ppuMemory, mcu: &mcu}

	if len(bootRom) > 0 {
		mcu.SetBootRom(bootRom)
	} else {
		// If no boot rom is given, Set a state similar to the DMG ROM.
		setDefaultState(cpu, &mcu)
	}
	mcu.SetRom(rom)

	var cpuRef Ticker = cpu
	if debug {
		cpuRef = &Debugger{cpu: cpu, paused: true}
	}

	e := Emulator{&mcu, cpuRef, &ppu, &ppuMemory, &dma, &timer, &joypad, &apu}
	return e
}

// Run runs the emulator. Blocking.
func (e *Emulator) Run() {
	targetCycleDuration := time.Second / clockFreq
	startTime := time.Now()
	var endTime time.Time
	for {
		e.Tick()
		endTime = time.Now()
		sleepTime := targetCycleDuration - endTime.Sub(startTime)
		startTime = endTime.Add(sleepTime) // Calculate startTime now, since we can't rely on time.Sleep to be exact
		time.Sleep(sleepTime)
	}
}

// A Tick of the emulator, should be called at 1Mhz for GMB original speed
func (e *Emulator) Tick() {
	e.cpu.Tick()
	e.dma.Tick()
	e.timer.Tick()
	e.apu.Tick()
	e.apu.Tick()

	// The PPU works at 4Mhz, so call it 4 times per tick
	e.ppu.Tick()
	e.ppu.Tick()
	e.ppu.Tick()
	e.ppu.Tick()
}

func setDefaultState(cpu *Cpu, mcu *Mcu) {
	// TODO: I believe the DMG rom sets other I/O registers as well.
	cpu.fc, cpu.fh, cpu.fz = true, true, true
	cpu.a, cpu.b, cpu.c, cpu.d, cpu.e, cpu.h, cpu.l = 0x01, 0x00, 0x13, 0x00, 0xD8, 0x01, 0x4D
	cpu.pc = 0x100
	mcu.Set(addrLcdControl, 0x91)
	mcu.Set(addrBgPalette, 0xfc)
}
