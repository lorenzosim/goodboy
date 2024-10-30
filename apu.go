package main

import "sync"

const (
	addrNr10        = 0xff10
	addrNr11        = 0xff11
	addrNr12        = 0xff12
	addrNr13        = 0xff13
	addrNr14        = 0xff14
	addrNr21        = 0xff16
	addrNr22        = 0xff17
	addrNr23        = 0xff18
	addrNr24        = 0xff19
	addrNr30        = 0xff1a
	addrNr31        = 0xff1b
	addrNr32        = 0xff1c
	addrNr33        = 0xff1d
	addrNr34        = 0xff1e
	addrNr41        = 0xff20
	addrNr42        = 0xff21
	addrNr43        = 0xff22
	addrNr44        = 0xff23
	addrNr50        = 0xff24
	addrNr51        = 0xff25
	addrNr52        = 0xff26
	addrWavePattern = 0xff30
	maxPeriod       = 0x7ff
)

// Apu or Audio Processing Unit.
// This component is quite complicated with lots of quirks and edge cases, see:
// https://gbdev.gg8.se/wiki/articles/Gameboy_sound_hardware
type Apu struct {
	divApu byte

	// Master volume & VIN panning
	masterVolume byte
	// Sound panning
	panning byte

	// Is the APU enabled?
	audioEnabled bool
	// Whether each channel is on. A channel is on after it is triggered.
	channelsOn [4]bool
	// Whether the DAC of each channel is enabled
	dacEnabled [4]bool
	// Timer that auto-turns off the channel when it reaches 0
	lengthTimer [4]uint16
	// Whether the above timer is enabled
	timerEnabled [4]bool
	// Volume of the channel.
	volume [4]byte
	// Volume envelope, which increases or decreases the volume over time. Not used for channel 3.
	volumeEnvelope [4]byte
	// Timer decrements every cycle. When it reaches 0, a new sample is produced.
	frequencyTimer [4]uint16
	// Current period of the channel (not used for channel 4), opposite of the frequency.
	period [3]uint16
	// Current position in the wave we are playing (not used for channel 4)
	wavePosition [3]int

	// Channel 1
	ch1DutyCycle         byte
	ch1Sweep             byte
	ch1SweepShadowPeriod uint16
	ch1SweepEnabled      bool
	ch1SweepTimer        byte

	// Channel 2
	ch2DutyCycle byte

	// Channel 3
	ch3Wave [16]byte

	// Channel 4
	ch4Randomness byte
	ch4Lfsr       uint16

	// Buffer of audio samples we have generated. A new sample is produced every ~48kHz.
	samples   []AudioSample
	samplesMu sync.Mutex

	// Incremented every APU tick (2Mhz)
	tick int
}

type AudioSample struct {
	left, right float32
}

func (a *Apu) OnTimer(divOld byte, divNew byte) {
	if !(isBitSet(divOld, 4) && !isBitSet(divNew, 4)) {
		return
	}
	a.divApu++

	if a.divApu%4 == 0 {
		a.handleSweep()
	}
	if a.divApu%8 == 0 {
		a.handleEnvelope()
	}
	if a.divApu%2 == 0 {
		a.handleLengthTimer()
	}
}

// handleSweep handles frequency sweeping (increase/decrease of frequency over time) for the first channel
func (a *Apu) handleSweep() {
	if a.ch1SweepTimer > 0 {
		a.ch1SweepTimer--
	}
	if a.ch1SweepTimer > 0 {
		return
	}

	// Reload the timer.
	a.updateSweepTimer()

	pace := (a.ch1Sweep & 0x70) >> 4
	if a.ch1SweepEnabled && pace > 0 {
		shift := a.ch1Sweep & 0x7
		newPeriod := a.calcNewSweepPeriod()
		if newPeriod > maxPeriod {
			a.channelsOn[0] = false
		} else if shift > 0 {
			a.ch1SweepShadowPeriod = newPeriod
			a.period[0] = newPeriod
			// For some reason, per specifications the check is performed again.
			if a.calcNewSweepPeriod() > maxPeriod {
				a.channelsOn[0] = false
			}
		}
	}
}

func (a *Apu) calcNewSweepPeriod() uint16 {
	shift := a.ch1Sweep & 0x7
	delta := a.ch1SweepShadowPeriod >> shift
	var newPeriod = a.ch1SweepShadowPeriod
	if !isBitSet(a.ch1Sweep, 3) {
		newPeriod += delta
	} else {
		newPeriod -= delta
	}
	return newPeriod
}

// updateSweepTimer updates the sweep timer. Quirk: if the pace is 0, the timer gets a value of 8.
func (a *Apu) updateSweepTimer() {
	pace := (a.ch1Sweep & 0x70) >> 4
	if pace > 0 {
		a.ch1SweepTimer = pace
	} else {
		a.ch1SweepTimer = 8
	}
}

// handleEnvelope handles volume sweeping (increase/decrease of volume over time) for channels 1,2,4.
func (a *Apu) handleEnvelope() {
	for _, c := range [3]int{0, 1, 3} {
		pace := a.volumeEnvelope[c] & 0x7
		if pace == 0 {
			// Envelope disabled
			continue
		}
		if a.divApu%(8*pace) == 0 {
			increaseVolume := isBitSet(a.volumeEnvelope[c], 3)
			if increaseVolume && a.volume[c] != 0xF {
				a.volume[c]++
			} else if !increaseVolume && a.volume[c] != 0 {
				a.volume[c]--
			}
		}
	}
}

// handleLengthTimer handles the length timer, which auto-disables a channel after a certain amount of time.
func (a *Apu) handleLengthTimer() {
	for c := 0; c < 4; c++ {
		if a.timerEnabled[c] {
			a.lengthTimer[c]--
			if a.lengthTimer[c] == 0 {
				// Turn off the channel
				a.channelsOn[c] = false
			}
		}
	}
}

func (a *Apu) Get(addr uint16) (byte, bool) {
	if addr >= addrWavePattern && addr < addrWavePattern+16 {
		return a.ch3Wave[addr-addrWavePattern], true
	}

	// Note: when getting a value, only some bits can be read. The other ones are set to 1, which is accomplished with a
	// bitwise AND.
	switch addr {
	case addrNr10:
		return a.ch1Sweep | 0x80, true
	case addrNr11:
		return (a.ch1DutyCycle << 6) | 0x3f, true
	case addrNr12:
		return a.volumeEnvelope[0], true
	case addrNr14:
		v := byte(0xff)
		if !a.timerEnabled[0] {
			v = clearBit(v, 6)
		}
		return v, true
	case addrNr21:
		return (a.ch2DutyCycle << 6) | 0x3f, true
	case addrNr22:
		return a.volumeEnvelope[1], true
	case addrNr24:
		v := byte(0xff)
		if !a.timerEnabled[1] {
			v = clearBit(v, 6)
		}
		return v, true
	case addrNr30:
		var v byte = 0xff
		if !a.dacEnabled[2] {
			v = clearBit(v, 7)
		}
		return v, true
	case addrNr32:
		return (a.volume[2] << 5) | 0x9f, true
	case addrNr34:
		var v byte = 0xff
		if !a.timerEnabled[2] {
			v = clearBit(v, 6)
		}
		return v, true
	case addrNr42:
		return a.volumeEnvelope[3], true
	case addrNr43:
		return a.ch4Randomness, true
	case addrNr44:
		v := byte(0xff)
		if !a.timerEnabled[3] {
			v = clearBit(v, 6)
		}
		return v, true
	case addrNr50:
		return a.masterVolume, true
	case addrNr51:
		return a.panning, true
	case addrNr52:
		v := byte(0)
		v = setBitValue(v, 0, a.channelsOn[0])
		v = setBitValue(v, 1, a.channelsOn[1])
		v = setBitValue(v, 2, a.channelsOn[2])
		v = setBitValue(v, 3, a.channelsOn[3])
		v = setBitValue(v, 7, a.audioEnabled)
		return v | 0x70, true
	}
	return 0, false
}

func (a *Apu) Set(addr uint16, v byte) bool {
	if addr >= addrWavePattern && addr < addrWavePattern+16 {
		a.ch3Wave[addr-addrWavePattern] = v
		return true
	}

	switch addr {
	case addrNr10:
		if a.audioEnabled {
			if a.channelsOn[0] && a.ch1SweepEnabled && !isBitSet(v, 3) && isBitSet(a.ch1Sweep, 3) {
				// Changing sweep direction pos->neg turns the channel off.
				a.channelsOn[0] = false
			}
			a.ch1Sweep = v
		}
		return true
	case addrNr11:
		if a.audioEnabled {
			a.lengthTimer[0] = 64 - uint16(v&0x3f)
			a.ch1DutyCycle = v >> 6
		}
		return true
	case addrNr12:
		if a.audioEnabled {
			a.volumeEnvelope[0] = v
			if v&0xf8 == 0 {
				a.channelsOn[0] = false
				a.dacEnabled[0] = false
			} else {
				a.dacEnabled[0] = true
			}
		}
		return true
	case addrNr13:
		if a.audioEnabled {
			a.period[0] = (a.period[0] & 0x700) | uint16(v)
		}
		return true
	case addrNr14:
		if a.audioEnabled {
			// Set period upper 3 bits
			a.period[0] = (uint16(v&0x7) << 8) | (a.period[0] & 0xff)
			a.ch1SweepShadowPeriod = a.period[0]
			// Set timer enable
			a.timerEnabled[0] = isBitSet(v, 6)
			if a.timerEnabled[0] && a.lengthTimer[0] == 0 {
				a.lengthTimer[0] = 64 // Timer 0 means max length
			}
			// Trigger channel
			trigger := isBitSet(v, 7)
			if a.dacEnabled[0] && trigger {
				a.volume[0] = (a.volumeEnvelope[0] & 0xf0) >> 4
				a.wavePosition[0] = 0
				a.frequencyTimer[0] = maxPeriod + 1 - a.period[0]

				hasSweepShift := (a.ch1Sweep & 0x7) != 0
				hasSweepPace := (a.ch1Sweep & 0x70) != 0
				a.updateSweepTimer()
				a.ch1SweepEnabled = hasSweepShift || hasSweepPace

				a.channelsOn[0] = true
				if hasSweepShift && a.calcNewSweepPeriod() > maxPeriod {
					a.channelsOn[0] = false
				}
			}
		}
		return true
	case addrNr21:
		if a.audioEnabled {
			a.lengthTimer[1] = 64 - uint16(v&0x3f)
			a.ch2DutyCycle = v >> 6
		}
		return true
	case addrNr22:
		if a.audioEnabled {
			a.volumeEnvelope[1] = v
			if v&0xf8 == 0 {
				a.channelsOn[1] = false
				a.dacEnabled[1] = false
			} else {
				a.dacEnabled[1] = true
			}
		}
		return true
	case addrNr23:
		if a.audioEnabled {
			a.period[1] = (a.period[1] & 0x700) | uint16(v)
		}
		return true
	case addrNr24:
		if a.audioEnabled {
			// Set period upper 3 bits
			a.period[1] = (uint16(v&0x7) << 8) | (a.period[1] & 0xff)
			// Set timer enable
			a.timerEnabled[1] = isBitSet(v, 6)
			if a.timerEnabled[1] && a.lengthTimer[1] == 0 {
				a.lengthTimer[1] = 64 // Timer 0 means max length
			}
			// Trigger channel
			trigger := isBitSet(v, 7)
			if a.dacEnabled[1] && trigger {
				a.volume[1] = (a.volumeEnvelope[1] & 0xf0) >> 4
				a.wavePosition[1] = 0
				a.frequencyTimer[1] = maxPeriod + 1 - a.period[1]
				a.channelsOn[1] = true
			}
		}
		return true
	case addrNr30:
		if a.audioEnabled {
			// Set channel 3 DAC
			a.dacEnabled[2] = isBitSet(v, 7)
			if !a.dacEnabled[2] {
				a.channelsOn[2] = false
			}
		}
		return true
	case addrNr31:
		if a.audioEnabled {
			a.lengthTimer[2] = 256 - uint16(v)
		}
		return true
	case addrNr32:
		if a.audioEnabled {
			a.volume[2] = v >> 5
		}
		return true
	case addrNr33:
		if a.audioEnabled {
			a.period[2] = (a.period[2] & 0x700) | uint16(v)
		}
		return true
	case addrNr34:
		if a.audioEnabled {
			// Set period upper 3 bits
			a.period[2] = (uint16(v&0x7) << 8) | (a.period[2] & 0xff)
			// Set timer enable
			a.timerEnabled[2] = isBitSet(v, 6)
			if a.timerEnabled[2] && a.lengthTimer[2] == 0 {
				a.lengthTimer[2] = 256 // Timer 0 means max length
			}
			// Trigger channel 3
			trigger := isBitSet(v, 7)
			if a.dacEnabled[2] && trigger {
				a.wavePosition[2] = 0
				a.frequencyTimer[2] = maxPeriod + 1 - a.period[2]
				a.channelsOn[2] = true
			}
		}
		return true
	case addrNr41:
		if a.audioEnabled {
			a.lengthTimer[3] = 64 - uint16(v&0x3f)
		}
		return true
	case addrNr42:
		if a.audioEnabled {
			a.volumeEnvelope[3] = v
			if v&0xf8 == 0 {
				a.dacEnabled[3] = false
				a.channelsOn[3] = false
			} else {
				a.dacEnabled[3] = true
			}
		}
		return true
	case addrNr43:
		if a.audioEnabled {
			a.ch4Randomness = v
		}
		return true
	case addrNr44:
		if a.audioEnabled {
			// Set timer enable
			a.timerEnabled[3] = isBitSet(v, 6)
			// Trigger channel
			if a.timerEnabled[3] && a.lengthTimer[3] == 0 {
				a.lengthTimer[3] = 64 // Timer 0 means max length
			}
			trigger := isBitSet(v, 7)
			if a.dacEnabled[3] && trigger {
				a.ch4Lfsr = 0x7fff
				a.volume[3] = (a.volumeEnvelope[3] & 0xf0) >> 4
				a.channelsOn[3] = true
			}
		}
		return true
	case addrNr50:
		if a.audioEnabled {
			a.masterVolume = v
		}
		return true
	case addrNr51:
		if a.audioEnabled {
			a.panning = v
		}
		return true
	case addrNr52:
		a.audioEnabled = isBitSet(v, 7)
		if !a.audioEnabled {
			a.powerOff()
		}
		return true
	}
	return false
}

func (a *Apu) powerOff() {
	a.audioEnabled = false
	a.masterVolume = 0
	a.panning = 0
	a.channelsOn = [4]bool{false, false, false, false}
	a.dacEnabled = [4]bool{false, false, false, false}
	a.lengthTimer = [4]uint16{0, 0, 0, 0}
	a.timerEnabled = [4]bool{false, false, false, false}
	a.wavePosition = [3]int{0, 0, 0}
	a.volume = [4]byte{0, 0, 0, 0}
	a.volumeEnvelope = [4]byte{0, 0, 0, 0}
	a.period = [3]uint16{0, 0, 0}

	a.ch1SweepEnabled = false
	a.ch1SweepShadowPeriod = 0
	a.ch1Sweep = 0
	a.ch1SweepTimer = 0
	a.ch1DutyCycle = 0

	a.ch2DutyCycle = 0

	a.ch4Randomness = 0
	a.ch4Lfsr = 0

	a.samples = a.samples[:0]
}
