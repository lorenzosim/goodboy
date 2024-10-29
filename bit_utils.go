package main

func merge(hi byte, lo byte) uint16 {
	return (uint16(hi) << 8) + uint16(lo)
}

func highNibble(value uint16) byte {
	return byte((value & 0xff00) >> 8)
}

func lowNibble(value uint16) byte {
	return byte(value & 0xff)
}

func split(value uint16) (byte, byte) {
	h := byte((value & 0xff00) >> 8)
	l := byte(value & 0xff)
	return h, l
}

func toSignedInt(value byte) int32 {
	if value < 128 {
		return int32(value)
	}
	return int32(value) - 256
}

func isBitSet(value byte, bit int) bool {
	return (value & (1 << bit)) != 0
}

func setBit(value byte, bit int) byte {
	return value | (1 << bit)
}

func clearBit(value byte, bit int) byte {
	return value & ^(1 << bit)
}

func setBitValue(value byte, bit int, bitSet bool) byte {
	if bitSet {
		return setBit(value, bit)
	}
	return clearBit(value, bit)
}
