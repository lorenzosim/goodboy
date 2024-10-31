package main

import (
	"fmt"
	"log"
	"time"
)

type Cartridge struct {
	// What the program can access now
	rom0           []byte
	rom1           []byte
	mappedRam      []byte
	currentRomBank int

	// The full data from the cartridge and ram, which is swapped into rom/mappedRam when the program
	// changes the control registers
	cartridge []byte
	fullRam   []byte

	// Cartridge info
	mapperType  int
	ramEnabled  bool
	numRomBanks int
	numRamBanks int

	// For MBC1 simple banking mode selected?
	mbc1SimpleBankingMode bool
	// For MBC3, RTC selected?
	mbc3ReadRtc bool
	// For MBC3, which RTC register to read/write
	mbc3RtcRegister int
}

func (c *Cartridge) Load(cartridge []byte) {
	c.cartridge = cartridge
	c.setCartridgeInfo()

	c.mbc1SimpleBankingMode = true
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
			if c.mapperType == 3 {
				return c.readMbc3Rtc()
			} else {
				return 0xFF // Open value if RAM disabled or not there
			}
		}
		return c.mappedRam[address-0xA000]
	}
	panic("Invalid address")
}

func (c *Cartridge) Write(address uint16, value byte) {
	if address >= 0xc000 {
		log.Fatal("Invalid address: ", address)
	}

	switch c.mapperType {
	case 0:
		// Do nothing.
	case 1:
		c.writeMbc1(address, value)
	case 3:
		c.writeMbc3(address, value)
	default:
		panic("Unsupported mapper type")
	}
}

func (c *Cartridge) setCartridgeInfo() {
	var ramBanksByCode = []int{0, 0, 1, 4, 16, 8}

	c.numRomBanks = 1 << (1 + c.cartridge[0x148])
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
	case 0xf, 0x11:
		c.mapperType = 3
		c.numRamBanks = 0
	case 0x10, 0x12, 0x13:
		c.mapperType = 3
		c.numRamBanks = ramBanksByCode[ramCode]
	default:
		panic(fmt.Sprintf("Unsupported cartridge type 0x%x", cartridgeType))
	}
}

func (c *Cartridge) writeMbc1(address uint16, value byte) {
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
		romBank := min(c.numRomBanks-1, max(1, int(value&0x1f)))
		c.rom1 = c.cartridge[0x4000*romBank : 0x4000*(romBank+1)]
		c.currentRomBank = romBank
		return
	}
	if address < 0x6000 {
		// Select RAM bank
		ramBank := int(value & 0x3)
		if ramBank < c.numRamBanks && !c.mbc1SimpleBankingMode {
			c.mappedRam = c.fullRam[8192*ramBank : 8192*(ramBank+1)]
		}
		return
	}
	if address >= 0xa000 && address < 0xc000 {
		if !c.ramEnabled {
			return // Ignore writes if RAM disabled or not there
		}
		c.mappedRam[address-0xa000] = value
		return
	}
	// Select banking mode
	c.mbc1SimpleBankingMode = value&0x1 == 0
}

func (c *Cartridge) readMbc3Rtc() byte {
	switch c.mbc3RtcRegister {
	case 0x08:
		return byte(time.Now().Second())
	case 0x09:
		return byte(time.Now().Minute())
	case 0x0a:
		return byte(time.Now().Hour())
	case 0x0b:
	case 0x0c:
		// TODO: Support day counter.
		return 0
	default:
		panic("Unexpected RTC register")
	}
	panic("Should not happen")
}

func (c *Cartridge) writeMbc3(address uint16, value byte) {
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
		romBank := min(c.numRomBanks-1, max(1, int(value)))
		c.rom1 = c.cartridge[0x4000*romBank : 0x4000*(romBank+1)]
		c.currentRomBank = romBank
		return
	}
	if address < 0x6000 {
		// Select RAM bank
		if value < 0x8 {
			ramBank := int(value & 0x3)
			if ramBank < c.numRamBanks {
				c.mappedRam = c.fullRam[8192*ramBank : 8192*(ramBank+1)]
			}
		} else if value < 0xd {
			c.mbc3RtcRegister = int(value)
		}
		return
	}
	if address < 0x8000 {
		// TODO: Support RTC latching.
		return
	}
	if address < 0xc000 {
		if c.ramEnabled {
			c.mappedRam[address-0xa000] = value
		} else {
			// TODO: Support writing to RTC registers.
		}
		return
	}
}
