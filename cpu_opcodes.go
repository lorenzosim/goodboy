package main

func (cpu *Cpu) makeRegOps() [][]func() {
	ops := [][]func(){
		// NOP
		{nop},
		// LD BC, n16
		{cpu.imm2z, cpu.imm2w, func() { cpu.setBc(cpu.zw()) }},
		// LD [BC], A
		{func() { cpu.mcu.Set(cpu.bc(), cpu.a) }, nop},
		// INC BC
		{func() { cpu.setBc(1 + cpu.bc()) }, nop},
		// INC B
		{func() { cpu.b = cpu.inc(cpu.b) }},
		// DEC B
		{func() { cpu.b = cpu.dec(cpu.b) }},
		// LD B, n8
		{cpu.imm2z, func() { cpu.b = cpu.z }},
		// RLCA
		{func() { cpu.a = cpu.rlc(cpu.a); cpu.fz = false }},
		// LD [a16], SP
		{cpu.imm2z, cpu.imm2w,
			func() { cpu.mcu.Set(cpu.zw(), lowNibble(cpu.sp)) },
			func() { cpu.mcu.Set(cpu.zw()+1, highNibble(cpu.sp)) },
			nop,
		},
		// ADD HL, BC
		cpu.addHlBc(),
		// LD A, [BC]
		{func() { cpu.z = cpu.mcu.Get(cpu.bc()) }, cpu.z2a},
		// DEC BC
		{func() { cpu.setBc(cpu.bc() - 1) }, nop},
		// INC C
		{func() { cpu.c = cpu.inc(cpu.c) }},
		// DEC C
		{func() { cpu.c = cpu.dec(cpu.c) }},
		// LD C, n8
		{cpu.imm2z, func() { cpu.c = cpu.z }},
		// RRCA
		{func() { cpu.a = cpu.rrc(cpu.a); cpu.fz = false }},
		// STOP n8
		{cpu.imm2z, cpu.stop},
		// LD DE, n16
		{cpu.imm2z, cpu.imm2w, func() { cpu.setDe(cpu.zw()) }},
		// LD [DE], A
		{func() { cpu.mcu.Set(cpu.de(), cpu.a) }, nop},
		// INC DE
		{func() { cpu.setDe(1 + cpu.de()) }, nop},
		// INC D
		{func() { cpu.d = cpu.inc(cpu.d) }},
		// DEC D
		{func() { cpu.d = cpu.dec(cpu.d) }},
		// LD D, n8
		{cpu.imm2z, func() { cpu.d = cpu.z }},
		// RLA
		{func() { cpu.a = cpu.rl(cpu.a); cpu.fz = false }},
		// JR e8
		{cpu.imm2z, func() { cpu.relJump(true, cpu.z) }},
		// ADD HL, DE
		cpu.addHlDe(),
		// LD A, [DE]
		{func() { cpu.z = cpu.mcu.Get(cpu.de()) }, cpu.z2a},
		// DEC DE
		{func() { cpu.setDe(cpu.de() - 1) }, nop},
		// INC E
		{func() { cpu.e = cpu.inc(cpu.e) }},
		// DEC E
		{func() { cpu.e = cpu.dec(cpu.e) }},
		// LD E, n8
		{cpu.imm2z, func() { cpu.e = cpu.z }},
		// RRA
		{func() { cpu.a = cpu.rr(cpu.a); cpu.fz = false }},
		// JR NZ, e8
		{cpu.imm2z, func() { cpu.relJump(!cpu.fz, cpu.z) }},
		// LD HL, n16
		{cpu.imm2z, cpu.imm2w, func() { cpu.setHl(cpu.zw()) }},
		// LD [HL+], A
		{func() { cpu.mcu.Set(cpu.hl(), cpu.a) }, func() { cpu.setHl(cpu.hl() + 1) }},
		// INC HL
		{func() { cpu.setHl(1 + cpu.hl()) }, nop},
		// INC H
		{func() { cpu.h = cpu.inc(cpu.h) }},
		// DEC H
		{func() { cpu.h = cpu.dec(cpu.h) }},
		// LD H, n8
		{cpu.imm2z, func() { cpu.h = cpu.z }},
		// DAA
		{cpu.daa},
		// JR Z, e8
		{cpu.imm2z, func() { cpu.relJump(cpu.fz, cpu.z) }},
		// ADD HL, HL
		cpu.addHlHl(),
		// LD A, [HL+]
		{func() { cpu.z = cpu.mcu.Get(cpu.hl()); cpu.setHl(cpu.hl() + 1) }, cpu.z2a},
		// DEC HL
		{func() { cpu.setHl(cpu.hl() - 1) }, nop},
		// INC L
		{func() { cpu.l = cpu.inc(cpu.l) }},
		// DEC L
		{func() { cpu.l = cpu.dec(cpu.l) }},
		// LD L, n8
		{cpu.imm2z, func() { cpu.l = cpu.z }},
		// CPL
		{cpu.cpl},
		// JR NC, e8
		{cpu.imm2z, func() { cpu.relJump(!cpu.fc, cpu.z) }},
		// LD SP, n16
		{cpu.imm2z, cpu.imm2w, func() { cpu.sp = cpu.zw() }},
		// LD [HL-], A
		{func() { cpu.mcu.Set(cpu.hl(), cpu.a) }, func() { cpu.setHl(cpu.hl() - 1) }},
		// INC SP
		{func() { cpu.sp = 1 + cpu.sp }, nop},
		// INC [HL]
		{cpu.memHl2z, func() { cpu.mcu.Set(cpu.hl(), cpu.inc(cpu.z)) }, nop},
		// DEC [HL]
		{cpu.memHl2z, func() { cpu.mcu.Set(cpu.hl(), cpu.dec(cpu.z)) }, nop},
		// LD [HL], n8
		{cpu.imm2z, func() { cpu.mcu.Set(cpu.hl(), cpu.z) }, nop},
		// SCF
		{func() { cpu.fn = false; cpu.fh = false; cpu.fc = true }},
		// JR C, e8
		{cpu.imm2z, func() { cpu.relJump(cpu.fc, cpu.z) }},
		// ADD HL, SP
		cpu.addHlSp(),
		// LD A, [HL-]
		{func() { cpu.memHl2z(); cpu.setHl(cpu.hl() - 1) }, cpu.z2a},
		// DEC SP
		{func() { cpu.sp = cpu.sp - 1 }, nop},
		// INC A
		{func() { cpu.a = cpu.inc(cpu.a) }},
		// DEC A
		{func() { cpu.a = cpu.dec(cpu.a) }},
		// LD A, n8
		{cpu.imm2z, cpu.z2a},
		// CCF
		{func() { cpu.fn = false; cpu.fh = false; cpu.fc = !cpu.fc }},
		// LD B, B
		{func() { cpu.b = cpu.b }},
		// LD B, C
		{func() { cpu.b = cpu.c }},
		// LD B, D
		{func() { cpu.b = cpu.d }},
		// LD B, E
		{func() { cpu.b = cpu.e }},
		// LD B, H
		{func() { cpu.b = cpu.h }},
		// LD B, L
		{func() { cpu.b = cpu.l }},
		// LD B, [HL]
		{cpu.memHl2z, func() { cpu.b = cpu.z }},
		// LD B, A
		{func() { cpu.b = cpu.a }},
		// LD C, B
		{func() { cpu.c = cpu.b }},
		// LD C, C
		{func() { cpu.c = cpu.c }},
		// LD C, D
		{func() { cpu.c = cpu.d }},
		// LD C, E
		{func() { cpu.c = cpu.e }},
		// LD C, H
		{func() { cpu.c = cpu.h }},
		// LD C, L
		{func() { cpu.c = cpu.l }},
		// LD C, [HL]
		{cpu.memHl2z, func() { cpu.c = cpu.z }},
		// LD C, A
		{func() { cpu.c = cpu.a }},
		// LD D, B
		{func() { cpu.d = cpu.b }},
		// LD D, C
		{func() { cpu.d = cpu.c }},
		// LD D, D
		{func() { cpu.d = cpu.d }},
		// LD D, E
		{func() { cpu.d = cpu.e }},
		// LD D, H
		{func() { cpu.d = cpu.h }},
		// LD D, L
		{func() { cpu.d = cpu.l }},
		// LD D, [HL]
		{cpu.memHl2z, func() { cpu.d = cpu.z }},
		// LD D, A
		{func() { cpu.d = cpu.a }},
		// LD E, B
		{func() { cpu.e = cpu.b }},
		// LD E, C
		{func() { cpu.e = cpu.c }},
		// LD E, D
		{func() { cpu.e = cpu.d }},
		// LD E, E
		{func() { cpu.e = cpu.e }},
		// LD E, H
		{func() { cpu.e = cpu.h }},
		// LD E, L
		{func() { cpu.e = cpu.l }},
		// LD E, [HL]
		{cpu.memHl2z, func() { cpu.e = cpu.z }},
		// LD E, A
		{func() { cpu.e = cpu.a }},
		// LD H, B
		{func() { cpu.h = cpu.b }},
		// LD H, C
		{func() { cpu.h = cpu.c }},
		// LD H, D
		{func() { cpu.h = cpu.d }},
		// LD H, E
		{func() { cpu.h = cpu.e }},
		// LD H, H
		{func() { cpu.h = cpu.h }},
		// LD H, L
		{func() { cpu.h = cpu.l }},
		// LD H, [HL]
		{cpu.memHl2z, func() { cpu.h = cpu.z }},
		// LD H, A
		{func() { cpu.h = cpu.a }},
		// LD L, B
		{func() { cpu.l = cpu.b }},
		// LD L, C
		{func() { cpu.l = cpu.c }},
		// LD L, D
		{func() { cpu.l = cpu.d }},
		// LD L, E
		{func() { cpu.l = cpu.e }},
		// LD L, H
		{func() { cpu.l = cpu.h }},
		// LD L, L
		{func() { cpu.l = cpu.l }},
		// LD L, [HL]
		{cpu.memHl2z, func() { cpu.l = cpu.z }},
		// LD L, A
		{func() { cpu.l = cpu.a }},
		// LD [HL], B
		{func() { cpu.mcu.Set(cpu.hl(), cpu.b) }, nop},
		// LD [HL], C
		{func() { cpu.mcu.Set(cpu.hl(), cpu.c) }, nop},
		// LD [HL], D
		{func() { cpu.mcu.Set(cpu.hl(), cpu.d) }, nop},
		// LD [HL], E
		{func() { cpu.mcu.Set(cpu.hl(), cpu.e) }, nop},
		// LD [HL], H
		{func() { cpu.mcu.Set(cpu.hl(), cpu.h) }, nop},
		// LD [HL], L
		{func() { cpu.mcu.Set(cpu.hl(), cpu.l) }, nop},
		// HALT
		{cpu.halt},
		// LD [HL], A
		{func() { cpu.mcu.Set(cpu.hl(), cpu.a) }, nop},
		// LD A, B
		{func() { cpu.a = cpu.b }},
		// LD A, C
		{func() { cpu.a = cpu.c }},
		// LD A, D
		{func() { cpu.a = cpu.d }},
		// LD A, E
		{func() { cpu.a = cpu.e }},
		// LD A, H
		{func() { cpu.a = cpu.h }},
		// LD A, L
		{func() { cpu.a = cpu.l }},
		// LD A, [HL]
		{cpu.memHl2z, cpu.z2a},
		// LD A, A
		{func() { cpu.a = cpu.a }},
		// ADD A, B
		{func() { cpu.a = cpu.add(cpu.a, cpu.b) }},
		// ADD A, C
		{func() { cpu.a = cpu.add(cpu.a, cpu.c) }},
		// ADD A, D
		{func() { cpu.a = cpu.add(cpu.a, cpu.d) }},
		// ADD A, E
		{func() { cpu.a = cpu.add(cpu.a, cpu.e) }},
		// ADD A, H
		{func() { cpu.a = cpu.add(cpu.a, cpu.h) }},
		// ADD A, L
		{func() { cpu.a = cpu.add(cpu.a, cpu.l) }},
		// ADD A, [HL]
		{cpu.memHl2z, func() { cpu.a = cpu.add(cpu.a, cpu.z) }},
		// ADD A, A
		{func() { cpu.a = cpu.add(cpu.a, cpu.a) }},
		// ADC A, B
		{func() { cpu.a = cpu.adc(cpu.a, cpu.b) }},
		// ADC A, C
		{func() { cpu.a = cpu.adc(cpu.a, cpu.c) }},
		// ADC A, D
		{func() { cpu.a = cpu.adc(cpu.a, cpu.d) }},
		// ADC A, E
		{func() { cpu.a = cpu.adc(cpu.a, cpu.e) }},
		// ADC A, H
		{func() { cpu.a = cpu.adc(cpu.a, cpu.h) }},
		// ADC A, L
		{func() { cpu.a = cpu.adc(cpu.a, cpu.l) }},
		// ADC A, [HL]
		{cpu.memHl2z, func() { cpu.a = cpu.adc(cpu.a, cpu.z) }},
		// ADC A, A
		{func() { cpu.a = cpu.adc(cpu.a, cpu.a) }},
		// SUB A, B
		{func() { cpu.a = cpu.sub(cpu.a, cpu.b) }},
		// SUB A, C
		{func() { cpu.a = cpu.sub(cpu.a, cpu.c) }},
		// SUB A, D
		{func() { cpu.a = cpu.sub(cpu.a, cpu.d) }},
		// SUB A, E
		{func() { cpu.a = cpu.sub(cpu.a, cpu.e) }},
		// SUB A, H
		{func() { cpu.a = cpu.sub(cpu.a, cpu.h) }},
		// SUB A, L
		{func() { cpu.a = cpu.sub(cpu.a, cpu.l) }},
		// SUB A, [HL]
		{cpu.memHl2z, func() { cpu.a = cpu.sub(cpu.a, cpu.z) }},
		// SUB A, A
		{func() { cpu.a = cpu.sub(cpu.a, cpu.a) }},
		// SBC A, B
		{func() { cpu.a = cpu.subc(cpu.a, cpu.b) }},
		// SBC A, C
		{func() { cpu.a = cpu.subc(cpu.a, cpu.c) }},
		// SBC A, D
		{func() { cpu.a = cpu.subc(cpu.a, cpu.d) }},
		// SBC A, E
		{func() { cpu.a = cpu.subc(cpu.a, cpu.e) }},
		// SBC A, H
		{func() { cpu.a = cpu.subc(cpu.a, cpu.h) }},
		// SBC A, L
		{func() { cpu.a = cpu.subc(cpu.a, cpu.l) }},
		// SBC A, [HL]
		{cpu.memHl2z, func() { cpu.a = cpu.subc(cpu.a, cpu.z) }},
		// SBC A, A
		{func() { cpu.a = cpu.subc(cpu.a, cpu.a) }},
		// AND A, B
		{func() { cpu.a = cpu.and(cpu.a, cpu.b) }},
		// AND A, C
		{func() { cpu.a = cpu.and(cpu.a, cpu.c) }},
		// AND A, D
		{func() { cpu.a = cpu.and(cpu.a, cpu.d) }},
		// AND A, E
		{func() { cpu.a = cpu.and(cpu.a, cpu.e) }},
		// AND A, H
		{func() { cpu.a = cpu.and(cpu.a, cpu.h) }},
		// AND A, L
		{func() { cpu.a = cpu.and(cpu.a, cpu.l) }},
		// AND A, [HL]
		{cpu.memHl2z, func() { cpu.a = cpu.and(cpu.a, cpu.z) }},
		// AND A, A
		{func() { cpu.a = cpu.and(cpu.a, cpu.a) }},
		// XOR A, B
		{func() { cpu.a = cpu.xor(cpu.a, cpu.b) }},
		// XOR A, C
		{func() { cpu.a = cpu.xor(cpu.a, cpu.c) }},
		// XOR A, D
		{func() { cpu.a = cpu.xor(cpu.a, cpu.d) }},
		// XOR A, E
		{func() { cpu.a = cpu.xor(cpu.a, cpu.e) }},
		// XOR A, H
		{func() { cpu.a = cpu.xor(cpu.a, cpu.h) }},
		// XOR A, L
		{func() { cpu.a = cpu.xor(cpu.a, cpu.l) }},
		// XOR A, [HL]
		{cpu.memHl2z, func() { cpu.a = cpu.xor(cpu.a, cpu.z) }},
		// XOR A, A
		{func() { cpu.a = cpu.xor(cpu.a, cpu.a) }},
		// OR A, B
		{func() { cpu.a = cpu.or(cpu.a, cpu.b) }},
		// OR A, C
		{func() { cpu.a = cpu.or(cpu.a, cpu.c) }},
		// OR A, D
		{func() { cpu.a = cpu.or(cpu.a, cpu.d) }},
		// OR A, E
		{func() { cpu.a = cpu.or(cpu.a, cpu.e) }},
		// OR A, H
		{func() { cpu.a = cpu.or(cpu.a, cpu.h) }},
		// OR A, L
		{func() { cpu.a = cpu.or(cpu.a, cpu.l) }},
		// OR A, [HL]
		{cpu.memHl2z, func() { cpu.a = cpu.or(cpu.a, cpu.z) }},
		// OR A, A
		{func() { cpu.a = cpu.or(cpu.a, cpu.a) }},
		// CP A, B
		{func() { cpu.cp(cpu.a, cpu.b) }},
		// CP A, C
		{func() { cpu.cp(cpu.a, cpu.c) }},
		// CP A, D
		{func() { cpu.cp(cpu.a, cpu.d) }},
		// CP A, E
		{func() { cpu.cp(cpu.a, cpu.e) }},
		// CP A, H
		{func() { cpu.cp(cpu.a, cpu.h) }},
		// CP A, L
		{func() { cpu.cp(cpu.a, cpu.l) }},
		// CP A, [HL]
		{cpu.memHl2z, func() { cpu.cp(cpu.a, cpu.z) }},
		// CP A, A
		{func() { cpu.cp(cpu.a, cpu.a) }},
		// RET NZ
		{nop, func() { cpu.ret(!cpu.fz) }},
		// POP BC
		{func() { cpu.z = cpu.pop() }, func() { cpu.w = cpu.pop() }, func() { cpu.setBc(cpu.zw()) }},
		// JP NZ, a16
		{cpu.imm2z, cpu.imm2w, func() { cpu.absJump(!cpu.fz, cpu.zw()) }},
		// JP a16
		{cpu.imm2z, cpu.imm2w, func() { cpu.absJump(true, cpu.zw()) }},
		// CALL NZ, a16
		{cpu.imm2z, cpu.imm2w, func() { cpu.call(!cpu.fz, cpu.zw()) }},
		// PUSH BC
		{nop, func() { cpu.push(cpu.b) }, func() { cpu.push(cpu.c) }, nop},
		// ADD A, n8
		{cpu.imm2z, func() { cpu.a = cpu.add(cpu.a, cpu.z) }},
		// RST $00
		call(cpu, 0x00),
		// RET Z
		{nop, func() { cpu.ret(cpu.fz) }},
		// RET
		{func() { cpu.z = cpu.pop() }, func() { cpu.w = cpu.pop() }, func() { cpu.pc = cpu.zw() }, nop},
		// JP Z, a16
		{cpu.imm2z, cpu.imm2w, func() { cpu.absJump(cpu.fz, cpu.zw()) }},
		// PREFIX
		{func() { panic("Prefix should be handled separately") }},
		// CALL Z, a16
		{cpu.imm2z, cpu.imm2w, func() { cpu.call(cpu.fz, cpu.zw()) }},
		// CALL a16
		{cpu.imm2z, cpu.imm2w, func() { cpu.call(true, cpu.zw()) }},
		// ADC A, n8
		{cpu.imm2z, func() { cpu.a = cpu.adc(cpu.a, cpu.z) }},
		// RST $08
		call(cpu, 0x08),
		// RET NC
		{nop, func() { cpu.ret(!cpu.fc) }},
		// POP DE
		{func() { cpu.z = cpu.pop() }, func() { cpu.w = cpu.pop() }, func() { cpu.setDe(cpu.zw()) }},
		// JP NC, a16
		{cpu.imm2z, cpu.imm2w, func() { cpu.absJump(!cpu.fc, cpu.zw()) }},
		// 0xd3
		{func() { panic("Illegal instruction 0xd3") }},
		// CALL NC, a16
		{cpu.imm2z, cpu.imm2w, func() { cpu.call(!cpu.fc, cpu.zw()) }},
		// PUSH DE
		{nop, func() { cpu.push(cpu.d) }, func() { cpu.push(cpu.e) }, nop},
		// SUB A, n8
		{cpu.imm2z, func() { cpu.a = cpu.sub(cpu.a, cpu.z) }},
		// RST $10
		call(cpu, 0x10),
		// RET C
		{nop, func() { cpu.ret(cpu.fc) }},
		// RETI
		{func() { cpu.z = cpu.pop() }, func() { cpu.w = cpu.pop() }, func() { cpu.pc = cpu.zw() }, func() { cpu.ime = true }},
		// JP C, a16
		{cpu.imm2z, cpu.imm2w, func() { cpu.absJump(cpu.fc, cpu.zw()) }},
		// 0xdb
		{func() { panic("Illegal instruction 0xdb") }},
		// CALL C, a16
		{cpu.imm2z, cpu.imm2w, func() { cpu.call(cpu.fc, cpu.zw()) }},
		// 0xdd
		{func() { panic("Illegal instruction 0xdd") }},
		// SBC A, n8
		{cpu.imm2z, func() { cpu.a = cpu.subc(cpu.a, cpu.z) }},
		// RST $18
		call(cpu, 0x18),
		// LDH [a8], A
		{cpu.imm2z, func() { cpu.mcu.Set(0xff00|uint16(cpu.z), cpu.a) }, nop},
		// POP HL
		{func() { cpu.z = cpu.pop() }, func() { cpu.w = cpu.pop() }, func() { cpu.setHl(cpu.zw()) }},
		// LD [C], A
		{func() { cpu.mcu.Set(0xff00|uint16(cpu.c), cpu.a) }, nop},
		// 0xe3
		{func() { panic("Illegal instruction 0xe3") }},
		// 0xe4
		{func() { panic("Illegal instruction 0xe4") }},
		// PUSH HL
		{nop, func() { cpu.push(cpu.h) }, func() { cpu.push(cpu.l) }, nop},
		// AND A, n8
		{cpu.imm2z, func() { cpu.a = cpu.and(cpu.a, cpu.z) }},
		// RST $20
		call(cpu, 0x20),
		// ADD SP, e8
		{cpu.imm2z, func() { cpu.sp = cpu.addOffset(cpu.sp, cpu.z) }, nop, nop},
		// JP HL
		{func() { cpu.pc = cpu.hl() }},
		// LD [a16], A
		{cpu.imm2z, cpu.imm2w,
			func() { cpu.mcu.Set(cpu.zw(), cpu.a) },
			nop,
		},
		// 0xeb
		{func() { panic("Illegal instruction 0xeb") }},
		// 0xec
		{func() { panic("Illegal instruction 0xec") }},
		// 0xed
		{func() { panic("Illegal instruction 0xed") }},
		// XOR A, n8
		{cpu.imm2z, func() { cpu.a = cpu.xor(cpu.a, cpu.z) }},
		// RST $28
		call(cpu, 0x28),
		// LDH A, [a8]
		{cpu.imm2z, func() { cpu.z = cpu.mcu.Get(0xff00 | uint16(cpu.z)) }, cpu.z2a},
		// POP AF
		{func() { cpu.z = cpu.pop() }, func() { cpu.a = cpu.pop() }, func() { cpu.flagsFromByte(cpu.z) }},
		// LD A, [C]
		{func() { cpu.z = cpu.mcu.Get(0xff00 | uint16(cpu.c)) }, cpu.z2a},
		// DI
		{func() { cpu.ime = false }},
		// 0xf4
		{func() { panic("Illegal instruction 0xf4") }},
		// PUSH AF
		{nop, func() { cpu.push(cpu.a) }, func() { cpu.push(cpu.flagsToByte()) }, nop},
		// OR A, n8
		{cpu.imm2z, func() { cpu.a = cpu.or(cpu.a, cpu.z) }},
		// RST $30
		call(cpu, 0x30),
		// LD HL, SP+, e8
		{cpu.imm2z, func() { cpu.setHl(cpu.addOffset(cpu.sp, cpu.z)) }, nop},
		// LD SP, HL
		{func() { cpu.sp = cpu.hl() }, nop},
		// LD A, [a16]
		{cpu.imm2z, cpu.imm2w, func() { cpu.z = cpu.mcu.Get(cpu.zw()) }, cpu.z2a},
		// EI
		{func() { cpu.ime = true }},
		// 0xfc
		{func() { panic("Illegal instruction 0xfc") }},
		// 0xfd
		{func() { panic("Illegal instruction 0xfd") }},
		// CP A, n8
		{cpu.imm2z, func() { cpu.cp(cpu.a, cpu.z) }},
		// RST $38
		call(cpu, 0x38),
	}
	return ops
}

func (cpu *Cpu) makeCbOps() [][]func() {
	ops := make([][]func(), 0, 0x100)

	// RLC, RRC, RL, RR, SLA, SRA, SWAP, SRL
	funcs1 := []func(v byte) byte{cpu.rlc, cpu.rrc, cpu.rl, cpu.rr, cpu.sla, cpu.sra, cpu.swap, cpu.srl}
	for _, fun := range funcs1 {
		ops = append(ops,
			[]func(){func() { cpu.b = fun(cpu.b) }},
			[]func(){func() { cpu.c = fun(cpu.c) }},
			[]func(){func() { cpu.d = fun(cpu.d) }},
			[]func(){func() { cpu.e = fun(cpu.e) }},
			[]func(){func() { cpu.h = fun(cpu.h) }},
			[]func(){func() { cpu.l = fun(cpu.l) }},
			[]func(){cpu.memHl2z, func() { cpu.mcu.Set(cpu.hl(), fun(cpu.z)) }, nop},
			[]func(){func() { cpu.a = fun(cpu.a) }})
	}

	// BIT
	for bit := 0; bit <= 7; bit++ {
		ops = append(ops,
			[]func(){func() { cpu.bit(bit, cpu.b) }},
			[]func(){func() { cpu.bit(bit, cpu.c) }},
			[]func(){func() { cpu.bit(bit, cpu.d) }},
			[]func(){func() { cpu.bit(bit, cpu.e) }},
			[]func(){func() { cpu.bit(bit, cpu.h) }},
			[]func(){func() { cpu.bit(bit, cpu.l) }},
			[]func(){cpu.memHl2z, func() { cpu.bit(bit, cpu.z) }},
			[]func(){func() { cpu.bit(bit, cpu.a) }})
	}

	// RES, SET
	funcs2 := []func(v byte, bit int) byte{clearBit, setBit}
	for _, fun := range funcs2 {
		for bit := 0; bit <= 7; bit++ {
			ops = append(ops,
				[]func(){func() { cpu.b = fun(cpu.b, bit) }},
				[]func(){func() { cpu.c = fun(cpu.c, bit) }},
				[]func(){func() { cpu.d = fun(cpu.d, bit) }},
				[]func(){func() { cpu.e = fun(cpu.e, bit) }},
				[]func(){func() { cpu.h = fun(cpu.h, bit) }},
				[]func(){func() { cpu.l = fun(cpu.l, bit) }},
				[]func(){cpu.memHl2z, func() { cpu.mcu.Set(cpu.hl(), fun(cpu.z, bit)) }, nop},
				[]func(){func() { cpu.a = fun(cpu.a, bit) }})
		}
	}
	return ops
}
