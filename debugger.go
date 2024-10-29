package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Debugger struct {
	cpu         *Cpu
	breakpoints []uint16
	paused      bool
}

func (dbg *Debugger) Tick() {
	if len(dbg.cpu.pendingOps) > 0 {
		// Previous instruction is still in progress
		dbg.cpu.Tick()
		return
	}

	if slices.Contains(dbg.breakpoints, dbg.cpu.pc) {
		dbg.paused = true
	}

	if dbg.paused {
		dbg.cpu.PrintInstr()
		dbg.prompt()
	}

	dbg.cpu.Tick()
}

func (dbg *Debugger) prompt() {
	for {
		fmt.Print(">")
		reader := bufio.NewReader(os.Stdin)
		cmd, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		args := strings.Fields(cmd)
		if len(args) != 1 {
			println("Unknown command, try 'help'")
			return
		}

		switch args[0] {
		case "r", "run":
			dbg.paused = false
			return
		case "s", "step":
			dbg.paused = true
			return
		case "b", "break":
			if len(args) == 2 {
				addressStr := args[1]
				address, err := strconv.ParseUint(addressStr, 0, 16)
				if err != nil {
					fmt.Println("Invalid address: ", addressStr)
					continue
				}
				dbg.breakpoints = append(dbg.breakpoints, uint16(address))
			} else {
				println("Usage: b <address>")
			}
		case "i", "info":
			if len(args) == 2 {
				switch args[1] {
				case "b":
					if len(dbg.breakpoints) == 0 {
						println("No breakpoints")
					} else {
						for _, bp := range dbg.breakpoints {
							fmt.Printf("0x%04X\n", bp)
						}
					}
				default:
					println("Usage: info b")
				}
			} else {
				println("Usage: info b")
			}
		case "x":
			if len(args) == 2 {
				loc, err := dbg.resolveAddress(args[1])
				if err != nil {
					println("Invalid address. Try for example 0x100 or $HL.")
					continue
				}
				fmt.Printf("0x%04X: %02X\n", loc, dbg.cpu.mcu.Get(loc))
			} else {
				println("Usage: x <addr|$reg>")
			}
		case "q", "quit":
			os.Exit(0)
		case "h", "help":
			println("r or run - run the program")
			println("s or step - step by one instruction")
			println("b or break <addr> - sets a breakpoint at the given address")
			println("i b or info b - prints all the breakpoints")
			println("x <addr|$reg> - prints the memory at the given address (e.g. 0xff) or register (e.g. $HL)")
			println("q or quit - quit")
		default:
			println("Unknown command, try 'help'")
		}
	}
}

// PrintInstr prints the instruction at the current PC, along with the value of the registers, for debugging.
func (cpu *Cpu) PrintInstr() {
	opcode := cpu.mcu.Get(cpu.pc)
	var info opcodeInformation
	if opcode == 0xcb {
		info = cbOpcodeInfo[cpu.mcu.Get(cpu.pc+1)]
	} else {
		info = opcodeInfo[opcode]
	}
	var operands []string
	pos := cpu.pc + 1
	for _, operand := range info.operands {
		name, offset := operandName(cpu.mcu, operand.name, pos)
		pos += offset
		if operand.immediate {
			operands = append(operands, name)
		} else {
			operands = append(operands, fmt.Sprintf("(%s)", name))
		}
	}

	flags := cpu.flagsToByte()
	romBank := "--"
	if cpu.pc < 0x4000 {
		romBank = "00"
	} else if cpu.pc < 0x8000 {
		romBank = fmt.Sprintf("%02X", cpu.mcu.cartridge.currentRomBank)
	}

	fmt.Printf("A:%02X F:%02X B:%02X C:%02X D:%02X E:%02X H:%02X L:%02X SP:%04X [%s]0x%04x: %s %s\n",
		cpu.a, flags, cpu.b, cpu.c, cpu.d, cpu.e, cpu.h, cpu.l, cpu.sp,
		romBank, cpu.pc, info.mnemonic, strings.Join(operands, ","))
}

func (dbg *Debugger) resolveAddress(addr string) (uint16, error) {
	switch strings.ToUpper(addr) {
	case "$BC":
		return dbg.cpu.bc(), nil
	case "$DE":
		return dbg.cpu.de(), nil
	case "$HL":
		return dbg.cpu.hl(), nil
	default:
		userAddr, err := strconv.ParseUint(addr, 0, 16)
		if err != nil {
			return 0, fmt.Errorf("invalid address. Try 0x100 or $HL")
		}
		return uint16(userAddr), nil
	}
}

func operandName(mem *Mcu, name string, pos uint16) (string, uint16) {
	switch name {
	case "e8":
		return fmt.Sprintf("%d", toSignedInt(mem.Get(pos))), 1
	case "n8", "a8":
		return fmt.Sprintf("0x%x", mem.Get(pos)), 1
	case "n16", "a16":
		return fmt.Sprintf("0x%x", mem.GetWord(pos)), 2
	default:
		return name, 0
	}
}
