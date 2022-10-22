package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

const ZERO_BIT uint8 = 0

func TestRun_UndefinedOpcode(t *testing.T) {
	memory := []uint8{0xff}
	c := NewCPU()
	c.LoadAndReset(memory)
	result := c.Run()
	assert.Equal(t, errors.New("Unimplemented opcode"), result)
}

func TestRun_Scenario(t *testing.T) {
	testCases := []struct {
		name                 string
		memory               []uint8
		expected_accumulator uint8
		expected_index_x     uint8
		expected_n_bit       uint8
		expected_z_bit       uint8
	}{
		{
			name:                 "LDA TAX INX BRK",
			memory:               []uint8{0xa9, 0xc0, 0xaa, 0xe8, 0x00},
			expected_accumulator: 0xc0,
			expected_index_x:     0xc1,
			expected_n_bit:       N_BIT_STATUS,
			expected_z_bit:       ZERO_BIT,
		},
	}

	for _, test := range testCases {
		callback := func(t *testing.T) {
			c := NewCPU()
			c.LoadAndReset(test.memory)
			c.Run()

			assert.Equal(t, test.expected_accumulator, c.accumulator, "Accumulator incorrect")
			assert.Equal(t, test.expected_index_x, c.index_x, "Index X incorrect")
			assert.Equal(t, test.expected_n_bit, c.status&N_BIT_STATUS, "Negative status bit incorrect")
			assert.Equal(t, test.expected_z_bit, c.status&Z_BIT_STATUS, "Zero status bit incorrect")
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
	testCases := []struct {
		name                 string
		memory               []uint8
		upper_memory         []uint8
		rom                  []uint8
		reg_x                uint8
		reg_y                uint8
		expected_accumulator uint8
		expected_status      uint8
	}{
		{
			name:                 "Positive value immediate",
			rom:                  []uint8{0xa9, 0x7b},
			expected_accumulator: 0x7b,
			expected_status:      ZERO_BIT,
		},
		{
			name:                 "Negative value immediate",
			rom:                  []uint8{0xa9, 0xfb},
			expected_accumulator: 0xfb,
			expected_status:      N_BIT_STATUS,
		},
		{
			name:                 "Zero value immediate",
			rom:                  []uint8{0xa9, 0x00},
			expected_accumulator: 0x00,
			expected_status:      Z_BIT_STATUS,
		},
		{
			name:                 "Positive value absolute",
			upper_memory:         []uint8{0x00, 0x00, 0x00, 0x7b},
			rom:                  []uint8{0xad, 0x03, 0x10},
			expected_accumulator: 0x7b,
			expected_status:      ZERO_BIT,
		},
		{
			name:                 "Negative value absolute",
			upper_memory:         []uint8{0x00, 0x00, 0x00, 0xfb},
			rom:                  []uint8{0xad, 0x03, 0x10},
			expected_accumulator: 0xfb,
			expected_status:      N_BIT_STATUS,
		},
		{
			name:                 "Zero value absolute",
			rom:                  []uint8{0xad, 0x03, 0x10},
			expected_accumulator: 0x00,
			expected_status:      Z_BIT_STATUS,
		},
		{
			name:                 "Positive value absolute X",
			upper_memory:         []uint8{0x00, 0x00, 0x00, 0x7b},
			rom:                  []uint8{0xbd, 0x01, 0x10},
			reg_x:                0x02,
			expected_accumulator: 0x7b,
			expected_status:      ZERO_BIT,
		},
		{
			name:                 "Negative value absolute X",
			upper_memory:         []uint8{0x00, 0x00, 0x00, 0xfb},
			rom:                  []uint8{0xbd, 0x01, 0x10},
			reg_x:                0x02,
			expected_accumulator: 0xfb,
			expected_status:      N_BIT_STATUS,
		},
		{
			name:                 "Zero value absolute X",
			rom:                  []uint8{0xbd, 0x01, 0x10},
			reg_x:                0x02,
			expected_accumulator: 0x00,
			expected_status:      Z_BIT_STATUS,
		},
		{
			name:                 "Positive value absolute Y",
			upper_memory:         []uint8{0x00, 0x00, 0x00, 0x7b},
			rom:                  []uint8{0xb9, 0x01, 0x10},
			reg_y:                0x02,
			expected_accumulator: 0x7b,
			expected_status:      ZERO_BIT,
		},
		{
			name:                 "Negative value absolute Y",
			upper_memory:         []uint8{0x00, 0x00, 0x00, 0xfb},
			rom:                  []uint8{0xb9, 0x01, 0x10},
			reg_y:                0x02,
			expected_accumulator: 0xfb,
			expected_status:      N_BIT_STATUS,
		},
		{
			name:                 "Zero value absolute Y",
			rom:                  []uint8{0xb9, 0x01, 0x10},
			reg_y:                0x02,
			expected_accumulator: 0x00,
			expected_status:      Z_BIT_STATUS,
		},
		{
			name:                 "Positive value zero-page",
			memory:               []uint8{0x00, 0x00, 0x00, 0x7b},
			rom:                  []uint8{0xa5, 0x03},
			expected_accumulator: 0x7b,
			expected_status:      ZERO_BIT,
		},
		{
			name:                 "Negative value zero-page",
			memory:               []uint8{0x00, 0x00, 0x00, 0xfb},
			rom:                  []uint8{0xa5, 0x03},
			expected_accumulator: 0xfb,
			expected_status:      N_BIT_STATUS,
		},
		{
			name:                 "Zero value zero-page",
			rom:                  []uint8{0xa5, 0x03},
			expected_accumulator: 0x00,
			expected_status:      Z_BIT_STATUS,
		},
		{
			name:                 "Positive value zero-page X",
			memory:               []uint8{0x00, 0x00, 0x00, 0x7b},
			rom:                  []uint8{0xb5, 0x02},
			reg_x:                0x01,
			expected_accumulator: 0x7b,
			expected_status:      ZERO_BIT,
		},
		{
			name:                 "Negative value zero-page X",
			memory:               []uint8{0x00, 0x00, 0x00, 0xfb},
			rom:                  []uint8{0xb5, 0x02},
			reg_x:                0x01,
			expected_accumulator: 0xfb,
			expected_status:      N_BIT_STATUS,
		},
		{
			name:                 "Zero value zero-page X",
			rom:                  []uint8{0xb5, 0x02},
			reg_x:                0x01,
			expected_accumulator: 0x00,
			expected_status:      Z_BIT_STATUS,
		},
		// Add cases for indirect address modes.
	}

	for _, test := range testCases {
		callback := func(t *testing.T) {
			c := NewCPU()
			c.LoadAndReset(test.rom)

			for i, v := range test.memory {
				c.memory[i] = v
			}
			for i, v := range test.upper_memory {
				c.memory[i+0x1000] = v
			}
			c.index_x = test.reg_x
			c.index_y = test.reg_y

			result := c.Run()

			assert.Nil(t, result, "Error was not nil")
			assert.Equal(t, test.expected_accumulator, c.accumulator, "Accumulator incorrect")
			assert.Equal(t, test.expected_status, c.status, "Status incorrect")
		}
		t.Run(test.name, callback)
	}
}

func TestRun_TAX(t *testing.T) {
	testCases := []struct {
		name             string
		accumulator      uint8
		initial_status   uint8
		expected_index_x uint8
		expected_status  uint8
	}{
		{
			name:             "Positive accumulator",
			accumulator:      0x77,
			initial_status:   N_BIT_STATUS | Z_BIT_STATUS,
			expected_index_x: 0x77,
			expected_status:  ZERO_BIT,
		},
		{
			name:             "Negative accumulator",
			accumulator:      0xa2,
			initial_status:   Z_BIT_STATUS,
			expected_index_x: 0xa2,
			expected_status:  N_BIT_STATUS,
		},
		{
			name:             "Zero accumulator",
			accumulator:      0x00,
			initial_status:   N_BIT_STATUS,
			expected_index_x: 0x00,
			expected_status:  Z_BIT_STATUS,
		},
	}

	for _, test := range testCases {
		callback := func(t *testing.T) {
			c := NewCPU()
			c.LoadAndReset([]uint8{0xaa})
			c.accumulator = test.accumulator
			c.status = test.initial_status
			c.Run()

			assert.Equal(t, test.expected_index_x, c.index_x, "Index X register incorrect")
			assert.Equal(t, test.expected_status, c.status, "Status incorrect")
		}
		t.Run(test.name, callback)
	}
}

func TestRun_INX_Implied(t *testing.T) {
	testCases := []struct {
		name             string
		index_x          uint8
		initial_status   uint8
		expected_index_x uint8
		expected_n_bit   uint8
		expected_z_bit   uint8
	}{
		{
			name:             "Positive value",
			index_x:          0x77,
			initial_status:   N_BIT_STATUS | Z_BIT_STATUS,
			expected_index_x: 0x78,
			expected_n_bit:   ZERO_BIT,
			expected_z_bit:   ZERO_BIT,
		},
		{
			name:             "Positive to negative",
			index_x:          0x7f,
			initial_status:   ZERO_BIT,
			expected_index_x: 0x80,
			expected_n_bit:   N_BIT_STATUS,
			expected_z_bit:   ZERO_BIT,
		},
		{
			name:             "Negative value",
			index_x:          0xa2,
			initial_status:   N_BIT_STATUS,
			expected_index_x: 0xa3,
			expected_n_bit:   N_BIT_STATUS,
			expected_z_bit:   ZERO_BIT,
		},
		{
			name:             "Zero value",
			index_x:          0x00,
			initial_status:   Z_BIT_STATUS,
			expected_index_x: 0x01,
			expected_n_bit:   ZERO_BIT,
			expected_z_bit:   ZERO_BIT,
		},
		{
			name:             "Overflow",
			index_x:          0xff,
			initial_status:   N_BIT_STATUS,
			expected_index_x: 0x00,
			expected_n_bit:   ZERO_BIT,
			expected_z_bit:   Z_BIT_STATUS,
		},
	}

	for _, test := range testCases {
		callback := func(t *testing.T) {
			c := NewCPU()
			c.LoadAndReset([]uint8{0xe8})
			c.index_x = test.index_x
			c.status = test.initial_status
			c.Run()

			assert.Equal(t, test.expected_index_x, c.index_x, "Index X register incorrect")
			assert.Equal(t, test.expected_n_bit, c.status&N_BIT_STATUS, "Negative status bit incorrect")
			assert.Equal(t, test.expected_z_bit, c.status&Z_BIT_STATUS, "Zero status bit incorrect")
		}
		t.Run(test.name, callback)
	}
}

func TestRun_INY_Implied(t *testing.T) {
	testCases := []struct {
		name             string
		index_y          uint8
		initial_status   uint8
		expected_index_y uint8
		expected_n_bit   uint8
		expected_z_bit   uint8
	}{
		{
			name:             "Positive value",
			index_y:          0x77,
			initial_status:   N_BIT_STATUS | Z_BIT_STATUS,
			expected_index_y: 0x78,
			expected_n_bit:   ZERO_BIT,
			expected_z_bit:   ZERO_BIT,
		},
		{
			name:             "Positive to negative",
			index_y:          0x7f,
			initial_status:   ZERO_BIT,
			expected_index_y: 0x80,
			expected_n_bit:   N_BIT_STATUS,
			expected_z_bit:   ZERO_BIT,
		},
		{
			name:             "Negative value",
			index_y:          0xa2,
			initial_status:   N_BIT_STATUS,
			expected_index_y: 0xa3,
			expected_n_bit:   N_BIT_STATUS,
			expected_z_bit:   ZERO_BIT,
		},
		{
			name:             "Zero value",
			index_y:          0x00,
			initial_status:   Z_BIT_STATUS,
			expected_index_y: 0x01,
			expected_n_bit:   ZERO_BIT,
			expected_z_bit:   ZERO_BIT,
		},
		{
			name:             "Overflow",
			index_y:          0xff,
			initial_status:   N_BIT_STATUS,
			expected_index_y: 0x00,
			expected_n_bit:   ZERO_BIT,
			expected_z_bit:   Z_BIT_STATUS,
		},
	}

	for _, test := range testCases {
		callback := func(t *testing.T) {
			c := NewCPU()
			c.LoadAndReset([]uint8{0xc8})
			c.index_y = test.index_y
			c.status = test.initial_status
			c.Run()

			assert.Equal(t, test.expected_index_y, c.index_y, "Index Y register incorrect")
			assert.Equal(t, test.expected_n_bit, c.status&N_BIT_STATUS, "Negative status bit incorrect")
			assert.Equal(t, test.expected_z_bit, c.status&Z_BIT_STATUS, "Zero status bit incorrect")
		}
		t.Run(test.name, callback)
	}
}

func TestRun_ADC(t *testing.T) {
	testCases := []struct {
		name                 string
		memory               []uint8
		upper_memory         []uint8
		rom                  []uint8
		initial_accumulator  uint8
		expected_accumulator uint8
		expected_status      uint8
	}{
		{
			name:                 "Immediate, positive, no carry",
			rom:                  []uint8{0x69, 0x02},
			initial_accumulator:  0x03,
			expected_accumulator: 0x05,
			expected_status:      ZERO_BIT,
		},
		{
			name:                 "Immediate, positive, carry",
			rom:                  []uint8{0x69, 0xff},
			initial_accumulator:  0x03,
			expected_accumulator: 0x02,
			expected_status:      C_BIT_STATUS,
		},
		{
			name:                 "Immediate, negative, no carry",
			rom:                  []uint8{0x69, 0x7f},
			initial_accumulator:  0x03,
			expected_accumulator: 0x82,
			expected_status:      N_BIT_STATUS,
		},
		{
			name:                 "Immediate, negative, carry",
			rom:                  []uint8{0x69, 0xff},
			initial_accumulator:  0xff,
			expected_accumulator: 0xfe,
			expected_status:      C_BIT_STATUS | N_BIT_STATUS,
		},
		{
			name:                 "Immediate, zero, no carry",
			rom:                  []uint8{0x69, 0x00},
			initial_accumulator:  0x00,
			expected_accumulator: 0x00,
			expected_status:      Z_BIT_STATUS,
		},
		{
			name:                 "Immediate, zero, carry",
			rom:                  []uint8{0x69, 0xff},
			initial_accumulator:  0x01,
			expected_accumulator: 0x00,
			expected_status:      Z_BIT_STATUS | C_BIT_STATUS,
		},
		{
			name:                 "Zero-page, positive no carry",
			memory:               []uint8{0x00, 0x00, 0x00, 0x7b},
			rom:                  []uint8{0x65, 0x03},
			initial_accumulator:  0x03,
			expected_accumulator: 0x7e,
			expected_status:      ZERO_BIT,
		},
		{
			name:                 "Zero-page, positive, carry",
			memory:               []uint8{0x00, 0x00, 0xf0},
			rom:                  []uint8{0x65, 0x02},
			initial_accumulator:  0x13,
			expected_accumulator: 0x03,
			expected_status:      C_BIT_STATUS,
		},
		{
			name:                 "Zero-page, negative, no carry",
			memory:               []uint8{0x00, 0x00, 0x70},
			rom:                  []uint8{0x65, 0x02},
			initial_accumulator:  0x13,
			expected_accumulator: 0x83,
			expected_status:      N_BIT_STATUS,
		},
		{
			name:                 "Zero-page, negative, carry",
			memory:               []uint8{0x00, 0x00, 0xf1},
			rom:                  []uint8{0x65, 0x02},
			initial_accumulator:  0xff,
			expected_accumulator: 0xf0,
			expected_status:      C_BIT_STATUS | N_BIT_STATUS,
		},
	}

	for _, test := range testCases {
		callback := func(t *testing.T) {
			c := NewCPU()
			c.LoadAndReset(test.rom)

			for i, v := range test.memory {
				c.memory[i] = v
			}
			for i, v := range test.upper_memory {
				c.memory[i+0x1000] = v
			}

			c.accumulator = test.initial_accumulator

			result := c.Run()

			assert.Nil(t, result, "Error was not nil")
			assert.Equal(t, test.expected_accumulator, c.accumulator, "Accumulator incorrect")
			assert.Equal(t, test.expected_status, c.status, "Status incorrect")
		}
		t.Run(test.name, callback)
	}
}
