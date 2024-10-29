package main

import (
	"sort"
)

const (
	DisplayWidth         = 160
	DisplayHeight        = 144
	maxSpritesPerRow     = 10
	numDotsPerLine       = 456
	numScanLines         = 154
	oamScanDuration      = 80
	minRenderingDuration = 252
	maxRenderingDuration = 369
)

type PpuMode byte

const (
	HBlank    PpuMode = 0
	VBlank            = 1
	OamScan           = 2
	Rendering         = 3
)

// Ppu or Pixel Processing Unit of the system.
// Processes pixels into 3 modes:
// - OAM Scan: finds sprites for the current line
// - HBlank: pause between rendering of current line and next line
// - VBlank: finished rendering frame, pause before starting again
type Ppu struct {
	interrupts     *Interrupts
	mcu            *Mcu
	mem            *PpuMemory
	renderer       *PpuRenderer
	sprites        []Sprite
	mode           PpuMode
	currentLineDot int
}

func MakePpu(mainMem *Mcu, mem *PpuMemory, interrupts *Interrupts, pixelSetter PixelSetter) Ppu {
	return Ppu{mcu: mainMem, mem: mem, interrupts: interrupts, mode: 2, renderer: MakePpuRenderer(mainMem, mem, pixelSetter)}
}

func (ppu *Ppu) Tick() {
	switch ppu.mode {
	case HBlank:
		ppu.hBlank()
	case VBlank:
		ppu.vBlank()
	case OamScan:
		ppu.oamScan()
	case Rendering:
		ppu.rendering()
	default:
		panic("Invalid PPU mode")
	}
}

func (ppu *Ppu) oamScan() {
	ppu.currentLineDot += 1

	spritesEnabled := ppu.mem.lcdObjEnabled()
	if spritesEnabled && ppu.currentLineDot%2 == 1 {
		// Check if a sprite overlaps with the current line. We do this every other dots, so this is called 40 times,
		// exactly matching the number of sprites in the system.
		spriteHeight := ppu.mem.lcdObjHeight()
		spriteId := byte(ppu.currentLineDot / 2) // 0-40
		sprite := LoadSprite(ppu.mcu, spriteId)

		if len(ppu.sprites) < maxSpritesPerRow &&
			ppu.mem.lcdLy+16 >= sprite.y && ppu.mem.lcdLy+16 < (sprite.y+spriteHeight) {
			// +16 because when s.y=16 the sprite is fully on screen on the first row.
			ppu.sprites = append(ppu.sprites, sprite)
		}
	}

	if ppu.currentLineDot == oamScanDuration {
		if spritesEnabled {
			// Sort sprites by x coordinate.
			sort.SliceStable(ppu.sprites, func(i, j int) bool {
				return ppu.sprites[i].x < ppu.sprites[j].x
			})
		}
		ppu.renderer.SetRow(ppu.mem.lcdLy, ppu.sprites)
		ppu.sprites = ppu.sprites[:0]
		ppu.switchMode(Rendering)
	}
}

func (ppu *Ppu) rendering() {
	ppu.renderer.Tick()

	ppu.currentLineDot += 1
	if ppu.renderer.IsDone() {
		// Finished rendering, go to HBlank
		if ppu.currentLineDot < minRenderingDuration || ppu.currentLineDot > maxRenderingDuration {
			panic("Unexpected rendering duration")
		}
		ppu.switchMode(HBlank)
	}
}

func (ppu *Ppu) hBlank() {
	ppu.currentLineDot += 1

	if ppu.currentLineDot == numDotsPerLine {
		ppu.nextLine()
	}
}

func (ppu *Ppu) vBlank() {
	ppu.currentLineDot += 1

	if ppu.currentLineDot == numDotsPerLine {
		ppu.nextLine()
	}
}

func (ppu *Ppu) switchMode(mode PpuMode) {
	ppu.mode = mode
	ppu.mem.setLcdPpuMode(byte(mode))
	switch mode {
	case HBlank:
		if ppu.mem.statMode0Selected() {
			ppu.interrupts.RequestInterruptStat()
		}
	case VBlank:
		ppu.interrupts.RequestInterruptVBlank()
		if ppu.mem.statMode1Selected() {
			ppu.interrupts.RequestInterruptStat()
		}
	case OamScan:
		if ppu.mem.statMode2Selected() {
			ppu.interrupts.RequestInterruptStat()
		}
	}
}

func (ppu *Ppu) nextLine() {
	ppu.mem.lcdLy = (ppu.mem.lcdLy + 1) % numScanLines
	ppu.currentLineDot = 0

	if ppu.mem.lcdLy < DisplayHeight {
		// Done with the line, onto oam scan for the next line
		ppu.switchMode(OamScan)
	} else if ppu.mem.lcdLy == DisplayHeight {
		// Done with all the on-screen pixel, onto v-blank
		ppu.renderer.Clear()
		ppu.switchMode(VBlank)
	}

	// Check if current line matches requested line
	ly := ppu.mem.lcdLy
	lyc := ppu.mem.lcdLyc
	ppu.mem.setLcdLyEqualsLyc(lyc == ly)
	if lyc == ly && ppu.mem.statLycSelected() {
		ppu.interrupts.RequestInterruptStat()
	}
}
