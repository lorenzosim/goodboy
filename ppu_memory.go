package main

const (
	addrLcdControl  = 0xff40
	addrLcdStat     = 0xff41
	addrLcdScrollY  = 0xff42
	addrLcdScrollX  = 0xff43
	addrLcdLy       = 0xff44
	addrLcdLyc      = 0xff45
	addrDmaAddr     = 0xff46
	addrBgPalette   = 0xff47
	addrObjPalette0 = 0xff48
	addrObjPalette1 = 0xff49
	addrWindowY     = 0xff4a
	addrWindowX     = 0xff4b
)

type PpuMemory struct {
	lcdControl  byte
	lcdStat     byte
	lcdScrollY  byte
	lcdScrollX  byte
	lcdLy       byte
	lcdLyc      byte
	bgPalette   byte
	objPalette0 byte
	objPalette1 byte
	dmaAddr     byte
	windowX     byte
	windowY     byte
}

func (m *PpuMemory) Get(addr uint16) (byte, bool) {
	switch addr {
	case addrLcdControl:
		return m.lcdControl, true
	case addrLcdStat:
		return m.lcdStat, true
	case addrLcdScrollY:
		return m.lcdScrollY, true
	case addrLcdScrollX:
		return m.lcdScrollX, true
	case addrLcdLy:
		return m.lcdLy, true
	case addrLcdLyc:
		return m.lcdLyc, true
	case addrBgPalette:
		return m.bgPalette, true
	case addrObjPalette0:
		return m.objPalette0, true
	case addrObjPalette1:
		return m.objPalette1, true
	case addrDmaAddr:
		return m.dmaAddr, true
	case addrWindowX:
		return m.windowX, true
	case addrWindowY:
		return m.windowY, true
	default:
		return 0, false
	}
}

func (m *PpuMemory) Set(addr uint16, v byte) bool {
	switch addr {
	case addrLcdControl:
		m.lcdControl = v
		return true
	case addrLcdStat:
		// Bits 0,1,2 are not writable.
		m.lcdStat = (v & 0xf8) | (m.lcdStat & 0x7)
		return true
	case addrLcdScrollY:
		m.lcdScrollY = v
		return true
	case addrLcdScrollX:
		m.lcdScrollX = v
		return true
	case addrLcdLyc:
		m.lcdLyc = v
		return true
	case addrBgPalette:
		m.bgPalette = v
		return true
	case addrObjPalette0:
		m.objPalette0 = v
		return true
	case addrObjPalette1:
		m.objPalette1 = v
		return true
	case addrDmaAddr:
		m.dmaAddr = v
		return true
	case addrWindowX:
		m.windowX = v
		return true
	case addrWindowY:
		m.windowY = v
		return true
	default:
		return false
	}
}

func (m *PpuMemory) bgWndEnabled() bool {
	return isBitSet(m.lcdControl, 0)
}

func (m *PpuMemory) lcdObjEnabled() bool {
	return isBitSet(m.lcdControl, 1)
}

func (m *PpuMemory) lcdObjHeight() byte {
	if isBitSet(m.lcdControl, 2) {
		return 16
	}
	return 8
}

func (m *PpuMemory) lcdBgTileMap() bool {
	return isBitSet(m.lcdControl, 3)
}

func (m *PpuMemory) lcdBgWndTiles() bool {
	return isBitSet(m.lcdControl, 4)
}

func (m *PpuMemory) wndEnabled() bool {
	return isBitSet(m.lcdControl, 5)
}

func (m *PpuMemory) lcdWndTileMap() bool {
	return isBitSet(m.lcdControl, 6)
}

func (m *PpuMemory) lcdOn() bool {
	return isBitSet(m.lcdControl, 7)
}

func (m *PpuMemory) setLcdPpuMode(mode byte) {
	// Mode is bits 0,1 of lcd stat
	m.lcdStat = (m.lcdStat & 0xfc) | (mode & 0x3)
}

func (m *PpuMemory) setLcdLyEqualsLyc(equal bool) {
	m.lcdStat = setBitValue(m.lcdStat, 2, equal)
}

func (m *PpuMemory) statMode0Selected() bool {
	return isBitSet(m.lcdStat, 3)
}

func (m *PpuMemory) statMode1Selected() bool {
	return isBitSet(m.lcdStat, 4)
}

func (m *PpuMemory) statMode2Selected() bool {
	return isBitSet(m.lcdStat, 5)
}

func (m *PpuMemory) statLycSelected() bool {
	return isBitSet(m.lcdStat, 6)
}
