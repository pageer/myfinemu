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
		// "Branch if carry clear" operation, branches if carry bit unset
		return generateBranchCallback(C_BIT_STATUS, false)

	case "BCS":
		// "Branch if carry set" operation, branches if carry bit set
		return generateBranchCallback(C_BIT_STATUS, true)

	case "BEQ":
		// "Branch if equal" operation, branches if zero bit set
		return generateBranchCallback(Z_BIT_STATUS, true)

	case "BIT":
		// "Bit test" operation, does AND with accumulator and sets Z, V, N bits
		return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
			value_address := c.getParameterAddress(mode)
			value := c.memory[value_address]
			result := c.accumulator & value
			c.updateStatusFlags(result)
			c.status = c.status | (result & V_BIT_STATUS)
			return InstructionContinue, nil
		}

	case "BMI":
		// "Branch if minus" operation, branches if nevatige bit set
		return generateBranchCallback(N_BIT_STATUS, true)

	case "BNE":
		// "Branch not equal" operation, branches if zero bit not set
		return generateBranchCallback(Z_BIT_STATUS, false)

	case "BPL":
		// "Branch if positive" operation, branches if negative bit not set
		return generateBranchCallback(N_BIT_STATUS, false)

	case "BVC":
		return generateBranchCallback(V_BIT_STATUS, false)
		// "Branch if overflow clear" operation, branches if overflow bit not set

	case "BVS":
		// "Branch if overflow set" operation, branches if overflow bit set
		return generateBranchCallback(V_BIT_STATUS, true)

	case "CLC":
		// "Clear cary" operation
		return generateClearCallback(C_BIT_STATUS)

	case "CLD":
		// "Clear decimal" operation
		return generateClearCallback(D_BIT_STATUS)

	case "CLI":
		// "Clear interrupt" operation
		return generateClearCallback(I_BIT_STATUS)

	case "CLV":
		// "Clear overflor" operation
		return generateClearCallback(V_BIT_STATUS)

	case "CMP":
		// "Compare" operation, sets flags as if subtrating from accumulator
		return generateCompareCallback(c.accumulator)

	case "CPX":
		// "Compare X" operation, sets flags as if subtrating from index_x
		return generateCompareCallback(c.index_x)

	case "CPY":
		// "Compare Y" operation, sets flags as if subtrating from index_y
		return generateCompareCallback(c.index_y)

	case "DEC":
		// "Decrement" operation, decrements memory location
		return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
			value_address := c.getParameterAddress(mode)
			value := c.memory[value_address]
			result := value - 0x01
			c.memory[value_address] = result
			c.updateStatusFlags(result)
			return InstructionContinue, nil
		}

	case "DEX":
		// "Decrement X register" operation
		return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
			result := c.index_x - 1
			c.index_x = result
			c.updateStatusFlags(result)
			return InstructionContinue, nil
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

func branchOnStatus(c *CPU, mode AddressMode, flag uint8, set bool) (InstructionPostProccessingMode, error) {
	var do_branch bool
	value_address := c.getParameterAddress(mode)
	value := c.memory[value_address]
	if set {
		do_branch = c.status&flag > 0
	} else {

		do_branch = c.status&flag == 0
	}
	if do_branch {
		c.program_counter += uint16(value)
		return InstructionProgramCounterUpdated, nil
	} else {
		return InstructionContinue, nil
	}
}

func generateBranchCallback(status uint8, set bool) func(*CPU, AddressMode) (InstructionPostProccessingMode, error) {
	return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
		return branchOnStatus(c, mode, status, set)
	}
}

func generateClearCallback(bit uint8) func(*CPU, AddressMode) (InstructionPostProccessingMode, error) {
	return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
		c.status = c.status & (bit ^ uint8(0xff))
		return InstructionContinue, nil
	}
}

func generateCompareCallback(register uint8) func(*CPU, AddressMode) (InstructionPostProccessingMode, error) {
	return func(c *CPU, mode AddressMode) (InstructionPostProccessingMode, error) {
		value_address := c.getParameterAddress(mode)
		value := c.memory[value_address]
		result := register - value
		c.updateStatusFlags(result)
		if register > value {
			c.status |= C_BIT_STATUS
		}
		return InstructionContinue, nil
	}
}
