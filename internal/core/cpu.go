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
	size uint
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
	opcodes := make(map[uint8]Instruction)
	opcodes[0x00] = Instruction{"BRK", AddrImplied, 1}
	opcodes[0x69] = Instruction{"ADC", AddrImmediate, 2}
	opcodes[0x65] = Instruction{"ADC", AddrZeroPage, 2}
	opcodes[0xa9] = Instruction{"LDA", AddrImmediate, 2}
	opcodes[0xa5] = Instruction{"LDA", AddrZeroPage, 2}
	opcodes[0xb5] = Instruction{"LDA", AddrZeroPageX, 2}
	opcodes[0xad] = Instruction{"LDA", AddrAbsolute, 3}
	opcodes[0xbd] = Instruction{"LDA", AddrAbsoluteX, 3}
	opcodes[0xb9] = Instruction{"LDA", AddrAbsoluteY, 3}
	opcodes[0xa1] = Instruction{"LDA", AddrIndirectX, 3}
	opcodes[0xb1] = Instruction{"LDA", AddrIndirectY, 3}
	opcodes[0xaa] = Instruction{"TAX", AddrImplied, 1}
	opcodes[0xe8] = Instruction{"INX", AddrImplied, 1}
	opcodes[0xc8] = Instruction{"INY", AddrImplied, 1}

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
		addr := c.readAddressValue(uint16(param_address))
		return uint16(c.index_y) + addr
	}
	return 0
}

func modularAdd(x uint8, y uint8) uint16 {
	return ((uint16(x) + uint16(y)) % 0x0100)
}
