package main

const addrOamRam uint16 = 0xFE00

// Sprite represents to the properties of a sprite (or object in GB terms). The actual pixels are stored in the tile
// memory.
type Sprite struct {
	// X and Y coordinates. Note that x=0, y=0 represent off-screen sprites.
	x, y byte
	// The tile that actually holds the data for this sprite.
	tileNum byte
	// If true, the background is drawn on top of the sprite.
	bgPriority bool
	// Whether the sprite is mirrored, horizontally or vertically
	xFlip, yFlip bool
	// If true use palette 0, otherwise palette 1
	palette0 bool
}

func LoadSprite(mcu *Mcu, id byte) Sprite {
	if id >= 40 {
		panic("Invalid sprite ID: Only 40 sprites are supported")
	}
	baseAddr := addrOamRam + uint16(id)*4
	flags := mcu.Get(baseAddr + 3)
	return Sprite{
		x:          mcu.Get(baseAddr + 1),
		y:          mcu.Get(baseAddr),
		tileNum:    mcu.Get(baseAddr + 2),
		bgPriority: isBitSet(flags, 7),
		xFlip:      isBitSet(flags, 5),
		yFlip:      isBitSet(flags, 6),
		palette0:   !isBitSet(flags, 4),
	}
}

// OamDma is responsible for transferring sprite data from ROM to RAM.
type OamDma struct {
	ppuMem       *PpuMemory
	mcu          *Mcu
	copyAddr     byte
	transferByte uint16
}

func (o *OamDma) Tick() {
	if o.copyAddr > 0 {
		// If a copy is in progress, copy a byte every tick
		startAddr := uint16(o.copyAddr) << 8
		srcByte := o.mcu.Get(startAddr + o.transferByte)
		o.mcu.Set(addrOamRam+o.transferByte, srcByte)

		o.transferByte += 1
		if o.transferByte == 160 {
			// Transfer complete.
			o.transferByte = 0
			o.copyAddr = 0
		}
	} else {
		copyAddr := o.ppuMem.dmaAddr
		if copyAddr != 0 {
			// Start a new transfer.
			o.copyAddr = copyAddr
			o.ppuMem.dmaAddr = 0
		}
	}
}
