package main

type Cpu struct {
	// Reference to memory controller and interrupt helper
	mcu    *Mcu
	interr *Interrupts
	// Immutable operations (micro-instructions) to execute for each opcode (and CB opcode)
	regOps [][]func()
	cbOps  [][]func()
	// Registers
	a, b, c, d, e, h, l byte
	// Flags (zero, subtraction, half carry, carry)
	fz, fn, fh, fc bool
	// Stack pointer
	sp uint16
	// Program counter
	pc uint16
	// Interrupt master enable
	ime bool
	// paused indicates whether the CPU is stopped (see halt mode)
	paused bool
	// List of operations that are pending
	pendingOps []func()
	// Internal registers, used to save data between micro-operations.
	z, w byte
	// Whether we print each instruction for debugging
	trace bool
}

func CreateCpu(mcu *Mcu, interrupts *Interrupts, trace bool) *Cpu {
	cpu := Cpu{mcu: mcu, interr: interrupts, sp: 0xfffe, trace: trace}
	cpu.regOps = cpu.makeRegOps()
	cpu.cbOps = cpu.makeCbOps()
	return &cpu
}

func (cpu *Cpu) Tick() {
	if len(cpu.pendingOps) > 0 {
		cpu.pendingOps[0]()
		cpu.pendingOps = cpu.pendingOps[1:]
		return
	}

	// Handle any interrupts.
	interrIndex := cpu.interr.ActiveInterruptIndex()
	if interrIndex >= 0 {
		cpu.paused = false
		if cpu.ime {
			cpu.callInterrupt(interrIndex)
			return
		}
	}

	if cpu.paused {
		return
	}

	if cpu.trace {
		cpu.PrintInstr()
	}
	cpu.execOpcode()
}

func (cpu *Cpu) execOpcode() {
	opcode := cpu.mcu.Get(cpu.pc)
	cpu.pc += 1

	if opcode == 0xcb {
		cpu.pendingOps = cpu.cbOps[cpu.mcu.Get(cpu.pc)]
		cpu.pc += 1
	} else {
		cpu.pendingOps = cpu.regOps[opcode]
		cpu.pendingOps[0]()
		cpu.pendingOps = cpu.pendingOps[1:]
	}
}

func (cpu *Cpu) callInterrupt(index int) {
	cpu.ime = false
	cpu.interr.interruptFlag = clearBit(cpu.interr.interruptFlag, index)
	cpu.pendingOps = call(cpu, interruptAddresses[index])
}

// Registers

func (cpu *Cpu) bc() uint16 {
	return merge(cpu.b, cpu.c)
}

func (cpu *Cpu) setBc(value uint16) {
	cpu.b, cpu.c = split(value)
}

func (cpu *Cpu) de() uint16 {
	return merge(cpu.d, cpu.e)
}

func (cpu *Cpu) setDe(value uint16) {
	cpu.d, cpu.e = split(value)
}

func (cpu *Cpu) hl() uint16 {
	return merge(cpu.h, cpu.l)
}

func (cpu *Cpu) setHl(value uint16) {
	cpu.h, cpu.l = split(value)
}

// Operations

func (cpu *Cpu) setCarryFlag(v uint16) {
	cpu.fc = v > 0xff
}

func (cpu *Cpu) setHalfCarryFlagAdd(a uint8, b uint8, c uint8) {
	cpu.fh = (((a & 0xF) + (b & 0xF) + (c & 0xF)) & 0x10) == 0x10
}

func (cpu *Cpu) setHalfCarryFlagSub(a uint8, b uint8, c uint8) {
	cpu.fh = (((a & 0xF) - (b & 0xF) - (c & 0xF)) & 0x10) == 0x10
}

func (cpu *Cpu) inc(v uint8) uint8 {
	res := v + 1
	cpu.setHalfCarryFlagAdd(v, 1, 0)
	cpu.fz = res == 0
	cpu.fn = false
	return res
}

func (cpu *Cpu) dec(v uint8) uint8 {
	res := uint16(v) - uint16(1)
	cpu.setHalfCarryFlagSub(v, 1, 0)

	res8 := uint8(res)
	cpu.fz = res8 == 0
	cpu.fn = true
	return res8
}

// adc adds two numbers with carry
func (cpu *Cpu) adc(a uint8, b uint8) uint8 {
	var carry uint8 = 0
	if cpu.fc {
		carry = 1
	}

	res := uint16(a) + uint16(b) + uint16(carry)
	cpu.setHalfCarryFlagAdd(a, b, carry)
	cpu.setCarryFlag(res)

	res8 := uint8(res)
	cpu.fz = res8 == 0
	cpu.fn = false
	return res8
}

func (cpu *Cpu) add(a uint8, b uint8) uint8 {
	res := uint16(a) + uint16(b)
	cpu.setHalfCarryFlagAdd(a, b, 0)
	cpu.setCarryFlag(res)

	res8 := uint8(res)
	cpu.fz = res8 == 0
	cpu.fn = false
	return res8
}

func (cpu *Cpu) addHlHl() []func() {
	return []func(){
		func() { fz := cpu.fz; cpu.l = cpu.add(cpu.l, cpu.l); cpu.fz = fz },
		func() { fz := cpu.fz; cpu.h = cpu.adc(cpu.h, cpu.h); cpu.fz = fz },
	}
}

func (cpu *Cpu) addHlBc() []func() {
	return []func(){
		func() { fz := cpu.fz; cpu.l = cpu.add(cpu.l, cpu.c); cpu.fz = fz },
		func() { fz := cpu.fz; cpu.h = cpu.adc(cpu.h, cpu.b); cpu.fz = fz },
	}
}

func (cpu *Cpu) addHlDe() []func() {
	return []func(){
		func() { fz := cpu.fz; cpu.l = cpu.add(cpu.l, cpu.e); cpu.fz = fz },
		func() { fz := cpu.fz; cpu.h = cpu.adc(cpu.h, cpu.d); cpu.fz = fz },
	}
}

func (cpu *Cpu) addHlSp() []func() {
	return []func(){
		func() { fz := cpu.fz; cpu.l = cpu.add(cpu.l, lowNibble(cpu.sp)); cpu.fz = fz },
		func() { fz := cpu.fz; cpu.h = cpu.adc(cpu.h, highNibble(cpu.sp)); cpu.fz = fz },
	}
}

func (cpu *Cpu) addOffset(a uint16, b uint8) uint16 {
	sb := toSignedInt(b)
	res := int16(a) + int16(sb)
	cpu.fc = uint8(a) > uint8(res)
	cpu.setHalfCarryFlagAdd(uint8(a), b, 0)
	cpu.fz = false
	cpu.fn = false
	return uint16(res)
}

func (cpu *Cpu) sub(a uint8, b uint8) uint8 {
	return cpu.sub3(a, b, 0)
}

// subc substracts two numbers with carry
func (cpu *Cpu) subc(a uint8, b uint8) uint8 {
	if cpu.fc {
		return cpu.sub3(a, b, 1)
	}
	return cpu.sub3(a, b, 0)
}

func (cpu *Cpu) sub3(a uint8, b uint8, c uint8) uint8 {
	res := uint16(a) - uint16(b) - uint16(c)
	cpu.setHalfCarryFlagSub(a, b, c)
	cpu.fc = (uint16(b) + uint16(c)) > uint16(a)

	res8 := uint8(res)
	cpu.fz = res8 == 0
	cpu.fn = true
	return res8
}

func (cpu *Cpu) call(cond bool, addr uint16) {
	if cond {
		cpu.pendingOps = append(
			cpu.pendingOps,
			func() { cpu.push(highNibble(cpu.pc)) },
			func() { cpu.push(lowNibble(cpu.pc)) },
			func() { cpu.pc = addr })
	}
}

func (cpu *Cpu) absJump(cond bool, addr uint16) {
	if cond {
		cpu.pendingOps = append(cpu.pendingOps, func() { cpu.pc = addr })
	}
}

func (cpu *Cpu) relJump(cond bool, offset byte) {
	if cond {
		cpu.pendingOps = append(cpu.pendingOps, func() { cpu.pc = uint16(int32(cpu.pc) + toSignedInt(offset)) })
	}
}

func (cpu *Cpu) push(value byte) {
	cpu.sp--
	cpu.mcu.Set(cpu.sp, value)
}

func (cpu *Cpu) pop() byte {
	value := cpu.mcu.Get(cpu.sp)
	cpu.sp++
	return value
}

func (cpu *Cpu) ret(cond bool) {
	if cond {
		cpu.pendingOps = append(cpu.pendingOps,
			func() { cpu.z = cpu.pop() },
			func() { cpu.w = cpu.pop() },
			func() { cpu.pc = cpu.zw() })
	}
}

func (cpu *Cpu) flagsToByte() byte {
	var flags byte = 0
	if cpu.fz {
		flags |= 1 << 7
	}
	if cpu.fn {
		flags |= 1 << 6
	}
	if cpu.fh {
		flags |= 1 << 5
	}
	if cpu.fc {
		flags |= 1 << 4
	}
	return flags
}

func (cpu *Cpu) flagsFromByte(flags byte) {
	cpu.fz = isBitSet(flags, 7)
	cpu.fn = isBitSet(flags, 6)
	cpu.fh = isBitSet(flags, 5)
	cpu.fc = isBitSet(flags, 4)
}

// cp is like sub, but it does not store the result.
func (cpu *Cpu) cp(a uint8, b uint8) {
	cpu.sub(a, b)
}

func (cpu *Cpu) xor(a uint8, b uint8) uint8 {
	res := a ^ b
	cpu.fz = res == 0
	cpu.fh = false
	cpu.fc = false
	cpu.fn = false
	return res
}

func (cpu *Cpu) or(a uint8, b uint8) uint8 {
	res := a | b
	cpu.fz = res == 0
	cpu.fh = false
	cpu.fc = false
	cpu.fn = false
	return res
}

func (cpu *Cpu) and(a uint8, b uint8) uint8 {
	res := a & b
	cpu.fz = res == 0
	cpu.fh = true
	cpu.fc = false
	cpu.fn = false
	return res
}

// rr does a bit rotation right through carry
func (cpu *Cpu) rr(v byte) byte {
	res := v >> 1
	if cpu.fc {
		res |= 0x80
	}
	cpu.fc = (v & 1) != 0
	cpu.fz = res == 0
	cpu.fn = false
	cpu.fh = false
	return res
}

// rr does a bit rotation right
func (cpu *Cpu) rrc(v byte) byte {
	cpu.fc = (v & 1) != 0
	res := v >> 1
	if cpu.fc {
		res |= 0x80
	}
	cpu.fz = res == 0
	cpu.fn = false
	cpu.fh = false
	return res
}

// srl does a bit right shift
func (cpu *Cpu) srl(v byte) byte {
	cpu.fc = (v & 1) != 0
	res := v >> 1
	cpu.fz = res == 0
	cpu.fn = false
	cpu.fh = false
	return res
}

// sra does a bit right shift arithmetically
func (cpu *Cpu) sra(v byte) byte {
	cpu.fc = (v & 1) != 0
	res := (v >> 1) | (v & 0x80)
	cpu.fz = res == 0
	cpu.fn = false
	cpu.fh = false
	return res
}

// rl rotates left through carry
func (cpu *Cpu) rl(v byte) byte {
	oldCarry := cpu.fc
	cpu.fc = (v & 0x80) != 0
	res := v << 1
	if oldCarry {
		res |= 1
	}
	cpu.fz = res == 0
	cpu.fn = false
	cpu.fh = false
	return res
}

// rlc rotates left
func (cpu *Cpu) rlc(v byte) byte {
	cpu.fc = (v & 0x80) != 0
	res := v << 1
	if cpu.fc {
		res |= 1
	}
	cpu.fz = res == 0
	cpu.fn = false
	cpu.fh = false
	return res
}

// sla shift left
func (cpu *Cpu) sla(v byte) byte {
	cpu.fc = (v & 0x80) != 0
	res := v << 1
	cpu.fz = res == 0
	cpu.fn = false
	cpu.fh = false
	return res
}

func (cpu *Cpu) daa() {
	if cpu.fn {
		// Addition
		if cpu.fc {
			cpu.a -= 0x60
		}
		if cpu.fh {
			cpu.a -= 0x6
		}
	} else {
		// Subtraction case
		if cpu.fc || (cpu.a > 0x99) {
			cpu.a += 0x60
			cpu.fc = true
		}
		if cpu.fh || ((cpu.a & 0x0f) > 0x09) {
			cpu.a += 0x06
		}
	}

	cpu.fz = cpu.a == 0
	cpu.fh = false
}

func (cpu *Cpu) cpl() {
	cpu.a = ^cpu.a
	cpu.fn = true
	cpu.fh = true
}

// swap swaps low and high nibble
func (cpu *Cpu) swap(v byte) byte {
	lower := v & 0xf
	upper := v & 0xf0
	res := (lower << 4) | (upper >> 4)
	cpu.fz = res == 0
	cpu.fn = false
	cpu.fh = false
	cpu.fc = false
	return res
}

func (cpu *Cpu) bit(bit int, v byte) {
	cpu.fz = (v & (1 << bit)) == 0
	cpu.fn = false
	cpu.fh = true
}

func (cpu *Cpu) imm2z() {
	cpu.z = cpu.mcu.Get(cpu.pc)
	cpu.pc += 1
}

func (cpu *Cpu) imm2w() {
	cpu.w = cpu.mcu.Get(cpu.pc)
	cpu.pc += 1
}

func (cpu *Cpu) z2a() {
	cpu.a = cpu.z
}

func (cpu *Cpu) memHl2z() {
	cpu.z = cpu.mcu.Get(cpu.hl())
}

func (cpu *Cpu) zw() uint16 {
	return merge(cpu.w, cpu.z)
}

func call(cpu *Cpu, addr uint16) []func() {
	return []func(){
		nop,
		func() { cpu.push(highNibble(cpu.pc)) },
		func() { cpu.push(lowNibble(cpu.pc)) },
		func() { cpu.pc = addr }}
}

func nop() {}

func (cpu *Cpu) stop() {
	cpu.paused = true
	// TODO: should reset timer
}

func (cpu *Cpu) halt() {
	cpu.paused = true
}
