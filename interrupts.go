package main

const (
	addrInterruptEnable = 0xffff
	addrInterruptFlag   = 0xff0f
)

var interruptAddresses = [...]uint16{0x40, 0x48, 0x50, 0x58, 0x60}

type Interrupts struct {
	interruptFlag   byte
	interruptEnable byte
}

func (i *Interrupts) Get(addr uint16) (byte, bool) {
	switch addr {
	case addrInterruptEnable:
		return i.interruptEnable | 0xe0, true
	case addrInterruptFlag:
		return i.interruptFlag | 0xe0, true
	default:
		return 0, false
	}
}

func (i *Interrupts) Set(addr uint16, v byte) bool {
	v &= 0xf // Keep only the lower nibble
	switch addr {
	case addrInterruptEnable:
		i.interruptEnable = v
		return true
	case addrInterruptFlag:
		i.interruptFlag = v
		return true
	default:
		return false
	}
}

// ActiveInterruptIndex returns the index of the first interrupt that's both requested and enabled.
// If there's no such interrupt, returns -1.
func (i *Interrupts) ActiveInterruptIndex() int {
	if i.interruptFlag == 0 {
		// Small optimization for the common case in which no interrupts are requested.
		return -1
	}

	for c := 0; c < len(interruptAddresses); c++ {
		if isBitSet(i.interruptFlag, c) && isBitSet(i.interruptEnable, c) {
			return c
		}
	}
	return -1
}

func (i *Interrupts) RequestInterruptVBlank() {
	i.interruptFlag = setBit(i.interruptFlag, 0)
}

func (i *Interrupts) RequestInterruptStat() {
	i.interruptFlag = setBit(i.interruptFlag, 1)
}

func (i *Interrupts) RequestInterruptTimer() {
	i.interruptFlag = setBit(i.interruptFlag, 2)
}

func (i *Interrupts) RequestInterruptJoypad() {
	i.interruptFlag = setBit(i.interruptFlag, 4)
}
