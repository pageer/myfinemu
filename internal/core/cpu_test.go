package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

const ZERO_BIT uint8 = 0

// Common structure and expectations for test data
type testInput struct {
	name                 string
	memory               []uint8
	upper_memory         []uint8
	rom                  []uint8
	initial_accumulator  uint8
	initial_index_x      uint8
	initial_index_y      uint8
	initial_status       uint8
	expected_accumulator uint8
	expected_index_x     uint8
	expected_index_y     uint8
	expected_status      uint8
}

// Set the CPU to the initial state from the test input
func initializeCpuState(c *CPU, test testInput) {
	for i, v := range test.memory {
		c.memory[i] = v
	}
	for i, v := range test.upper_memory {
		c.memory[i+0x1000] = v
	}

	c.accumulator = test.initial_accumulator
	c.index_x = test.initial_index_x
	c.index_y = test.initial_index_y
	c.status = test.initial_status
}

func TestRun_UndefinedOpcode(t *testing.T) {
	memory := []uint8{0xff}
	c := NewCPU()
	c.LoadAndReset(memory)
	result := c.Run()
	assert.Equal(t, errors.New("Unimplemented opcode"), result)
}

func TestRun_Scenario(t *testing.T) {
	testCases := []testInput{
		{
			name:                 "LDA TAX INX BRK",
			memory:               []uint8{0xa9, 0xc0, 0xaa, 0xe8, 0x00},
			expected_accumulator: 0xc0,
			expected_index_x:     0xc1,
			expected_status:      N_BIT_STATUS,
		},
	}

	for _, test := range testCases {
		callback := func(t *testing.T) {
			c := NewCPU()
			c.LoadAndReset(test.memory)
			c.Run()

			assert.Equal(t, test.expected_accumulator, c.accumulator, "Accumulator incorrect")
			assert.Equal(t, test.expected_index_x, c.index_x, "Index X incorrect")
			assert.Equal(t, test.expected_status, c.status, "Status is incorrect")
		}
		t.Run(test.name, callback)
	}
}

func TestLoadRom(t *testing.T) {
	c := NewCPU()
	err := c.LoadROM([]uint8{0x1, 0x02, 0x03})

	assert.Equal(t, nil, err)
	assert.Equal(t, uint8(0x01), c.memory[ROM_SEGMENT_START])
	assert.Equal(t, uint8(0x02), c.memory[ROM_SEGMENT_START+1])
	assert.Equal(t, uint8(0x03), c.memory[ROM_SEGMENT_START+2])
}

func TestLoadRom_Overflow(t *testing.T) {
	var memory [0x030000]uint8
	for i := range memory {
		memory[i] = 1
	}

	c := NewCPU()
	err := c.LoadROM(memory[:])

	assert.NotEqual(t, nil, err)
}

func TestLoadAndReset(t *testing.T) {
	c := NewCPU()
	err := c.LoadAndReset([]uint8{0x1, 0x02, 0x03})

	assert.Equal(t, nil, err)
	assert.Equal(t, ROM_SEGMENT_START, c.program_counter)
	assert.Equal(t, uint8(0x01), c.memory[ROM_SEGMENT_START])
	assert.Equal(t, uint8(0x02), c.memory[ROM_SEGMENT_START+1])
	assert.Equal(t, uint8(0x03), c.memory[ROM_SEGMENT_START+2])
}

func TestLoadAndReset_Overflow(t *testing.T) {
	var memory [0x030000]uint8
	for i := range memory {
		memory[i] = 1
	}

	c := NewCPU()
	err := c.LoadAndReset(memory[:])

	assert.NotEqual(t, nil, err)
}

func TestRun_LDA(t *testing.T) {
	testCases := []testInput{
		mkImmediate("Positive value immediate", 0xa9, 0x7b, 0x00, 0x7b, ZERO_BIT),
		mkImmediate("Negative value immediate", 0xa9, 0xfb, 0x00, 0xfb, N_BIT_STATUS),
		mkImmediate("Zero value immediate", 0xa9, 0x00, 0x01, 0x00, Z_BIT_STATUS),
		mkAbsolute("Positive value absolute", 0xad, 0x7b, 0x00, 0x7b, ZERO_BIT),
		mkAbsolute("Negative value absolute", 0xad, 0xfb, 0x00, 0xfb, N_BIT_STATUS),
		mkAbsolute("Zero value absolute", 0xad, 0x00, 0x01, 0x00, Z_BIT_STATUS),
		mkAbsoluteX("Positive value absolute X", 0xbd, 0x7b, 0x00, 0x7b, ZERO_BIT),
		mkAbsoluteX("Negative value absolute X", 0xbd, 0xfb, 0x00, 0xfb, N_BIT_STATUS),
		mkAbsoluteX("Zero value absolute X", 0xbd, 0x00, 0x01, 0x00, Z_BIT_STATUS),
		mkAbsoluteY("Positive value absolute Y", 0xb9, 0x7b, 0x00, 0x7b, ZERO_BIT),
		mkAbsoluteY("Negative value absolute Y", 0xb9, 0xfb, 0x00, 0xfb, N_BIT_STATUS),
		mkAbsoluteY("Zero value absolute Y", 0xb9, 0x00, 0x01, 0x00, Z_BIT_STATUS),
		mkZeroPage("Positive value zero-page", 0xa5, 0x7b, 0x00, 0x7b, ZERO_BIT),
		mkZeroPage("Negative value zero-page", 0xa5, 0xfb, 0x00, 0xfb, N_BIT_STATUS),
		mkZeroPage("Zero value zero-page", 0xa5, 0x00, 0x01, 0x00, Z_BIT_STATUS),
		mkZeroPageX("Positive value zero-page X", 0xb5, 0x7b, 0x00, 0x7b, ZERO_BIT),
		mkZeroPageX("Negative value zero-page X", 0xb5, 0xfb, 0x00, 0xfb, N_BIT_STATUS),
		mkZeroPageX("Zero value zero-page X", 0xb5, 0x00, 0x01, 0x00, Z_BIT_STATUS),
		{
			name:                 "Positive value zero-page X, wrap-around",
			memory:               []uint8{0x00, 0x7b},
			rom:                  []uint8{0xb5, 0xfe},
			initial_index_x:      0x03,
			expected_accumulator: 0x7b,
			expected_status:      ZERO_BIT,
		},
		mkIndirectX("Positive value, indirect X", 0xa1, 0x7b, 0x00, 0x7b, ZERO_BIT),
		mkIndirectX("Negative value, indirect X", 0xa1, 0xfb, 0x00, 0xfb, N_BIT_STATUS),
		mkIndirectX("Zero value, indirect X", 0xa1, 0x00, 0x01, 0x00, Z_BIT_STATUS),
		mkIndirectY("Positive value, indirect Y", 0xb1, 0x7b, 0x00, 0x7b, ZERO_BIT),
		mkIndirectY("Negative value, indirect Y", 0xb1, 0xfb, 0x00, 0xfb, N_BIT_STATUS),
		mkIndirectY("Zero value, indirect Y", 0xb1, 0x00, 0x01, 0x00, Z_BIT_STATUS),
	}

	for _, test := range testCases {
		callback := func(t *testing.T) {
			c := NewCPU()
			c.LoadAndReset(test.rom)

			initializeCpuState(c, test)

			result := c.Run()

			assert.Nil(t, result, "Error was not nil")
			assert.Equal(t, test.expected_accumulator, c.accumulator, "Accumulator incorrect")
			assert.Equal(t, test.expected_status, c.status, "Status incorrect")
		}
		t.Run(test.name, callback)
	}
}

func TestRun_TAX(t *testing.T) {
	testCases := []testInput{
		{name: "Positive accumulator", initial_accumulator: 0x77, initial_status: N_BIT_STATUS | Z_BIT_STATUS, expected_index_x: 0x77, expected_status: ZERO_BIT},
		{name: "Negative accumulator", initial_accumulator: 0xa2, initial_status: Z_BIT_STATUS, expected_index_x: 0xa2, expected_status: N_BIT_STATUS},
		{name: "Zero accumulator", initial_accumulator: 0x00, initial_status: N_BIT_STATUS, expected_index_x: 0x00, expected_status: Z_BIT_STATUS},
	}

	for _, test := range testCases {
		callback := func(t *testing.T) {
			c := NewCPU()
			c.LoadAndReset([]uint8{0xaa})
			initializeCpuState(c, test)
			c.Run()

			assert.Equal(t, test.expected_index_x, c.index_x, "Index X register incorrect")
			assert.Equal(t, test.expected_status, c.status, "Status incorrect")
		}
		t.Run(test.name, callback)
	}
}

func TestRun_INX_Implied(t *testing.T) {
	testCases := []testInput{
		{name: "Positive value", initial_index_x: 0x77, initial_status: N_BIT_STATUS | Z_BIT_STATUS, expected_index_x: 0x78, expected_status: ZERO_BIT},
		{name: "Positive to negative", initial_index_x: 0x7f, initial_status: ZERO_BIT, expected_index_x: 0x80, expected_status: N_BIT_STATUS},
		{name: "Negative value", initial_index_x: 0xa2, initial_status: N_BIT_STATUS, expected_index_x: 0xa3, expected_status: N_BIT_STATUS},
		{name: "Zero value", initial_index_x: 0x00, initial_status: Z_BIT_STATUS, expected_index_x: 0x01, expected_status: ZERO_BIT},
		{name: "Overflow", initial_index_x: 0xff, initial_status: N_BIT_STATUS, expected_index_x: 0x00, expected_status: Z_BIT_STATUS},
	}

	for _, test := range testCases {
		callback := func(t *testing.T) {
			c := NewCPU()
			c.LoadAndReset([]uint8{0xe8})
			initializeCpuState(c, test)
			c.Run()

			assert.Equal(t, test.expected_index_x, c.index_x, "Index X register incorrect")
			assert.Equal(t, test.expected_status, c.status, "Status is incorrect")
		}
		t.Run(test.name, callback)
	}
}

func TestRun_INY_Implied(t *testing.T) {
	testCases := []testInput{
		{name: "Positive value", initial_index_y: 0x77, initial_status: N_BIT_STATUS | Z_BIT_STATUS, expected_index_y: 0x78, expected_status: ZERO_BIT},
		{name: "Positive to negative", initial_index_y: 0x7f, initial_status: ZERO_BIT, expected_index_y: 0x80, expected_status: N_BIT_STATUS},
		{name: "Negative value", initial_index_y: 0xa2, initial_status: N_BIT_STATUS, expected_index_y: 0xa3, expected_status: N_BIT_STATUS},
		{name: "Zero value", initial_index_y: 0x00, initial_status: Z_BIT_STATUS, expected_index_y: 0x01, expected_status: ZERO_BIT},
		{name: "Overflow", initial_index_y: 0xff, initial_status: N_BIT_STATUS, expected_index_y: 0x00, expected_status: Z_BIT_STATUS},
	}

	for _, test := range testCases {
		callback := func(t *testing.T) {
			c := NewCPU()
			c.LoadAndReset([]uint8{0xc8})
			initializeCpuState(c, test)
			c.Run()

			assert.Equal(t, test.expected_index_y, c.index_y, "Index Y register incorrect")
			assert.Equal(t, test.expected_status, c.status, "Status is incorrect")
		}
		t.Run(test.name, callback)
	}
}

func TestRun_ADC(t *testing.T) {
	testCases := []testInput{
		mkImmediate("Immediate, positive, no carry", 0x69, 0x02, 0x03, 0x05, ZERO_BIT),
		mkImmediate("Immediate, positive, carry", 0x69, 0xff, 0x03, 0x02, C_BIT_STATUS),
		mkImmediate("Immediate, negative, no carry", 0x69, 0x7f, 0x03, 0x82, N_BIT_STATUS),
		mkImmediate("Immediate, negative, carry", 0x69, 0xff, 0xff, 0xfe, C_BIT_STATUS|N_BIT_STATUS),
		mkImmediate("Immediate, zero, no carry", 0x69, 0x00, 0x00, 0x00, Z_BIT_STATUS),
		mkImmediate("Immediate, zero, carry", 0x69, 0xff, 0x01, 0x00, Z_BIT_STATUS|C_BIT_STATUS),
		mkZeroPage("Zero-page, positive no carry", 0x65, 0x7b, 0x03, 0x7e, ZERO_BIT),
		mkZeroPage("Zero-page, positive, carry", 0x65, 0xf0, 0x13, 0x03, C_BIT_STATUS),
		mkZeroPage("Zero-page, negative, no carry", 0x65, 0x70, 0x13, 0x83, N_BIT_STATUS),
		mkZeroPage("Zero-page, negative, carry", 0x65, 0xf1, 0xff, 0xf0, C_BIT_STATUS|N_BIT_STATUS),
		mkZeroPage("Zero-page, zero, no carry", 0x65, 0x00, 0x00, 0x00, Z_BIT_STATUS),
		mkZeroPage("Zero-page, zero, carry", 0x65, 0x01, 0xff, 0x00, C_BIT_STATUS|Z_BIT_STATUS),
		mkZeroPageX("Zero-page X, positive no carry", 0x75, 0x7b, 0x03, 0x7e, ZERO_BIT),
		mkZeroPageX("Zero-page X, positive, carry", 0x75, 0xf0, 0x13, 0x03, C_BIT_STATUS),
		mkZeroPageX("Zero-page X, negative, no carry", 0x75, 0x70, 0x13, 0x83, N_BIT_STATUS),
		mkZeroPageX("Zero-page X, negative, carry", 0x75, 0xf1, 0xff, 0xf0, C_BIT_STATUS|N_BIT_STATUS),
		mkZeroPageX("Zero-page X wrap-around, positive, no carry", 0x75, 0x7b, 0x03, 0x7e, ZERO_BIT),
		mkZeroPageX("Zero-page X wrap-around, negative, no carry", 0x75, 0x7b, 0x06, 0x81, N_BIT_STATUS),
		mkZeroPageX("Zero-page X wrap-around, positive, carry", 0x75, 0xeb, 0x16, 0x01, C_BIT_STATUS),
		mkZeroPageX("Zero-page X wrap-around, negative, carry", 0x75, 0x95, 0xf5, 0x8a, N_BIT_STATUS|C_BIT_STATUS),
		mkAbsolute("Absolute, positive, no carry", 0x6d, 0x7b, 0x03, 0x7e, ZERO_BIT),
		mkAbsolute("Absolute, negative, no carry", 0x6d, 0x8b, 0x03, 0x8e, N_BIT_STATUS),
		mkAbsolute("Absolute, positive, carry", 0x6d, 0xff, 0x03, 0x02, C_BIT_STATUS),
		mkAbsolute("Absolute, negative, carry", 0x6d, 0x95, 0xf5, 0x8a, N_BIT_STATUS|C_BIT_STATUS),
		mkAbsolute("Absolute, zero, no carry", 0x6d, 0x00, 0x00, 0x00, Z_BIT_STATUS),
		mkAbsolute("Absolute, zero, carry", 0x6d, 0xfd, 0x03, 0x00, Z_BIT_STATUS|C_BIT_STATUS),
		mkAbsoluteX("Absolute X, positive, no carry", 0x7d, 0x7b, 0x03, 0x7e, ZERO_BIT),
		mkAbsoluteX("Absolute X, negative, no carry", 0x7d, 0x7b, 0x06, 0x81, N_BIT_STATUS),
		mkAbsoluteX("Absolute X, positive, carry", 0x7d, 0xfe, 0x03, 0x01, C_BIT_STATUS),
		mkAbsoluteX("Absolute X, negative, carry", 0x7d, 0xff, 0x82, 0x81, N_BIT_STATUS|C_BIT_STATUS),
		mkAbsoluteX("Absolute X, zero, no carry", 0x7d, 0x00, 0x00, 0x00, Z_BIT_STATUS),
		mkAbsoluteX("Absolute X, zero, carry", 0x7d, 0xff, 0x01, 0x00, Z_BIT_STATUS|C_BIT_STATUS),
		mkAbsoluteY("Absolute Y, positive, no carry", 0x79, 0x7b, 0x03, 0x7e, ZERO_BIT),
		mkAbsoluteY("Absolute Y, negative, no carry", 0x79, 0x7b, 0x06, 0x81, N_BIT_STATUS),
		mkAbsoluteY("Absolute Y, positive, carry", 0x79, 0xfe, 0x03, 0x01, C_BIT_STATUS),
		mkAbsoluteY("Absolute Y, negative, carry", 0x79, 0xff, 0x82, 0x81, N_BIT_STATUS|C_BIT_STATUS),
		mkAbsoluteY("Absolute Y, zero, no carry", 0x79, 0x00, 0x00, 0x00, Z_BIT_STATUS),
		mkAbsoluteY("Absolute Y, zero, carry", 0x79, 0xff, 0x01, 0x00, Z_BIT_STATUS|C_BIT_STATUS),
		mkIndirectX("Indirect X, positive, no carry", 0x61, 0x7b, 0x02, 0x7d, ZERO_BIT),
		mkIndirectXWrapAround("Indirect X, positive, no carry, wrap-around", 0x61, 0x7b, 0x02, 0x7d, ZERO_BIT),
		mkIndirectX("Indirect X, positive, carry", 0x61, 0xfb, 0x12, 0x0d, C_BIT_STATUS),
		mkIndirectXWrapAround("Indirect X, positive, carry, wrap-around", 0x61, 0xfb, 0x12, 0x0d, C_BIT_STATUS),
		mkIndirectX("Indirect X, negative, no carry", 0x61, 0x8b, 0x02, 0x8d, N_BIT_STATUS),
		mkIndirectXWrapAround("Indirect X, negative, no carry, wrap-around", 0x61, 0x8b, 0x02, 0x8d, N_BIT_STATUS),
		mkIndirectX("Indirect X, negative, carry", 0x61, 0xfb, 0x92, 0x8d, C_BIT_STATUS|N_BIT_STATUS),
		mkIndirectXWrapAround("Indirect X, negative, carry, wrap-around", 0x61, 0xfb, 0x92, 0x8d, C_BIT_STATUS|N_BIT_STATUS),
		mkIndirectX("Indirect X, zero, carry", 0x61, 0xfb, 0x05, 0x00, C_BIT_STATUS|Z_BIT_STATUS),
		mkIndirectXWrapAround("Indirect X, zero, carry, wrap-around", 0x61, 0xfb, 0x05, 0x00, C_BIT_STATUS|Z_BIT_STATUS),
		mkIndirectY("Indirect Y, positive, no carry", 0x71, 0x7b, 0x2, 0x7d, ZERO_BIT),
		mkIndirectY("Indirect Y, positive, carry", 0x71, 0xfb, 0x07, 0x02, C_BIT_STATUS),
		mkIndirectY("Indirect Y, negative, no carry", 0x71, 0x8b, 0x02, 0x8d, N_BIT_STATUS),
		mkIndirectY("Indirect Y, negative, carry", 0x71, 0xfb, 0x92, 0x8d, C_BIT_STATUS|N_BIT_STATUS),
		mkIndirectY("Indirect Y, zero, no carry", 0x71, 0x00, 0x00, 0x00, Z_BIT_STATUS),
		mkIndirectY("Indirect Y, zero, carry", 0x71, 0xfb, 0x05, 0x00, C_BIT_STATUS|Z_BIT_STATUS),
	}

	for _, test := range testCases {
		callback := func(t *testing.T) {
			c := NewCPU()
			c.LoadAndReset(test.rom)

			initializeCpuState(c, test)

			result := c.Run()

			assert.Nil(t, result, "Error was not nil")
			assert.Equal(t, test.expected_accumulator, c.accumulator, "Accumulator incorrect")
			assert.Equal(t, test.expected_status, c.status, "Status incorrect")
		}
		t.Run(test.name, callback)
	}
}

func TestRun_AND(t *testing.T) {
	testCases := []testInput{
		mkImmediate("Immediate, positive", 0x29, 0x02, 0x03, 0x02, ZERO_BIT),
		mkImmediate("Immediate, negative", 0x29, 0x8f, 0x83, 0x83, N_BIT_STATUS),
		mkImmediate("Immediate, zero", 0x29, 0x02, 0x01, 0x00, Z_BIT_STATUS),
		mkZeroPage("Zero-page, positive", 0x25, 0x7b, 0x03, 0x03, ZERO_BIT),
		mkZeroPage("Zero-page, negative", 0x25, 0x82, 0x93, 0x82, N_BIT_STATUS),
		mkZeroPage("Zero-page, zero", 0x25, 0x30, 0x01, 0x00, Z_BIT_STATUS),
		mkZeroPageX("Zero-page X, positive", 0x35, 0x7b, 0x03, 0x03, ZERO_BIT),
		mkZeroPageX("Zero-page X, negative", 0x35, 0x82, 0x93, 0x82, N_BIT_STATUS),
		mkZeroPageX("Zero-page X, zero", 0x35, 0x30, 0x01, 0x00, Z_BIT_STATUS),
		mkAbsolute("Absolute, positive", 0x2d, 0x7b, 0x03, 0x03, ZERO_BIT),
		mkAbsolute("Absolute, negative", 0x2d, 0x82, 0x93, 0x82, N_BIT_STATUS),
		mkAbsolute("Absolute, zero", 0x2d, 0x30, 0x01, 0x00, Z_BIT_STATUS),
		mkAbsoluteX("Absolute X, positive", 0x3d, 0x7b, 0x03, 0x03, ZERO_BIT),
		mkAbsoluteX("Absolute X, negative", 0x3d, 0x82, 0x93, 0x82, N_BIT_STATUS),
		mkAbsoluteX("Absolute X, zero", 0x3d, 0x30, 0x01, 0x00, Z_BIT_STATUS),
		mkAbsoluteY("Absolute Y, positive", 0x39, 0x7b, 0x03, 0x03, ZERO_BIT),
		mkAbsoluteY("Absolute Y, negative", 0x39, 0x82, 0x93, 0x82, N_BIT_STATUS),
		mkAbsoluteY("Absolute Y, zero", 0x39, 0x30, 0x01, 0x00, Z_BIT_STATUS),
		mkIndirectX("Indirect X, positive", 0x21, 0x7b, 0x03, 0x03, ZERO_BIT),
		mkIndirectX("Indirect X, negative", 0x21, 0x82, 0x93, 0x82, N_BIT_STATUS),
		mkIndirectX("Indirect X, zero", 0x21, 0x30, 0x01, 0x00, Z_BIT_STATUS),
		mkIndirectXWrapAround("Indirect X, wrap-around, positive", 0x21, 0x7b, 0x03, 0x03, ZERO_BIT),
		mkIndirectXWrapAround("Indirect X, wrap-around, negative", 0x21, 0x82, 0x93, 0x82, N_BIT_STATUS),
		mkIndirectXWrapAround("Indirect X, wrap-around, zero", 0x21, 0x30, 0x01, 0x00, Z_BIT_STATUS),
		mkIndirectY("Indirect Y, positive", 0x31, 0x7b, 0x03, 0x03, ZERO_BIT),
		mkIndirectY("Indirect Y, negative", 0x31, 0x82, 0x93, 0x82, N_BIT_STATUS),
		mkIndirectY("Indirect Y, zero", 0x31, 0x30, 0x01, 0x00, Z_BIT_STATUS),
	}

	for _, test := range testCases {
		callback := func(t *testing.T) {
			c := NewCPU()
			c.LoadAndReset(test.rom)

			initializeCpuState(c, test)

			result := c.Run()

			assert.Nil(t, result, "Error was not nil")
			assert.Equal(t, test.expected_accumulator, c.accumulator, "Accumulator incorrect")
			assert.Equal(t, test.expected_status, c.status, "Status incorrect")
		}
		t.Run(test.name, callback)
	}
}

func TestRun_ASL(t *testing.T) {
	runMemoryTests := func(testCases []testInput, location uint16) {
		for _, test := range testCases {
			callback := func(t *testing.T) {
				c := NewCPU()
				c.LoadAndReset(test.rom)

				initializeCpuState(c, test)

				result := c.Run()

				assert.Nil(t, result, "Error was not nil")
				assert.Equal(t, test.expected_accumulator, c.memory[location], "Memory value incorrect")
				assert.Equal(t, test.expected_status, c.status, "Status incorrect")
			}
			t.Run(test.name, callback)
		}
	}

	testCases := []testInput{
		mkAccumulator("Accumulator, positive", 0x0a, 0x03, 0x06, ZERO_BIT),
		mkAccumulator("Accumulator, positive, carry", 0x0a, 0x83, 0x06, C_BIT_STATUS),
		mkAccumulator("Accumulator, negative", 0x0a, 0x47, 0x8e, N_BIT_STATUS),
		mkAccumulator("Accumulator, negative, carry", 0x0a, 0xc6, 0x8c, N_BIT_STATUS|C_BIT_STATUS),
		mkAccumulator("Accumulator, zero", 0x0a, 0x00, 0x00, Z_BIT_STATUS),
		mkAccumulator("Accumulator, zero, carry", 0x0a, 0x80, 0x00, Z_BIT_STATUS|C_BIT_STATUS),
	}

	for _, test := range testCases {
		callback := func(t *testing.T) {
			c := NewCPU()
			c.LoadAndReset(test.rom)

			initializeCpuState(c, test)

			result := c.Run()

			assert.Nil(t, result, "Error was not nil")
			assert.Equal(t, test.expected_accumulator, c.accumulator, "Accumulator incorrect")
			assert.Equal(t, test.expected_status, c.status, "Status incorrect")
		}
		t.Run(test.name, callback)
	}

	// For this, initial_accumulator and expected_accumulator do double-duty assert
	// the memory location and expected value resectively.
	memoryTestCases := []testInput{
		mkZeroPage("Zero-page, positive", 0x06, 0x06, ZERO_BIT, 0x0c, ZERO_BIT),
		mkZeroPage("Zero-page, positive, carry", 0x06, 0x84, ZERO_BIT, 0x08, C_BIT_STATUS),
		mkZeroPage("Zero-page, negative", 0x06, 0x66, ZERO_BIT, 0xcc, N_BIT_STATUS),
		mkZeroPage("Zero-page, negative, carry", 0x06, 0xc2, ZERO_BIT, 0x84, N_BIT_STATUS|C_BIT_STATUS),
		mkZeroPage("Zero-page, zero", 0x06, 0x00, ZERO_BIT, 0x00, Z_BIT_STATUS),
		mkZeroPage("Zero-page, zero, carry", 0x06, 0x80, ZERO_BIT, 0x00, Z_BIT_STATUS|C_BIT_STATUS),
	}

	runMemoryTests(memoryTestCases, uint16(0x03))

	zeroPageXTests := []testInput{
		mkZeroPageX("Zero-page X, positive", 0x16, 0x06, ZERO_BIT, 0x0c, ZERO_BIT),
		mkZeroPageX("Zero-page X, positive, carry", 0x16, 0x84, ZERO_BIT, 0x08, C_BIT_STATUS),
		mkZeroPageX("Zero-page X, negative", 0x16, 0x66, ZERO_BIT, 0xcc, N_BIT_STATUS),
		mkZeroPageX("Zero-page X, negative, carry", 0x16, 0xc2, ZERO_BIT, 0x84, N_BIT_STATUS|C_BIT_STATUS),
		mkZeroPageX("Zero-page X, zero", 0x16, 0x00, ZERO_BIT, 0x00, Z_BIT_STATUS),
		mkZeroPageX("Zero-page X, zero, carry", 0x16, 0x80, ZERO_BIT, 0x00, Z_BIT_STATUS|C_BIT_STATUS),
	}

	runMemoryTests(zeroPageXTests, uint16(0x03))

	absoluteTests := []testInput{
		mkAbsolute("Absolute, positive", 0x0e, 0x06, ZERO_BIT, 0x0c, ZERO_BIT),
		mkAbsolute("Absolute, positive, carry", 0x0e, 0x84, ZERO_BIT, 0x08, C_BIT_STATUS),
		mkAbsolute("Absolute, negative", 0x0e, 0x66, ZERO_BIT, 0xcc, N_BIT_STATUS),
		mkAbsolute("Absolute, negative, carry", 0x0e, 0xc2, ZERO_BIT, 0x84, N_BIT_STATUS|C_BIT_STATUS),
		mkAbsolute("Absolute, zero", 0x0e, 0x00, ZERO_BIT, 0x00, Z_BIT_STATUS),
		mkAbsolute("Absolute, zero, carry", 0x0e, 0x80, ZERO_BIT, 0x00, Z_BIT_STATUS|C_BIT_STATUS),
	}

	runMemoryTests(absoluteTests, uint16(0x1001))

	absoluteXTests := []testInput{
		mkAbsoluteX("Absolute X, positive", 0x1e, 0x06, ZERO_BIT, 0x0c, ZERO_BIT),
		mkAbsoluteX("Absolute X, positive, carry", 0x1e, 0x84, ZERO_BIT, 0x08, C_BIT_STATUS),
		mkAbsoluteX("Absolute X, negative", 0x1e, 0x66, ZERO_BIT, 0xcc, N_BIT_STATUS),
		mkAbsoluteX("Absolute X, negative, carry", 0x1e, 0xc2, ZERO_BIT, 0x84, N_BIT_STATUS|C_BIT_STATUS),
		mkAbsoluteX("Absolute X, zero", 0x1e, 0x00, ZERO_BIT, 0x00, Z_BIT_STATUS),
		mkAbsoluteX("Absolute X, zero, carry", 0x1e, 0x80, ZERO_BIT, 0x00, Z_BIT_STATUS|C_BIT_STATUS),
	}

	runMemoryTests(absoluteXTests, uint16(0x1003))
}

func TestRun_BCC_CarrySet(t *testing.T) {
	rom := []uint8{0x90, 0x07}
	c := NewCPU()
	c.LoadAndReset(rom)
	c.status = C_BIT_STATUS

	result := c.Run()

	assert.Nil(t, result, "Error was not nil")
	// Start at 0x8000, 2 bytes for BCC, 1 byte for BRK
	assert.Equal(t, uint16(0x8003), c.program_counter, "Program counter incorrect")
}

func TestRun_BCC_CarryClear(t *testing.T) {
	rom := []uint8{0x90, 0x07}
	c := NewCPU()
	c.LoadAndReset(rom)
	c.status = ZERO_BIT

	result := c.Run()

	assert.Nil(t, result, "Error was not nil")
	assert.Equal(t, uint16(0x8009), c.program_counter, "Program counter incorrect")
}

func TestRun_BCS_CarrySet(t *testing.T) {
	rom := []uint8{0xb0, 0x07}
	c := NewCPU()
	c.LoadAndReset(rom)
	c.status = C_BIT_STATUS

	result := c.Run()

	assert.Nil(t, result, "Error was not nil")
	assert.Equal(t, uint16(0x8009), c.program_counter, "Program counter incorrect")
}

func TestRun_BCS_CarryClear(t *testing.T) {
	rom := []uint8{0xb0, 0x07}
	c := NewCPU()
	c.LoadAndReset(rom)
	c.status = ZERO_BIT

	result := c.Run()

	assert.Nil(t, result, "Error was not nil")
	// Start at 0x8000, 2 bytes for BCC, 1 byte for BRK
	assert.Equal(t, uint16(0x8003), c.program_counter, "Program counter incorrect")
}

func setCommonFields(test testInput, name string, initial, expected, status uint8) testInput {
	test.name = name
	test.initial_accumulator = initial
	test.expected_accumulator = expected
	test.expected_status = status
	return test
}

func mkAccumulator(name string, opcode, initial, expected, status uint8) testInput {
	test := testInput{
		rom: []uint8{opcode},
	}
	return setCommonFields(test, name, initial, expected, status)
}

func mkImmediate(name string, opcode, param, initial, expected, status uint8) testInput {
	test := testInput{
		rom: []uint8{opcode, param},
	}
	return setCommonFields(test, name, initial, expected, status)
}

func mkZeroPage(name string, opcode, param, initial, expected, status uint8) testInput {
	test := testInput{
		memory: []uint8{0x00, 0x00, 0x00, param},
		rom:    []uint8{opcode, 0x03},
	}
	return setCommonFields(test, name, initial, expected, status)
}

func mkZeroPageX(name string, opcode, param, initial, expected, status uint8) testInput {
	test := testInput{
		memory:          []uint8{0x00, 0x00, 0x00, param},
		rom:             []uint8{opcode, 0x01},
		initial_index_x: 0x02,
	}
	return setCommonFields(test, name, initial, expected, status)
}

func mkAbsolute(name string, opcode, param, initial, expected, status uint8) testInput {
	test := testInput{
		upper_memory: []uint8{0x00, param},
		rom:          []uint8{opcode, 0x01, 0x10},
	}
	return setCommonFields(test, name, initial, expected, status)
}

func mkAbsoluteX(name string, opcode, param, initial, expected, status uint8) testInput {
	test := testInput{
		upper_memory:    []uint8{0x00, 0x00, 0x00, param},
		rom:             []uint8{opcode, 0x01, 0x10},
		initial_index_x: 0x02,
	}
	return setCommonFields(test, name, initial, expected, status)
}

func mkAbsoluteY(name string, opcode, param, initial, expected, status uint8) testInput {
	test := testInput{
		upper_memory:    []uint8{0x00, 0x00, 0x00, param},
		rom:             []uint8{opcode, 0x01, 0x10},
		initial_index_y: 0x02,
	}
	return setCommonFields(test, name, initial, expected, status)
}

func mkIndirectX(name string, opcode, value, initial, expected, status uint8) testInput {
	test := testInput{
		rom:             []uint8{opcode, 0x01},
		memory:          []uint8{0x00, 0x00, 0x01, 0x10},
		upper_memory:    []uint8{0x00, value},
		initial_index_x: 0x01,
	}
	return setCommonFields(test, name, initial, expected, status)
}

func mkIndirectXWrapAround(name string, opcode, value, initial, expected, status uint8) testInput {
	test := testInput{
		rom:             []uint8{opcode, 0xff},
		memory:          []uint8{0x00, 0x00, 0x01, 0x10},
		upper_memory:    []uint8{0x00, value},
		initial_index_x: 0x03,
	}
	return setCommonFields(test, name, initial, expected, status)
}

func mkIndirectY(name string, opcode, value, initial, expected, status uint8) testInput {
	test := testInput{
		rom:             []uint8{opcode, 0x02},
		memory:          []uint8{0x00, 0x00, 0x01, 0x10},
		upper_memory:    []uint8{0x00, 0x00, value},
		initial_index_y: 0x01,
	}
	return setCommonFields(test, name, initial, expected, status)
}
