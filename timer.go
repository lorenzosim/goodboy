package main

var clockCycles = []uint16{256, 4, 16, 64}

const (
	addrDiv  = 0xff04
	addrTima = 0xff05
	addrTma  = 0xff06
	addrTac  = 0xff07
)

type Timer struct {
	apu          *Apu
	interrupts   *Interrupts
	ticks        uint16
	div          byte
	timerEnabled bool
	clockSelect  byte
	tma          byte
	tima         byte
}

func (t *Timer) Get(addr uint16) (byte, bool) {
	switch addr {
	case addrDiv:
		return t.div, true
	case addrTima:
		return t.tima, true
	case addrTma:
		return t.tma, true
	case addrTac:
		v := t.clockSelect
		if t.timerEnabled {
			v = setBit(v, 2)
		}
		return v | 0xf8, true
	default:
		return 0, false
	}
}

func (t *Timer) Set(addr uint16, v byte) bool {
	switch addr {
	case addrDiv:
		t.div = 0
		t.ticks = 0
		return true
	case addrTima:
		t.tima = v
		return true
	case addrTma:
		t.tma = v
		return true
	case addrTac:
		t.timerEnabled = isBitSet(v, 2)
		t.clockSelect = v & 0x3
		return true
	default:
		return false
	}
}

func (t *Timer) Tick() {
	t.ticks++

	if t.ticks%64 == 0 {
		// Increment DIV
		oldDiv := t.div
		t.div = oldDiv + 1
		t.apu.OnTimer(oldDiv, t.div)
	}

	if t.timerEnabled {
		selectedCycle := clockCycles[t.clockSelect]

		if t.ticks%selectedCycle == 0 {
			// Increment TIMA. Trigger interrupt on overflow and reset to TMA
			if t.tima == 0xff {
				t.tima = t.tma
				t.interrupts.RequestInterruptTimer()
			} else {
				t.tima++
			}
		}
	}
}
