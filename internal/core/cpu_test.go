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
	setCommonFields := func(test testInput, name string, value, status uint8) testInput {
		test.name = name
		test.expected_accumulator = value
		test.expected_status = status
		return test
	}
	mkImmediate := func(name string, value, status uint8) testInput {
		return setCommonFields(testInput{rom: []uint8{0xa9, value}}, name, value, status)
	}
	mkAbsolute := func(name string, value, status uint8) testInput {
		testCase := testInput{
			rom:          []uint8{0xad, 0x03, 0x10},
			upper_memory: []uint8{0x00, 0x00, 0x00, value},
		}
		return setCommonFields(testCase, name, value, status)
	}
	mkAbsoluteX := func(name string, value, status uint8) testInput {
		testCase := testInput{
			upper_memory:    []uint8{0x00, 0x00, 0x00, value},
			rom:             []uint8{0xbd, 0x01, 0x10},
			initial_index_x: 0x02,
		}
		return setCommonFields(testCase, name, value, status)
	}
	mkAbsoluteY := func(name string, value, status uint8) testInput {
		testCase := testInput{
			upper_memory:    []uint8{0x00, 0x00, 0x00, value},
			rom:             []uint8{0xb9, 0x01, 0x10},
			initial_index_y: 0x02,
		}
		return setCommonFields(testCase, name, value, status)
	}
	mkZeroPage := func(name string, value, status uint8) testInput {
		testCase := testInput{
			memory: []uint8{0x00, 0x00, 0x00, value},
			rom:    []uint8{0xa5, 0x03},
		}
		return setCommonFields(testCase, name, value, status)
	}
	mkZeroPageX := func(name string, value, status uint8) testInput {
		testCase := testInput{
			memory:          []uint8{0x00, 0x00, 0x00, value},
			rom:             []uint8{0xb5, 0x02},
			initial_index_x: 0x01,
		}
		return setCommonFields(testCase, name, value, status)
	}
	mkIndirectX := func(name string, value, status uint8) testInput {
		testCase := testInput{
			initial_index_x: 0x01,
			rom:             []uint8{0xa1, 0x01},
			memory:          []uint8{0x00, 0x00, 0x01, 0x10},
			upper_memory:    []uint8{0x00, value},
		}
		return setCommonFields(testCase, name, value, status)
	}
	mkIndirectY := func(name string, value, status uint8) testInput {
		testCase := testInput{
			initial_index_y: 0x01,
			rom:             []uint8{0xb1, 0x01},
			memory:          []uint8{0x00, 0x01, 0x10},
			upper_memory:    []uint8{0x00, 0x00, value},
		}
		return setCommonFields(testCase, name, value, status)
	}
	testCases := []testInput{
		mkImmediate("Positive value immediate", 0x7b, ZERO_BIT),
		mkImmediate("Negative value immediate", 0xfb, N_BIT_STATUS),
		mkImmediate("Zero value immediate", 0x00, Z_BIT_STATUS),
		mkAbsolute("Positive value absolute", 0x7b, ZERO_BIT),
		mkAbsolute("Negative value absolute", 0xfb, N_BIT_STATUS),
		mkAbsolute("Zero value absolute", 0x00, Z_BIT_STATUS),
		mkAbsoluteX("Positive value absolute X", 0x7b, ZERO_BIT),
		mkAbsoluteX("Negative value absolute X", 0xfb, N_BIT_STATUS),
		mkAbsoluteX("Zero value absolute X", 0x00, Z_BIT_STATUS),
		mkAbsoluteY("Positive value absolute Y", 0x7b, ZERO_BIT),
		mkAbsoluteY("Negative value absolute Y", 0xfb, N_BIT_STATUS),
		mkAbsoluteY("Zero value absolute Y", 0x00, Z_BIT_STATUS),
		mkZeroPage("Positive value zero-page", 0x7b, ZERO_BIT),
		mkZeroPage("Negative value zero-page", 0xfb, N_BIT_STATUS),
		mkZeroPage("Zero value zero-page", 0x00, Z_BIT_STATUS),
		mkZeroPageX("Positive value zero-page X", 0x7b, ZERO_BIT),
		mkZeroPageX("Negative value zero-page X", 0xfb, N_BIT_STATUS),
		mkZeroPageX("Zero value zero-page X", 0x00, Z_BIT_STATUS),
		{
			name:                 "Positive value zero-page X, wrap-around",
			memory:               []uint8{0x00, 0x7b},
			rom:                  []uint8{0xb5, 0xfe},
			initial_index_x:      0x03,
			expected_accumulator: 0x7b,
			expected_status:      ZERO_BIT,
		},
		mkIndirectX("Positive value, indirect X", 0x7b, ZERO_BIT),
		mkIndirectX("Negative value, indirect X", 0xfb, N_BIT_STATUS),
		mkIndirectX("Zero value, indirect X", 0x00, Z_BIT_STATUS),
		mkIndirectY("Positive value, indirect Y", 0x7b, ZERO_BIT),
		mkIndirectY("Negative value, indirect Y", 0xfb, N_BIT_STATUS),
		mkIndirectY("Zero value, indirect Y", 0x00, Z_BIT_STATUS),
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
	setCommonFields := func(test testInput, name string, initial, expected, status uint8) testInput {
		test.name = name
		test.initial_accumulator = initial
		test.expected_accumulator = expected
		test.expected_status = status
		return test
	}
	mkImmediate := func(name string, param, initial, expected, status uint8) testInput {
		test := testInput{
			rom: []uint8{0x69, param},
		}
		return setCommonFields(test, name, initial, expected, status)
	}
	mkZeroPage := func(name string, param, initial, expected, status uint8) testInput {
		test := testInput{
			memory: []uint8{0x00, 0x00, 0x00, param},
			rom:    []uint8{0x65, 0x03},
		}
		return setCommonFields(test, name, initial, expected, status)
	}
	mkZeroPageX := func(name string, param, initial, expected, status uint8) testInput {
		test := testInput{
			memory:          []uint8{0x00, 0x00, 0x00, param},
			rom:             []uint8{0x75, 0x01},
			initial_index_x: 0x02,
		}
		return setCommonFields(test, name, initial, expected, status)
	}
	mkAbsolute := func(name string, param, initial, expected, status uint8) testInput {
		test := testInput{
			upper_memory: []uint8{0x00, param},
			rom:          []uint8{0x6d, 0x01, 0x10},
		}
		return setCommonFields(test, name, initial, expected, status)
	}
	mkAbsoluteX := func(name string, param, initial, expected, status uint8) testInput {
		test := testInput{
			upper_memory:    []uint8{0x00, 0x00, 0x00, param},
			rom:             []uint8{0x7d, 0x01, 0x10},
			initial_index_x: 0x02,
		}
		return setCommonFields(test, name, initial, expected, status)
	}
	mkAbsoluteY := func(name string, param, initial, expected, status uint8) testInput {
		test := testInput{
			upper_memory:    []uint8{0x00, 0x00, 0x00, param},
			rom:             []uint8{0x79, 0x01, 0x10},
			initial_index_y: 0x02,
		}
		return setCommonFields(test, name, initial, expected, status)
	}
	mkIndirectX := func(name string, value, initial, expected, status uint8) testInput {
		test := testInput{
			rom:             []uint8{0x61, 0x01},
			memory:          []uint8{0x00, 0x00, 0x01, 0x10},
			upper_memory:    []uint8{0x00, value},
			initial_index_x: 0x01,
		}
		return setCommonFields(test, name, initial, expected, status)
	}
	mkIndirectXWrapAround := func(name string, value, initial, expected, status uint8) testInput {
		test := testInput{
			rom:             []uint8{0x61, 0xff},
			memory:          []uint8{0x00, 0x00, 0x01, 0x10},
			upper_memory:    []uint8{0x00, value},
			initial_index_x: 0x03,
		}
		return setCommonFields(test, name, initial, expected, status)
	}
	mkIndirectY := func(name string, value, initial, expected, status uint8) testInput {
		test := testInput{
			rom:             []uint8{0x71, 0x02},
			memory:          []uint8{0x00, 0x00, 0x01, 0x10},
			upper_memory:    []uint8{0x00, 0x00, value},
			initial_index_y: 0x01,
		}
		return setCommonFields(test, name, initial, expected, status)
	}

	testCases := []testInput{
		mkImmediate("Immediate, positive, no carry", 0x02, 0x03, 0x05, ZERO_BIT),
		mkImmediate("Immediate, positive, carry", 0xff, 0x03, 0x02, C_BIT_STATUS),
		mkImmediate("Immediate, negative, no carry", 0x7f, 0x03, 0x82, N_BIT_STATUS),
		mkImmediate("Immediate, negative, carry", 0xff, 0xff, 0xfe, C_BIT_STATUS|N_BIT_STATUS),
		mkImmediate("Immediate, zero, no carry", 0x00, 0x00, 0x00, Z_BIT_STATUS),
		mkImmediate("Immediate, zero, carry", 0xff, 0x01, 0x00, Z_BIT_STATUS|C_BIT_STATUS),
		mkZeroPage("Zero-page, positive no carry", 0x7b, 0x03, 0x7e, ZERO_BIT),
		mkZeroPage("Zero-page, positive, carry", 0xf0, 0x13, 0x03, C_BIT_STATUS),
		mkZeroPage("Zero-page, negative, no carry", 0x70, 0x13, 0x83, N_BIT_STATUS),
		mkZeroPage("Zero-page, negative, carry", 0xf1, 0xff, 0xf0, C_BIT_STATUS|N_BIT_STATUS),
		mkZeroPage("Zero-page, zero, no carry", 0x00, 0x00, 0x00, Z_BIT_STATUS),
		mkZeroPage("Zero-page, zero, carry", 0x01, 0xff, 0x00, C_BIT_STATUS|Z_BIT_STATUS),
		mkZeroPageX("Zero-page X, positive no carry", 0x7b, 0x03, 0x7e, ZERO_BIT),
		mkZeroPageX("Zero-page X, positive, carry", 0xf0, 0x13, 0x03, C_BIT_STATUS),
		mkZeroPageX("Zero-page X, negative, no carry", 0x70, 0x13, 0x83, N_BIT_STATUS),
		mkZeroPageX("Zero-page X, negative, carry", 0xf1, 0xff, 0xf0, C_BIT_STATUS|N_BIT_STATUS),
		mkZeroPageX("Zero-page X wrap-around, positive, no carry", 0x7b, 0x03, 0x7e, ZERO_BIT),
		mkZeroPageX("Zero-page X wrap-around, negative, no carry", 0x7b, 0x06, 0x81, N_BIT_STATUS),
		mkZeroPageX("Zero-page X wrap-around, positive, carry", 0xeb, 0x16, 0x01, C_BIT_STATUS),
		mkZeroPageX("Zero-page X wrap-around, negative, carry", 0x95, 0xf5, 0x8a, N_BIT_STATUS|C_BIT_STATUS),
		mkAbsolute("Absolute, positive, no carry", 0x7b, 0x03, 0x7e, ZERO_BIT),
		mkAbsolute("Absolute, negative, no carry", 0x8b, 0x03, 0x8e, N_BIT_STATUS),
		mkAbsolute("Absolute, positive, carry", 0xff, 0x03, 0x02, C_BIT_STATUS),
		mkAbsolute("Absolute, negative, carry", 0x95, 0xf5, 0x8a, N_BIT_STATUS|C_BIT_STATUS),
		mkAbsolute("Absolute, zero, no carry", 0x00, 0x00, 0x00, Z_BIT_STATUS),
		mkAbsolute("Absolute, zero, carry", 0xfd, 0x03, 0x00, Z_BIT_STATUS|C_BIT_STATUS),
		mkAbsoluteX("Absolute X, positive, no carry", 0x7b, 0x03, 0x7e, ZERO_BIT),
		mkAbsoluteX("Absolute X, negative, no carry", 0x7b, 0x06, 0x81, N_BIT_STATUS),
		mkAbsoluteX("Absolute X, positive, carry", 0xfe, 0x03, 0x01, C_BIT_STATUS),
		mkAbsoluteX("Absolute X, negative, carry", 0xff, 0x82, 0x81, N_BIT_STATUS|C_BIT_STATUS),
		mkAbsoluteX("Absolute X, zero, no carry", 0x00, 0x00, 0x00, Z_BIT_STATUS),
		mkAbsoluteX("Absolute X, zero, carry", 0xff, 0x01, 0x00, Z_BIT_STATUS|C_BIT_STATUS),
		mkAbsoluteY("Absolute Y, positive, no carry", 0x7b, 0x03, 0x7e, ZERO_BIT),
		mkAbsoluteY("Absolute Y, negative, no carry", 0x7b, 0x06, 0x81, N_BIT_STATUS),
		mkAbsoluteY("Absolute Y, positive, carry", 0xfe, 0x03, 0x01, C_BIT_STATUS),
		mkAbsoluteY("Absolute Y, negative, carry", 0xff, 0x82, 0x81, N_BIT_STATUS|C_BIT_STATUS),
		mkAbsoluteY("Absolute Y, zero, no carry", 0x00, 0x00, 0x00, Z_BIT_STATUS),
		mkAbsoluteY("Absolute Y, zero, carry", 0xff, 0x01, 0x00, Z_BIT_STATUS|C_BIT_STATUS),
		mkIndirectX("Indirect X, positive, no carry", 0x7b, 0x02, 0x7d, ZERO_BIT),
		mkIndirectXWrapAround("Indirect X, positive, no carry, wrap-around", 0x7b, 0x02, 0x7d, ZERO_BIT),
		mkIndirectX("Indirect X, positive, carry", 0xfb, 0x12, 0x0d, C_BIT_STATUS),
		mkIndirectXWrapAround("Indirect X, positive, carry, wrap-around", 0xfb, 0x12, 0x0d, C_BIT_STATUS),
		mkIndirectX("Indirect X, negative, no carry", 0x8b, 0x02, 0x8d, N_BIT_STATUS),
		mkIndirectXWrapAround("Indirect X, negative, no carry, wrap-around", 0x8b, 0x02, 0x8d, N_BIT_STATUS),
		mkIndirectX("Indirect X, negative, carry", 0xfb, 0x92, 0x8d, C_BIT_STATUS|N_BIT_STATUS),
		mkIndirectXWrapAround("Indirect X, negative, carry, wrap-around", 0xfb, 0x92, 0x8d, C_BIT_STATUS|N_BIT_STATUS),
		mkIndirectX("Indirect X, zero, carry", 0xfb, 0x05, 0x00, C_BIT_STATUS|Z_BIT_STATUS),
		mkIndirectXWrapAround("Indirect X, zero, carry, wrap-around", 0xfb, 0x05, 0x00, C_BIT_STATUS|Z_BIT_STATUS),
		mkIndirectY("Indirect Y, positive, no carry", 0x7b, 0x2, 0x7d, ZERO_BIT),
		mkIndirectY("Indirect Y, positive, carry", 0xfb, 0x07, 0x02, C_BIT_STATUS),
		mkIndirectY("Indirect Y, negative, no carry", 0x8b, 0x02, 0x8d, N_BIT_STATUS),
		mkIndirectY("Indirect Y, negative, carry", 0xfb, 0x92, 0x8d, C_BIT_STATUS|N_BIT_STATUS),
		mkIndirectY("Indirect Y, zero, no carry", 0x8b, 0x02, 0x8d, N_BIT_STATUS),
		mkIndirectY("Indirect Y, zero, carry", 0xfb, 0x92, 0x8d, C_BIT_STATUS|N_BIT_STATUS),
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
