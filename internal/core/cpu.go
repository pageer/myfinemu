package core

import (
	"errors"
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
	PC_RESET_ADDRESS         = 0xfffc
)

type AddressMode int

const (
	// The parameter IS the value - #$01 = 1
	AddrImmediate AddressMode = iota
	// Single byte address - $c0 = value at address 0xc0
	AddrZeroPage
	// Single byte address, but reg X is added to it.
	// if x = 1, "$c0,x" = value at address 0xc1
	AddrZeroPageX
	// Two-byte absolute address - $c000 = value at address 0xc000
	AddrAbsolute
	// Two-byte absolute address with reg X added to it.
	// e.g. if x = 1, $c000,x = value at address 0xc001
	AddrAbsoluteX
	// Two-byte absolute address with reg Y added to it.
	// e.g. if y = 1, $c000,y = value at address 0xc001
	AddrAbsoluteY
	AddrIndirectX
	AddrIndirectY
	AddrImplied
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
		Instruction{"BRK", AddrImplied, 0x00, 1},
		Instruction{"ADC", AddrImmediate, 0x69, 2},
		Instruction{"ADC", AddrZeroPage, 0x65, 2},
		Instruction{"ADC", AddrZeroPageX, 0x75, 2},
		Instruction{"ADC", AddrAbsolute, 0x6d, 3},
		Instruction{"ADC", AddrAbsoluteX, 0x7d, 3},
		Instruction{"ADC", AddrAbsoluteY, 0x79, 3},
		Instruction{"ADC", AddrIndirectX, 0x61, 3},
		Instruction{"ADC", AddrIndirectY, 0x71, 3},
		Instruction{"LDA", AddrImmediate, 0xa9, 2},
		Instruction{"LDA", AddrZeroPage, 0xa5, 2},
		Instruction{"LDA", AddrZeroPageX, 0xb5, 2},
		Instruction{"LDA", AddrAbsolute, 0xad, 3},
		Instruction{"LDA", AddrAbsoluteX, 0xbd, 3},
		Instruction{"LDA", AddrAbsoluteY, 0xb9, 3},
		Instruction{"LDA", AddrIndirectX, 0xa1, 3},
		Instruction{"LDA", AddrIndirectY, 0xb1, 3},
		Instruction{"TAX", AddrImplied, 0xaa, 1},
		Instruction{"INX", AddrImplied, 0xe8, 1},
		Instruction{"INY", AddrImplied, 0xc8, 1},
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
		instruction := c.getNextInstructionByte()
		keepLooping, err = c.processInstruction(instruction)
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
	c.program_counter = ROM_SEGMENT_START
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
	high := uint8(value & 0xff00)
	c.memory[address] = low
	c.memory[address+1] = high
}

func (c *CPU) getNextInstructionByte() uint8 {
	result := c.memory[c.program_counter]
	c.program_counter++
	return result
}

func (c *CPU) processInstruction(instruction uint8) (bool, error) {
	operation := opcodes[instruction]
	//fmt.Println(instruction, operation)
	keepLooping, err := c.runOpcode(operation)
	c.program_counter += uint16(operation.size - 1)
	return keepLooping, err
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

// Gets the address to read for the parameter to an opcode.
func (c *CPU) getParameterAddress(mode AddressMode) uint16 {
	param_address := c.program_counter
	switch mode {
	case AddrImmediate:
		// In immediate mode, the next byte is the param address
		return param_address
	case AddrZeroPage:
		return uint16(c.memory[param_address])
	case AddrZeroPageX:
		//return modularAdd(c.memory[param_address], c.index_x)
		return modularAdd(c.memory[param_address], c.index_x)
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
