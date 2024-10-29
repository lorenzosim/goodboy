package main

import (
	"fmt"
	"log"
)

type Cartridge struct {
	// What the program can access now
	rom0              []byte
	rom1              []byte
	mappedRam         []byte
	simpleBankingMode bool
	currentRomBank    int

	// The full data from the cartridge and ram, which is swapped into rom/mappedRam when the program
	// changes the control registers
	cartridge []byte
	fullRam   []byte

	// Cartridge info
	mapperType  int
	ramEnabled  bool
	numRomBanks int
	numRamBanks int
}

func (c *Cartridge) Load(cartridge []byte) {
	c.cartridge = cartridge
	c.setCartridgeInfo()

	c.simpleBankingMode = true
	c.currentRomBank = 1
	c.rom0 = c.cartridge[0:0x4000]
	c.rom1 = c.cartridge[0x4000:0x8000]
	c.fullRam = make([]byte, c.numRamBanks*8*1024)
}

func (c *Cartridge) Read(address uint16) byte {
	if address < 0x4000 {
		return c.rom0[address]
	}
	if address < 0x8000 {
		return c.rom1[address-0x4000]
	}
	if address >= 0xA000 && address < 0xc000 {
		if !c.ramEnabled {
			return 0xFF // Open value if RAM disabled or not there
		}
		return c.mappedRam[address-0xA000]
	}
	panic("Invalid address")
}

func (c *Cartridge) Write(address uint16, value byte) {
	if address >= 0xa000 && address < 0xc000 {
		if !c.ramEnabled {
			return // Ignore writes if RAM disabled or not there
		}
		c.mappedRam[address-0xa000] = value
		return
	}
	if address >= 0x8000 {
		panic("Invalid address")
	}

	switch c.mapperType {
	case 0:
		// Do nothing.
	case 1:
		if address < 0x2000 {
			c.ramEnabled = c.numRamBanks > 0 && (value&0xf) == 0xa
			if c.ramEnabled && c.mappedRam == nil {
				// If enabling the RAM before selecting a bank, default to the first bank.
				c.mappedRam = c.fullRam[:8192]
			}
			return
		}
		if address < 0x4000 {
			// Select ROM bank
			// TODO: more complicated than this
			romBank := min(c.numRomBanks-1, max(1, int(value&0x1f)))
			c.rom1 = c.cartridge[0x4000*romBank : 0x4000*(romBank+1)]
			c.currentRomBank = romBank
			return
		}
		if address < 0x6000 {
			// Select RAM bank
			ramBank := int(value & 0x3)
			if ramBank < c.numRamBanks && !c.simpleBankingMode {
				c.mappedRam = c.fullRam[8192*ramBank : 8192*(ramBank+1)]
			}
			return
		}
		// Select banking mode
		c.simpleBankingMode = value&0x1 == 0
		panic("Setting banking mode is not supported yet")
	default:
		panic("Unsupported mapper type")
	}
}

func (c *Cartridge) setCartridgeInfo() {
	var ramBanksByCode = []int{0, 1, 4, 16, 8}

	c.numRomBanks = 1 << (1 + c.cartridge[0x148])
	if c.numRomBanks > 32 {
		// TODO: add support for bigger cartridges.
		log.Fatal("ROM size over 512Kb not supported yet", c.numRamBanks)
	}
	cartridgeType := c.cartridge[0x147]
	ramCode := c.cartridge[0x149]

	switch cartridgeType {
	case 0x00:
		c.mapperType = 0
		c.numRamBanks = 0
	case 0x01:
		c.mapperType = 1
		c.numRamBanks = 0
	case 0x02, 0x03:
		c.mapperType = 1
		c.numRamBanks = ramBanksByCode[ramCode]
	default:
		panic(fmt.Sprintf("Unsupported cartridge type 0x%x", cartridgeType))
	}
}
