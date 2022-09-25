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
	AddrImmediate AddressMode = iota
	AddrZeroPage
	AddrZeroPageX
	AddrAbsolute
	AddrAbsoluteX
	AddrAbsoluteY
	AddrIndirectX
	AddrIndirectY
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

func NewCPU() *CPU {
	c := &CPU{}
	return c
}

func (c *CPU) Run() {
	var keepLooping bool = true
	for keepLooping {
		instruction := c.getNextInstructionByte()
		keepLooping = c.processInstruction(instruction)
	}
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

func (c *CPU) processInstruction(instruction uint8) bool {
	switch instruction {
	case 0xa9: // LDA
		value := c.getNextInstructionByte()
		c.accumulator = value
		c.updateStatusFlags(value)

	case 0xaa: // TAX
		value := c.accumulator
		c.index_x = value
		c.updateStatusFlags(c.index_x)

	case 0xe8: // INX
		if c.index_x == 0xff {
			c.index_x = 0x01
		} else {
			c.index_x = c.index_x + 1
		}
		c.updateStatusFlags(c.index_x)

	case 0x00: // BRK
		return false
	}

	return true
}

func (c *CPU) updateStatusFlags(value uint8) {
	if value&NEG_BIT > 0 {
		c.status = c.status | N_BIT_STATUS
	} else {
		c.status = c.status & (0xff ^ N_BIT_STATUS)
	}
	if value == 0 {
		c.status = c.status | Z_BIT_STATUS
	} else {
		c.status = c.status & (0xff ^ Z_BIT_STATUS)
	}
}

// Gets the address to read for the parameter to an opcode.
func (c *CPU) getParameterAddress(mode AddressMode) uint16 {
	switch mode {
	case AddrImmediate:
		return c.program_counter
	case AddrZeroPage:
		return uint16(c.memory[c.program_counter])
	case AddrZeroPageX:
		return modularAdd(c.memory[c.program_counter], c.index_x)
	case AddrAbsolute:
		return c.readAddressValue(c.program_counter)
	case AddrAbsoluteX:
	case AddrAbsoluteY:
	case AddrIndirectX:
	case AddrIndirectY:
	}
	return 0
}

func modularAdd(x uint8, y uint8) uint16 {
	return ((uint16(x) + uint16(y)) % 0x0100)
}
