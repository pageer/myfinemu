package core

import (
	"errors"
	"pageer/myfinemu/internal/logging"
)

const (
	N_BIT_STATUS uint8 = 0b10000000
	V_BIT_STATUS uint8 = 0b01000000
	B_BIT_STATUS uint8 = 0b00010000
	D_BIT_STATUS uint8 = 0b00001000
	I_BIT_STATUS uint8 = 0b00000100
	Z_BIT_STATUS uint8 = 0b00000010
	C_BIT_STATUS uint8 = 0b00000001

	NEG_BIT uint8 = 0b10000000

	// Memory locations
	MEMORY_SIZE              = 0x010000
	ROM_SEGMENT_START uint16 = 0x8000
	STACK_START       uint16 = 0x0100
	PC_RESET_ADDRESS         = 0xfffc
)

type AddressMode int

const (
	// The parameter IS the value - #$01 = 1
	// This is also used for "relative" addressing mode for branch instructions,
	// as it amounts to the same thing.
	AddrImmediate AddressMode = iota
	// Single byte address - $c0 = value at address 0xc0
	AddrZeroPage
	// Single byte address, but reg X is added to it.
	// if x = 1, "$c0,x" = value at address 0xc1
	AddrZeroPageX
	// Single byte address, but reg y is added to it.
	// if x = 1, "$c0,x" = value at address 0xc1
	AddrZeroPageY
	// Two-byte absolute address - $c000 = value at address 0xc000
	AddrAbsolute
	// Two-byte absolute address with reg X added to it.
	// e.g. if x = 1, $c000,x = value at address 0xc001
	AddrAbsoluteX
	// Two-byte absolute address with reg Y added to it.
	// e.g. if y = 1, $c000,y = value at address 0xc001
	AddrAbsoluteY
	// Indexed indirect - take param and add X reg (with zero-page wrap-around) to get address.
	// e.g. if x = 1 and param = 2, address = 0x0003
	AddrIndirectX
	// Indirect indexed - similar to above, but reads address at param and then adds Y.
	// e.g. if y = 2, param = 0x10, and byte at 0x10 = 7, then the address is 0x009.
	AddrIndirectY
	// No explicit address needed - the parameter is implied by the instruction
	AddrImplied
	// Indirect address accessed via stack.  Only used for JMP instruction.
	AddrIndirect
)

type CPU struct {
	program_counter uint16
	stack_pointer   uint8
	accumulator     uint8
	index_x         uint8
	index_y         uint8
	status          uint8
	memory          [MEMORY_SIZE]uint8
}

type Instruction struct {
	name string
	mode AddressMode
	hex  uint8
	size uint
}

var opcodes map[uint8]Instruction

func init() {
	opcodeList := []Instruction{
		{"ADC", AddrImmediate, 0x69, 2},
		{"ADC", AddrZeroPage, 0x65, 2},
		{"ADC", AddrZeroPageX, 0x75, 2},
		{"ADC", AddrAbsolute, 0x6d, 3},
		{"ADC", AddrAbsoluteX, 0x7d, 3},
		{"ADC", AddrAbsoluteY, 0x79, 3},
		{"ADC", AddrIndirectX, 0x61, 3},
		{"ADC", AddrIndirectY, 0x71, 3},
		{"AND", AddrImmediate, 0x29, 2},
		{"AND", AddrZeroPage, 0x25, 2},
		{"AND", AddrZeroPageX, 0x35, 2},
		{"AND", AddrAbsolute, 0x2d, 3},
		{"AND", AddrAbsoluteX, 0x3d, 3},
		{"AND", AddrAbsoluteY, 0x39, 3},
		{"AND", AddrIndirectX, 0x21, 3},
		{"AND", AddrIndirectY, 0x31, 3},
		{"ASL", AddrImplied, 0x0a, 1},
		{"ASL", AddrZeroPage, 0x06, 2},
		{"ASL", AddrZeroPageX, 0x16, 2},
		{"ASL", AddrAbsolute, 0x0e, 3},
		{"ASL", AddrAbsoluteX, 0x1e, 3},
		{"BCC", AddrImmediate, 0x90, 2},
		{"BCS", AddrImmediate, 0xb0, 2},
		{"BEQ", AddrImmediate, 0xf0, 2},
		{"BIT", AddrZeroPage, 0x24, 2},
		{"BIT", AddrAbsolute, 0x2c, 3},
		{"BMI", AddrImmediate, 0x30, 2},
		{"BNE", AddrImmediate, 0xd0, 2},
		{"BPL", AddrImmediate, 0x10, 2},
		{"BRK", AddrImplied, 0x00, 1},
		{"BVC", AddrImmediate, 0x50, 2},
		{"BVS", AddrImmediate, 0x70, 2},
		{"CLC", AddrImmediate, 0x18, 1},
		{"CLD", AddrImmediate, 0xd8, 1},
		{"CLI", AddrImmediate, 0x58, 1},
		{"CLV", AddrImmediate, 0xb8, 1},
		{"CMP", AddrImmediate, 0xc9, 2},
		{"CMP", AddrZeroPage, 0xc5, 2},
		{"CMP", AddrZeroPageX, 0xd5, 2},
		{"CMP", AddrAbsolute, 0xcd, 3},
		{"CMP", AddrAbsoluteX, 0xdd, 3},
		{"CMP", AddrAbsoluteY, 0xd9, 3},
		{"CMP", AddrIndirectX, 0xc1, 2},
		{"CMP", AddrIndirectY, 0xd1, 2},
		{"CPX", AddrImmediate, 0xe0, 2},
		{"CPX", AddrZeroPage, 0xe4, 2},
		{"CPX", AddrAbsolute, 0xec, 3},
		{"CPY", AddrImmediate, 0xc0, 2},
		{"CPY", AddrZeroPage, 0xc4, 2},
		{"CPY", AddrAbsolute, 0xcc, 3},
		{"DEC", AddrZeroPage, 0xc6, 2},
		{"DEC", AddrZeroPageX, 0xd6, 2},
		{"DEC", AddrAbsolute, 0xce, 3},
		{"DEC", AddrAbsoluteX, 0xde, 3},
		{"DEX", AddrImplied, 0xca, 1},
		{"DEY", AddrImplied, 0x88, 1},
		{"EOR", AddrImmediate, 0x49, 2},
		{"EOR", AddrZeroPage, 0x45, 2},
		{"EOR", AddrZeroPageX, 0x55, 2},
		{"EOR", AddrAbsolute, 0x4d, 3},
		{"EOR", AddrAbsoluteX, 0x5d, 3},
		{"EOR", AddrAbsoluteY, 0x59, 3},
		{"EOR", AddrIndirectX, 0x41, 2},
		{"EOR", AddrIndirectY, 0x51, 2},
		{"INC", AddrZeroPage, 0xe6, 2},
		{"INC", AddrZeroPageX, 0xf6, 2},
		{"INC", AddrAbsolute, 0xee, 3},
		{"INC", AddrAbsoluteX, 0xfe, 3},
		{"INX", AddrImplied, 0xe8, 1},
		{"INY", AddrImplied, 0xc8, 1},
		{"JMP", AddrAbsolute, 0x4c, 3},
		{"JMP", AddrIndirect, 0x6c, 3},
		{"JSR", AddrAbsolute, 0x20, 3},
		{"LDA", AddrImmediate, 0xa9, 2},
		{"LDA", AddrZeroPage, 0xa5, 2},
		{"LDA", AddrZeroPageX, 0xb5, 2},
		{"LDA", AddrAbsolute, 0xad, 3},
		{"LDA", AddrAbsoluteX, 0xbd, 3},
		{"LDA", AddrAbsoluteY, 0xb9, 3},
		{"LDA", AddrIndirectX, 0xa1, 3},
		{"LDA", AddrIndirectY, 0xb1, 3},
		{"LDX", AddrImmediate, 0xa2, 2},
		{"LDX", AddrZeroPage, 0xa6, 2},
		{"LDX", AddrZeroPageY, 0xb6, 2},
		{"LDX", AddrAbsolute, 0xae, 3},
		{"LDX", AddrAbsoluteY, 0xbe, 3},
		{"TAX", AddrImplied, 0xaa, 1},
	}

	opcodes = make(map[uint8]Instruction)
	for _, value := range opcodeList {
		opcodes[value.hex] = value
	}
}

func NewCPU() *CPU {
	c := &CPU{}
	return c
}

func (c *CPU) Run() error {
	var keepLooping bool = true
	var err error
	for keepLooping {
		keepLooping, err = c.processNextInstruction()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CPU) LoadROM(memory []uint8) error {
	if len(memory) > MEMORY_SIZE {
		return errors.New("ROM image too big")
	}

	// There's probably a better way to do this...
	position := ROM_SEGMENT_START
	for _, value := range memory {
		c.memory[position] = value
		position++
	}

	c.writeAddressValue(PC_RESET_ADDRESS, ROM_SEGMENT_START)

	return nil
}

func (c *CPU) Reset() {
	c.stack_pointer = 0
	c.accumulator = 0
	c.index_y = 0
	c.index_y = 0
	c.status = 0

	c.program_counter = c.readAddressValue(PC_RESET_ADDRESS)
}

func (c *CPU) LoadAndReset(memory []uint8) error {
	err := c.LoadROM(memory)

	if err != nil {
		return err
	}

	c.Reset()
	return nil
}

// Read a little-endian 2-byte value from the given location
func (c *CPU) readAddressValue(address uint16) uint16 {
	low := c.memory[address]
	high := c.memory[address+1]
	return uint16(high)<<8 | uint16(low)
}

// Write a little-endian 2-byte value to the given location
func (c *CPU) writeAddressValue(address uint16, value uint16) {
	low := uint8(value & 0x00ff)
	high := uint8((value & 0xff00) >> 8)
	c.memory[address] = low
	c.memory[address+1] = high
}

func (c *CPU) pushStack(value uint8) {
	c.memory[STACK_START+uint16(c.stack_pointer)] = value
	c.stack_pointer++
}

func (c *CPU) pushStackAddress(address uint16) {
	c.writeAddressValue(STACK_START+uint16(c.stack_pointer), address)
	c.stack_pointer += 2
}

func (c *CPU) popStack() uint8 {
	c.stack_pointer--
	return c.memory[STACK_START+uint16(c.stack_pointer)]
}

func (c *CPU) processNextInstruction() (bool, error) {
	instruction := c.memory[c.program_counter]
	operation := opcodes[instruction]
	init_pc := c.program_counter
	c.program_counter++
	postProcessing, err := c.runOpcode(operation)
	if postProcessing != InstructionProgramCounterUpdated {
		c.program_counter += uint16(operation.size - 1)
	}
	logging.LogDebug("Instruction: %#v, PC start: %#v, PC end: %#v", operation, init_pc, c.program_counter)
	return postProcessing != InstructionHalt, err
}

func (c *CPU) updateStatusFlags(value uint8) {
	if value&NEG_BIT > 0 {
		c.setFlag(N_BIT_STATUS)
	} else {
		c.clearFlag(N_BIT_STATUS)
	}
	if value == 0 {
		c.setFlag(Z_BIT_STATUS)
	} else {
		c.clearFlag(Z_BIT_STATUS)
	}
}

func (c *CPU) setFlag(flag uint8) {
	c.status = c.status | flag
}

func (c *CPU) clearFlag(flag uint8) {
	c.status = c.status & (0xff ^ flag)
}

// Gets the value for the parameter to an opcode.
// This looks at the address mode and then looks up and
// returns the appropriate value from memory.
func (c *CPU) getParameterValue(mode AddressMode) uint16 {
	param_address := c.program_counter
	switch mode {
	case AddrImmediate:
		// In immediate mode, the next byte is the param
		return param_address
	case AddrZeroPage:
		return uint16(c.memory[param_address])
	case AddrZeroPageX:
		return modularAdd(c.memory[param_address], c.index_x)
	case AddrZeroPageY:
		return modularAdd(c.memory[param_address], c.index_y)
	case AddrAbsolute:
		return c.readAddressValue(param_address)
	case AddrAbsoluteX:
		return c.readAddressValue(param_address) + uint16(c.index_x)
	case AddrAbsoluteY:
		return c.readAddressValue(param_address) + uint16(c.index_y)
	case AddrIndirectX:
		// Get the parameter
		// Add X register to it, treating it as a zero-page address.
		// Read that address.  That's where our param lives.
		zero_page_addr := modularAdd(c.memory[param_address], c.index_x)
		return c.readAddressValue(zero_page_addr)
	case AddrIndirectY:
		// Get the parameter.  Treat it as a zero-page address.
		// Get the two-bytes at that zero-page. That's our base address.
		// Add the Y register to that address.  That's the param address.
		addr := c.readAddressValue(uint16(c.memory[param_address]))
		return uint16(c.index_y) + addr
	}
	return 0
}

func modularAdd(x uint8, y uint8) uint16 {
	return ((uint16(x) + uint16(y)) % 0x0100)
}
