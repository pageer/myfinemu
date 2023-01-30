package core

import (
	"errors"
)

type InstructionPostProccessingMode int

const (
	InstructionHalt InstructionPostProccessingMode = iota
	InstructionContinue
	InstructionProgramCounterUpdated
)

func (c *CPU) runOpcode(operation Instruction) (InstructionPostProccessingMode, error) {
	callback := c.getOpcodeImpl(operation.name)
	return callback(c, operation.mode)
}

func (c *CPU) getOpcodeImpl(operation string) func(*CPU, AddressMode) (InstructionPostProccessingMode, error) {
	switch operation {
	case "ADC":
		// "Add with carry" operation.
		// Add the parameter value and the carry bit to the accumulator
		// and store the result back to A register.  If there's an overflow,
		// set the carry bit.
		// Example: If A=#80 and the carry bit is 1, then "ADC $#80" gives A=#02
		// and carry bit 1.
		// NOTE: Apparently the NES CPU doesn't have a decimal mode, so BCD is ignored.
		return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
			value_address := c.getParameterAddress(mode)
			value := c.memory[value_address]

			carry_bit := uint8(0)
			if c.status&C_BIT_STATUS > 0 {
				carry_bit = uint8(1)
			}

			result, carry := addWithCarry(c.accumulator, value, carry_bit)

			c.accumulator = result
			setCarryFlag(c, carry)
			c.updateStatusFlags(result)
			return InstructionContinue, nil
		}

	case "AND":
		return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
			value_address := c.getParameterAddress(mode)
			value := c.memory[value_address]
			c.accumulator = c.accumulator & value
			c.updateStatusFlags(c.accumulator)
			return InstructionContinue, nil
		}

	case "ASL":
		// "Arithmetic shift left" operation
		return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
			var (
				carry bool
				value uint8
			)
			if mode == AddrImplied {
				value, carry = shiftLeftWithCarry(c.accumulator)
				c.accumulator = value
				c.updateStatusFlags(c.accumulator)
			} else {
				value_address := c.getParameterAddress(mode)
				value, carry = shiftLeftWithCarry(c.memory[value_address])
				c.memory[value_address] = value
				c.updateStatusFlags(value)
			}
			setCarryFlag(c, carry)
			return InstructionContinue, nil
		}

	case "BCC":
		// "Branch if carry clear" operation
		return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
			value_address := c.getParameterAddress(mode)
			value := c.memory[value_address]
			if c.status&C_BIT_STATUS == 0 {
				c.program_counter += uint16(value)
				return InstructionProgramCounterUpdated, nil
			} else {
				return InstructionContinue, nil
			}
		}

	case "BCS":
		// "Branch if carry set" operation
		return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
			value_address := c.getParameterAddress(mode)
			value := c.memory[value_address]
			if c.status&C_BIT_STATUS > 0 {
				c.program_counter += uint16(value)
				return InstructionProgramCounterUpdated, nil
			} else {
				return InstructionContinue, nil
			}
		}

	case "LDA":
		// "Load accumulator" operation.
		// Stores the parameter into the A register.
		return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
			value_address := c.getParameterAddress(mode)
			value := c.memory[value_address]
			c.accumulator = value
			c.updateStatusFlags(value)
			return InstructionContinue, nil
		}

	case "TAX":
		// "Transfer A to X"
		// Copies the value in the accumulator to the index X register.
		return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
			value := c.accumulator
			c.index_x = value
			c.updateStatusFlags(c.index_x)
			return InstructionContinue, nil
		}

	case "INX":
		// "Increment X"
		// Add 1 to the X register with overflow (no carry)
		return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
			if c.index_x == 0xff {
				c.index_x = 0x00
			} else {
				c.index_x = c.index_x + 1
			}
			c.updateStatusFlags(c.index_x)
			return InstructionContinue, nil
		}

	case "INY":
		// "Increment Y"
		// Add 1 to the Y register with overflow (no carry)
		return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
			if c.index_y == 0xff {
				c.index_y = 0x00
			} else {
				c.index_y = c.index_y + 1
			}
			c.updateStatusFlags(c.index_y)
			return InstructionContinue, nil
		}

	case "BRK":
		// "Break"
		// Stop execution
		return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
			return InstructionHalt, nil
		}
	}

	return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
		return InstructionHalt, errors.New("Unimplemented opcode")
	}
}

func addWithCarry(acc uint8, val uint8, carry uint8) (uint8, bool) {
	result := uint16(acc) + uint16(val) + uint16(carry)
	return returnByteWithCarry(result)
}

func shiftLeftWithCarry(val uint8) (uint8, bool) {
	result := uint16(val) * 2
	return returnByteWithCarry(result)
}

func returnByteWithCarry(result uint16) (uint8, bool) {
	if result > 0xff {
		return uint8(result & 0xff), true
	}

	return uint8(result), false
}

func setCarryFlag(c *CPU, carry bool) {
	if carry {
		c.setFlag(C_BIT_STATUS)
	} else {
		c.clearFlag(C_BIT_STATUS)
	}
}
