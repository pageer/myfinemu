package core

import (
	"errors"
)

func (c *CPU) runOpcode(operation Instruction) (bool, error) {
	callback := c.getOpcodeImpl(operation.name)
	return callback(c, operation.mode)
}

func (c *CPU) getOpcodeImpl(operation string) func(*CPU, AddressMode) (bool, error) {
	switch operation {
	case "ADC":
		// "Add with carry" operation.
		// Add the parameter value and the carry bit to the accumulator
		// and store the result back to A register.  If there's an overflow,
		// set the carry bit.
		// Example: If A=#80 and the carry bit is 1, then "ADC $#80" gives A=#02
		// and carry bit 1.
		return func(c *CPU, mode AddressMode) (bool, error) {
			value_address := c.getParameterAddress(mode)
			value := c.memory[value_address]

			carry_bit := uint8(0)
			if c.status&C_BIT_STATUS > 0 {
				carry_bit = uint8(1)
			}

			result, carry := addWithCarry(c.accumulator, value, carry_bit)

			c.accumulator = result
			if carry {
				c.setFlag(C_BIT_STATUS)
			} else {
				c.clearFlag(C_BIT_STATUS)
			}
			c.updateStatusFlags(result)
			return true, nil
		}
	case "LDA":
		// "Load accumulator" operation.
		// Stores the parameter into the A register.
		return func(c *CPU, mode AddressMode) (bool, error) {
			value_address := c.getParameterAddress(mode)
			value := c.memory[value_address]
			c.accumulator = value
			c.updateStatusFlags(value)
			return true, nil
		}

	case "TAX":
		// "Transfer A to X"
		// Copies the value in the accumulator to the index X register.
		return func(c *CPU, mode AddressMode) (bool, error) {
			value := c.accumulator
			c.index_x = value
			c.updateStatusFlags(c.index_x)
			return true, nil
		}

	case "INX":
		// "Increment X"
		// Add 1 to the X register with overflow (no carry)
		return func(c *CPU, mode AddressMode) (bool, error) {
			if c.index_x == 0xff {
				c.index_x = 0x00
			} else {
				c.index_x = c.index_x + 1
			}
			c.updateStatusFlags(c.index_x)
			return true, nil
		}

	case "INY":
		// "Increment Y"
		// Add 1 to the Y register with overflow (no carry)
		return func(c *CPU, mode AddressMode) (bool, error) {
			if c.index_y == 0xff {
				c.index_y = 0x00
			} else {
				c.index_y = c.index_y + 1
			}
			c.updateStatusFlags(c.index_y)
			return true, nil
		}

	case "BRK":
		// "Break"
		// Stop execution
		return func(c *CPU, mode AddressMode) (bool, error) {
			return false, nil
		}
	}

	return func(c *CPU, mode AddressMode) (bool, error) {
		return false, errors.New("Unimplemented opcode")
	}
}

func addWithCarry(acc uint8, val uint8, carry uint8) (uint8, bool) {
	result := uint16(acc) + uint16(val) + uint16(carry)
	if result > 0xff {
		return uint8(result & 0xff), true
	}

	return uint8(result), false
}
