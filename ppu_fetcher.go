package main

import (
	"log"
)

var tileMaps = []uint16{0x9800, 0x9C00}
var tileBlocks = []uint16{0x8000, 0x8800, 0x9000}

type FetcherType int

const (
	Background FetcherType = iota
	Window
	Object
)

// PpuFetcher is responsible for fetching tile data from memory and returning pixels.
type PpuFetcher struct {
	mcu         *Mcu
	mem         *PpuMemory
	fetcherType FetcherType
	sprite      *Sprite
	row         int
	col         int
	// Where to look for in memory for tile map and tile data
	tileDataStart []uint16
	tileMapStart  uint16
	// Row position within a tile
	tileRow int
	tileId  byte
}

func CreateFetcher(mcu *Mcu, mem *PpuMemory, fetcherType FetcherType) PpuFetcher {
	p := PpuFetcher{mcu: mcu, mem: mem, fetcherType: fetcherType}
	p.updateAddresses()
	return p
}

func (p *PpuFetcher) SetFetcherType(fetcherType FetcherType) {
	p.fetcherType = fetcherType
	p.updateAddresses()
}

func (p *PpuFetcher) SetRow(row int) {
	p.row = row
	p.col = 0
	p.tileRow = p.row % 8 // Only for background. For sprites, overwritten in SetSprite.
}

// GetNextPixels returns the next 8 pixels of the current row.
func (p *PpuFetcher) GetNextPixels() [8]byte {
	if p.sprite == nil {
		p.tileId = p.getTileId()
	}
	tileData0, tileData1 := p.getTileData()

	var pixels [8]byte
	var xFlipped = false
	if p.sprite != nil && p.sprite.xFlip {
		xFlipped = true
	}
	var columnMask byte = 128

	for col := 0; col < 8; col++ {
		var leftBit byte = 0
		var rightBit byte = 0
		if (tileData1 & columnMask) != 0 {
			leftBit = 1
		}
		if (tileData0 & columnMask) != 0 {
			rightBit = 1
		}
		colorId := (leftBit << 1) + rightBit
		if colorId < 0 || colorId > 3 {
			panic("Unexpected color")
		}
		if xFlipped {
			pixels[7-col] = colorId
		} else {
			pixels[col] = colorId
		}
		columnMask >>= 1
	}

	p.col = (p.col + 8) % 256
	return pixels
}

func (p *PpuFetcher) SetSprite(sprite *Sprite) {
	if p.fetcherType != Object {
		log.Fatalf("Sprites can only be set on Object fetchers")
	}

	p.sprite = sprite
	if p.mem.lcdObjHeight() == 8 {
		p.tileId = sprite.tileNum
	} else {
		// With 8x16 tiles, the tileNum refers to the top tile and the bottom one is immediately after (reverse if flipped).
		// Additionally, the bit 0 in the tile number is ignored.
		topTile := sprite.tileNum & 0xfe
		bottomTile := topTile + 1
		isBottomRow := p.row+16 > int(p.sprite.y+8)
		if isBottomRow != sprite.yFlip {
			p.tileId = bottomTile
		} else {
			p.tileId = topTile
		}
	}

	p.tileRow = (p.row + 16 - int(sprite.y)) % 8 // % 8 to account for 8x16 tiles
	if p.sprite.yFlip && p.mem.lcdObjHeight() == 8 {
		p.tileRow = 7 - p.tileRow
	}
}

// getTileId returns the id of the tile to draw the current row/column.
func (p *PpuFetcher) getTileId() byte {
	if p.fetcherType == Object {
		panic("getTileId can not be called on Object fetcher")
	}

	tileIndex := ((p.row / 8) * 32) + (p.col / 8) // Tile map is 32x32 and each tile is 8x8
	tileId := p.mcu.Get(p.tileMapStart + uint16(tileIndex))
	return tileId
}

// getTileData returns the 2 bytes of tile data needed to draw a row of the tile
func (p *PpuFetcher) getTileData() (byte, byte) {
	index := uint16(p.tileRow * 2) // Each tile row is 2 bytes long
	if index < 0 || index > 14 {
		log.Fatalf("Invalid tile index: %d", index)
	}

	tileBlockStart := p.tileDataStart[p.tileId/128]
	tileBlockOffset := 16 * uint16(p.tileId%128)
	tileStart := tileBlockStart + tileBlockOffset
	return p.mcu.Get(tileStart + index), p.mcu.Get(tileStart + index + 1)
}

func (p *PpuFetcher) updateAddresses() {
	// Update tile data address
	if p.fetcherType == Object || p.mem.lcdBgWndTiles() {
		p.tileDataStart = []uint16{tileBlocks[0], tileBlocks[1]}
	} else {
		p.tileDataStart = []uint16{tileBlocks[2], tileBlocks[1]}
	}

	// ...and tile map address
	p.tileMapStart = tileMaps[0]
	if (p.fetcherType == Background && p.mem.lcdBgTileMap()) ||
		(p.fetcherType != Background && p.mem.lcdWndTileMap()) {
		p.tileMapStart = tileMaps[1]
	}
}
