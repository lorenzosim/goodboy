package main

import (
	"math"
)

var pulseWaves = [][]byte{
	{0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 1, 1, 1},
	{0, 1, 1, 1, 1, 1, 1, 0},
}

const maxSamples = 1000
const samplesPerRead = 10
const ticksPerSample = clockFreq * 2 / audioSampleRate // *2 because APU runs at 2Mhz

// Read fills the buffer with 32 bit stereo PCM audio samples
func (a *Apu) Read(buf []byte) (int, error) {
	a.samplesMu.Lock()
	if len(a.samples) == 0 {
		a.samplesMu.Unlock()
		return 0, nil
	}
	numSamples := min(len(buf), len(a.samples), samplesPerRead)
	for i := 0; i < numSamples; i++ {
		addSampleToBuf(buf, i, a.samples[i])
	}
	a.samples = a.samples[numSamples:]
	a.samplesMu.Unlock()
	return 8 * numSamples, nil
}

// Tick advances the APU one step, which should be called at 2Mhz.
// This generates new audio samples and moves the position of the current wave in each channel.
func (a *Apu) Tick() {
	a.tick += 1

	// Produce an audio sample
	if a.tick%ticksPerSample == 0 {
		sample := a.genSample()
		a.samplesMu.Lock()
		if len(a.samples) > maxSamples {
			// TODO: This is a very crude way to do this...use a better method
			a.samples = a.samples[:0]
		}
		a.samples = append(a.samples, sample)
		a.samplesMu.Unlock()
	}

	// Update the state of each channel
	a.tickChannel(2, 32)
	if a.tick%2 == 0 {
		// Channel 1-3 operate at 1Mhz, so do this every other time.
		a.tickChannel(0, 8)
		a.tickChannel(1, 8)
		a.tickNoiseChannel()
	}
}

func (a *Apu) tickChannel(chNum int, waveLen int) {
	a.frequencyTimer[chNum]--
	if a.frequencyTimer[chNum] == 0 {
		a.frequencyTimer[chNum] = maxPeriod + 1 - a.period[chNum]
		a.wavePosition[chNum] = (a.wavePosition[chNum] + 1) % waveLen
	}
}

func (a *Apu) tickNoiseChannel() {
	a.frequencyTimer[3]--
	if a.frequencyTimer[3] == 0 {
		// Note: since we do this at 1Mhz, the divider will be 1/4 of the divisor code map shown in various websites.
		// For example, divisor code 6 will be 24, not 96.
		divisorCode := int(a.ch4Randomness & 0x7)
		shift := (a.ch4Randomness & 0xf0) >> 4
		shortMode := isBitSet(a.ch4Randomness, 3)
		if divisorCode == 0 {
			a.frequencyTimer[3] = 2
		} else {
			a.frequencyTimer[3] = uint16(divisorCode << 2)
		}
		a.frequencyTimer[3] <<= shift

		xorBit := (a.ch4Lfsr & 1) ^ ((a.ch4Lfsr & 2) >> 1)
		a.ch4Lfsr = a.ch4Lfsr>>1 | xorBit<<14
		if shortMode {
			a.ch4Lfsr &= ^uint16(1 << 6)
			a.ch4Lfsr |= xorBit << 6
		}
	}
}

func (a *Apu) genSample() AudioSample {
	var sample AudioSample

	if !a.audioEnabled {
		return sample
	}
	for i := 0; i < 4; i++ {
		a.mixChannelOnSample(&sample, i)
	}

	leftVolume := float32((a.masterVolume&0x70)>>4+1) / 8
	rightVolume := float32(a.masterVolume&0x7+1) / 8

	// Normalize the volume by diving by 4, since we are adding up 4 channels.
	sample.left = sample.left / 4 * leftVolume
	sample.right = sample.right / 4 * rightVolume
	return sample
}

func (a *Apu) mixChannelOnSample(sample *AudioSample, chNum int) {
	if !a.channelsOn[chNum] || a.volume[chNum] == 0 {
		return
	}

	var s float32
	switch chNum {
	case 0:
		s = float32(pulseWaves[a.ch1DutyCycle][a.wavePosition[0]]) * float32(a.volume[0])
	case 1:
		s = float32(pulseWaves[a.ch2DutyCycle][a.wavePosition[1]]) * float32(a.volume[1])
	case 2:
		s = float32(a.getWaveSample())
	case 3:
		s = float32(^a.ch4Lfsr&1) * float32(a.volume[3])
	}

	s /= 0xf // Normalize sample between 0 and 1
	if isBitSet(a.panning, 4+chNum) {
		sample.left += s
	}
	if isBitSet(a.panning, chNum) {
		sample.right += s
	}
}

func (a *Apu) getWaveSample() byte {
	// Two 4-bit samples are stored in each byte, so pick the byte first and then the nibble
	waveSample := a.ch3Wave[a.wavePosition[2]/2]
	if a.wavePosition[2]%2 == 0 {
		waveSample = (waveSample & 0xf0) >> 4
	} else {
		waveSample = waveSample & 0xf
	}
	// Unlike other channels, there's only 4 volume levels, 100% requires no changes and 0% is already handled at the
	// top, so we care only about 50% and 25%.
	if a.volume[2] == 2 {
		waveSample = waveSample >> 1
	} else if a.volume[2] == 3 {
		waveSample = waveSample >> 2
	}
	return waveSample
}

func addSampleToBuf(buf []byte, pos int, audioSample AudioSample) {
	// 4 bytes per channel as a float 32
	leftBits := math.Float32bits(audioSample.left)
	buf[8*pos] = byte(leftBits)
	buf[8*pos+1] = byte(leftBits >> 8)
	buf[8*pos+2] = byte(leftBits >> 16)
	buf[8*pos+3] = byte(leftBits >> 24)

	rightBits := math.Float32bits(audioSample.right)
	buf[8*pos+4] = byte(rightBits)
	buf[8*pos+5] = byte(rightBits >> 8)
	buf[8*pos+6] = byte(rightBits >> 16)
	buf[8*pos+7] = byte(rightBits >> 24)
}
