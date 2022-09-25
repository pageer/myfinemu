package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const ZERO_BIT uint8 = 0

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

func TestRun_LDA(t *testing.T) {
	testCases := []struct {
		name                 string
		memory               []uint8
		initial_status       uint8
		expected_accumulator uint8
		expected_status      uint8
	}{
		{
			name:                 "Positive value",
			memory:               []uint8{0xa9, 0x7b},
			initial_status:       N_BIT_STATUS | Z_BIT_STATUS,
			expected_accumulator: 0x7b,
			expected_status:      ZERO_BIT,
		},
		{
			name:                 "Negative value",
			memory:               []uint8{0xa9, 0xfb},
			initial_status:       Z_BIT_STATUS,
			expected_accumulator: 0xfb,
			expected_status:      N_BIT_STATUS,
		},
		{
			name:                 "Zero value",
			memory:               []uint8{0xa9, 0x00},
			initial_status:       N_BIT_STATUS,
			expected_accumulator: 0x00,
			expected_status:      Z_BIT_STATUS,
		},
	}

	for _, test := range testCases {
		callback := func(t *testing.T) {
			c := NewCPU()
			c.LoadAndReset(test.memory)
			c.status = test.initial_status
			c.Run()

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
			expected_index_x: 0x01,
			expected_n_bit:   ZERO_BIT,
			expected_z_bit:   ZERO_BIT,
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
