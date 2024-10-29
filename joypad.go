package main

const addrJoypad = 0xff00

type JoyPad struct {
	interrupts *Interrupts
	// Whether buttons and/or dpads are select (only bits 4 and 5 can be 0)
	selection     byte
	dpadNibble    byte
	buttonsNibble byte
}

func (j *JoyPad) SetPressedKeys(pressedKeys PressedKeys) {
	oldValue, _ := j.Get(addrJoypad)

	// Construct nibbles for dpad and buttons
	j.dpadNibble =
		buildNibble([4]bool{pressedKeys.right, pressedKeys.left, pressedKeys.up, pressedKeys.down})
	j.buttonsNibble =
		buildNibble([4]bool{pressedKeys.aBtn, pressedKeys.bBtn, pressedKeys.selectBtn, pressedKeys.startBtn})

	// Trigger interrupt if a button state changed high->low
	newValue, _ := j.Get(addrJoypad)
	if (oldValue & ^newValue) != 0 {
		j.interrupts.RequestInterruptJoypad()
	}
}

func (j *JoyPad) Get(addr uint16) (byte, bool) {
	if addr != addrJoypad {
		return 0, false
	}

	var result = j.selection
	if !isBitSet(j.selection, 4) {
		result &= j.dpadNibble
	}
	if !isBitSet(j.selection, 5) {
		result &= j.buttonsNibble
	}
	return result, true
}

func (j *JoyPad) Set(addr uint16, v byte) bool {
	if addr != addrJoypad {
		return false
	}
	j.selection = v | 0xcf // Set unused bits to 1.
	return true
}

func buildNibble(pressedKeys [4]bool) byte {
	var nibble byte = 0xff // All bits initially Set to 1 (no button pressed).
	for i, pressed := range pressedKeys {
		if pressed {
			nibble = clearBit(nibble, i)
		}
	}
	return nibble
}
