package main

const (
	addrBootRomEnd  = 0x100
	addrRomEnd      = 0x8000
	addrVideoRamEnd = 0xa000
	addrCartRamEnd  = 0xc000
	addrWorkRamEnd  = 0xe000
	addrEchoRamEnd  = 0xfe00
	addrOamEnd      = 0xfea0
	addrUnusableEnd = 0xff00
	addrIoEnd       = 0xff80
	addrUseBootRom  = 0xff50
	openValue       = 0xff
)

// Mcu (Memory Control Unit) of the system.
type Mcu struct {
	bootRom        []byte // Boot ROM internal to the device, 256 bytes
	bootRomEnabled bool   // If true, access to addresses 0-0x100 goes to the cartridge

	vram [8 * 1024]byte // Video RAM
	wram [8 * 1024]byte // Work RAM
	oam  [160]byte      // Object Attribute Memory, where sprite attributes are stored
	hram [127]byte      // High ram

	cartridge  *Cartridge
	ioHandlers []IoHandler
}

// IoHandler handles reads and writes for an I/O component (e.g. joypad, apu)
type IoHandler interface {
	// Get returns the value of the I/O register at the given address and 'true'. If the address does not map to a
	// register, returns an arbitrary value and 'false'
	Get(addr uint16) (byte, bool)

	// Set writes the given value to the I/O register at the given address. Returns 'true' if the value was written,
	// 'false' if this device does not handle the given address)
	Set(addr uint16, value byte) bool
}

func CreateMemory(ioHandlers []IoHandler) Mcu {
	mcu := Mcu{ioHandlers: ioHandlers}
	mcu.bootRomEnabled = false // Default to no boot rom
	return mcu
}

func (mcu *Mcu) SetBootRom(rom []byte) {
	if len(rom) != addrBootRomEnd {
		panic("Rom is wrong size")
	}
	mcu.bootRom = rom
	mcu.bootRomEnabled = true
}

func (mcu *Mcu) SetRom(rom []byte) {
	mcu.cartridge = &Cartridge{}
	mcu.cartridge.Load(rom)
}

func (mcu *Mcu) GetWord(address uint16) uint16 {
	return merge(mcu.Get(address+1), mcu.Get(address))
}

func (mcu *Mcu) Get(address uint16) byte {
	switch {
	case address < addrBootRomEnd && mcu.bootRomEnabled:
		return mcu.bootRom[address]
	case address < addrRomEnd:
		return mcu.cartridge.Read(address)
	case address < addrVideoRamEnd:
		return mcu.vram[address-addrRomEnd]
	case address < addrCartRamEnd:
		return mcu.cartridge.Read(address)
	case address < addrWorkRamEnd:
		return mcu.wram[address-addrCartRamEnd]
	case address < addrEchoRamEnd:
		return mcu.wram[address-addrWorkRamEnd]
	case address < addrOamEnd:
		return mcu.oam[address-addrEchoRamEnd]
	case address < addrUnusableEnd:
		return openValue // Not usable, return open value
	case address < addrIoEnd || address == addrInterruptEnable:
		for _, handler := range mcu.ioHandlers {
			if value, handled := handler.Get(address); handled {
				return value
			}
		}
		return openValue
	case address < addrInterruptEnable:
		return mcu.hram[address-addrIoEnd]
	default:
		panic("Unexpected address")
	}
}

func (mcu *Mcu) Set(address uint16, value byte) {
	switch {
	case address < addrBootRomEnd && mcu.bootRomEnabled:
		// Do nothing, can't write to boot rom
	case address < addrRomEnd:
		mcu.cartridge.Write(address, value)
	case address < addrVideoRamEnd:
		mcu.vram[address-addrRomEnd] = value
	case address < addrCartRamEnd:
		mcu.cartridge.Write(address, value)
	case address < addrWorkRamEnd:
		mcu.wram[address-addrCartRamEnd] = value
	case address < addrEchoRamEnd:
		mcu.wram[address-addrWorkRamEnd] = value
	case address < addrOamEnd:
		mcu.oam[address-addrEchoRamEnd] = value
	case address < addrUnusableEnd:
		// Do nothing, unusable area
	case address < addrIoEnd || address == addrInterruptEnable:
		if address == addrUseBootRom && value != 0 {
			mcu.bootRomEnabled = false
		} else {
			for _, handler := range mcu.ioHandlers {
				if handler.Set(address, value) {
					return
				}
			}
		}
	case address < addrInterruptEnable:
		mcu.hram[address-addrIoEnd] = value
	default:
		panic("Unexpected address")
	}
}

func (mcu *Mcu) SetWord(address uint16, value uint16) {
	hi, lo := split(value)
	mcu.Set(address+1, hi)
	mcu.Set(address, lo)
}
