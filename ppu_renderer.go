package main

type ObjEntry struct {
	sprite *Sprite
	pixel  byte
}

type PpuRenderer struct {
	mem               *PpuMemory
	bgFifo            []byte
	objFifo           []ObjEntry
	bgFetcher         PpuFetcher
	objFetcher        PpuFetcher
	pixelSetter       PixelSetter
	sprites           []Sprite
	row, col          int
	step              int
	windowLineCounter int
	bgPalette         [4]byte
	spritePalette0    [4]byte
	spritePalette1    [4]byte

	mainMem *Mcu
}

func MakePpuRenderer(mcu *Mcu, mem *PpuMemory, pixelSetter PixelSetter) *PpuRenderer {
	return &PpuRenderer{
		mem:         mem,
		bgFetcher:   CreateFetcher(mcu, mem, Background),
		objFetcher:  CreateFetcher(mcu, mem, Object),
		pixelSetter: pixelSetter,
	}
}

func (p *PpuRenderer) Clear() {
	p.SetRow(0, []Sprite{})
	p.windowLineCounter = 0
}

func (p *PpuRenderer) SetRow(row byte, sprites []Sprite) {
	p.step = 0
	p.row = int(row)
	p.col = 0
	p.sprites = sprites

	p.bgPalette = makePalette(p.mem.bgPalette)
	p.spritePalette0 = makePalette(p.mem.objPalette0)
	p.spritePalette1 = makePalette(p.mem.objPalette1)

	p.objFetcher.SetRow(p.row)
	// Update addresses, in case the program moved the tiles somewhere else
	p.objFetcher.updateAddresses()
	p.objFifo = make([]ObjEntry, 0, 8)

	scRow := (p.row + int(p.mem.lcdScrollY)) % 256
	scCol := (p.col + int(p.mem.lcdScrollX)) % 256
	p.bgFetcher.SetRow(scRow)
	p.bgFetcher.col = scCol / 8 * 8
	// Re-set the fetcher type in case we switched it to Window previously
	p.bgFetcher.SetFetcherType(Background)
	p.bgFifo = make([]byte, 0, 16)
}

func (p *PpuRenderer) Tick() {
	if p.col == DisplayWidth {
		// TODO: Currently we draws the pixels immediately. Instead, use the correct timing sequencing.
		p.step++
		return
	}

	// Switch fetcher to window
	if p.mem.wndEnabled() && p.bgFetcher.fetcherType == Background &&
		byte(p.row) >= p.mem.windowY && byte(p.col) == p.mem.windowX-7 {
		p.bgFetcher.SetFetcherType(Window)
		p.bgFetcher.SetRow(p.windowLineCounter)
		p.bgFetcher.col = 0
		p.bgFifo = p.bgFifo[:0]
		p.windowLineCounter++
	}

	if len(p.bgFifo) < 8 {
		// Not enough pixels to draw, fetch a new tile.
		isFirstTile := (int(p.mem.lcdScrollX) / 8 * 8) == p.bgFetcher.col
		pixels := p.bgFetcher.GetNextPixels()
		if isFirstTile && p.mem.lcdScrollX%8 != 0 {
			// We are scrolled in-between a tile, so insert into the renderer the remaining pixels of that tile and
			// fetch the next tile.
			offset := p.mem.lcdScrollX % 8
			for i := offset; i < 8; i++ {
				p.bgFifo = append(p.bgFifo, pixels[i])
			}
			pixels = p.bgFetcher.GetNextPixels()
		}
		p.bgFifo = append(p.bgFifo, pixels[0], pixels[1], pixels[2], pixels[3], pixels[4], pixels[5], pixels[6], pixels[7])
	}

	if len(p.bgFifo) >= 8 {
		// We have pixels ready to draw, check if there's an overlapping sprite at this column.
		if len(p.sprites) > 0 && byte(p.col)+8 >= p.sprites[0].x {
			p.loadNextSprite()
		}

		// Output a pixel
		pixelColor := p.calcPixelColor()
		p.pixelSetter.SetPixel(p.row, p.col, pixelColor)
		p.col++

		// Advance the FIFOs
		if len(p.bgFifo) > 0 {
			p.bgFifo = p.bgFifo[1:]
		}
		if len(p.objFifo) > 0 {
			p.objFifo = p.objFifo[1:]
		}
	}
	p.step++
}

func (p *PpuRenderer) IsDone() bool {
	// TODO: 172 is the shortest duration of the renderer. Instead of this, we should use appropriate timing.
	return p.step == 172
}

func (p *PpuRenderer) loadNextSprite() {
	sprite := p.sprites[0]
	p.objFetcher.SetSprite(&sprite)
	spritePixels := p.objFetcher.GetNextPixels()

	// The pixels that overlap with the current sprite need to be mixed.
	// The previous sprite goes on top, unless it is transparent.
	i := 0
	for ; i < len(p.objFifo); i++ {
		if p.objFifo[i].pixel == 0 {
			p.objFifo[i] = ObjEntry{sprite: &sprite, pixel: spritePixels[i]}
		}
	}

	// All the remaining pixels can go in the fifo directly.
	for ; i < len(spritePixels); i++ {
		p.objFifo = append(p.objFifo, ObjEntry{sprite: &sprite, pixel: spritePixels[i]})
	}
	p.sprites = p.sprites[1:]
}

func (p *PpuRenderer) calcPixelColor() byte {
	bgColor := p.bgFifo[0]

	var objColor byte
	var bgPriority bool
	if len(p.objFifo) > 0 {
		objColor = p.objFifo[0].pixel
		bgPriority = p.objFifo[0].sprite.bgPriority
	} else {
		objColor = 0
		bgPriority = false
	}

	if objColor == 0 || (bgPriority && bgColor != 0) {
		// Background pixel
		if p.mem.bgWndEnabled() {
			return p.bgPalette[bgColor]
		}
		return p.bgPalette[0]
	}

	// Sprite pixel
	if p.objFifo[0].sprite.palette0 {
		return p.spritePalette0[objColor]
	}
	return p.spritePalette1[objColor]
}

func makePalette(v byte) [4]byte {
	return [4]byte{
		v & 0x3,
		(v & 0xc) >> 2,
		(v & 0x30) >> 4,
		(v & 0xc0) >> 6}
}
