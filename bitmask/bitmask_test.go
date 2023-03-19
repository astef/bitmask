package bitmask

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmpty(t *testing.T) {
	bm := BitMask{}
	bm2 := New(0)
	assert.Equal(t, bm.Len(), uint(0))
	assert.Equal(t, bm2.Len(), uint(0))
}

type sctTestCase struct {
	n           uint
	setEvery    uint
	clearEvery  uint
	toggleEvery uint
}

func TestSetClearToggle(t *testing.T) {
	nS, nM, nL, nXL := genN()
	each := uint(1)
	none := uint(math.MaxUint)

	tests := map[string]sctTestCase{
		"only_set_S":  {nS, each, none, none},
		"only_set_M":  {nM, each, none, none},
		"only_set_L":  {nL, each, none, none},
		"only_set_XL": {nXL, each, none, none},

		"only_clear_S":  {nS, none, each, none},
		"only_clear_M":  {nM, none, each, none},
		"only_clear_L":  {nL, none, each, none},
		"only_clear_XL": {nXL, none, each, none},

		"set_and_clear_every_S":             {nS, each, each, none},
		"set_and_clear_every_M":             {nM, each, each, none},
		"set_and_clear_every_L":             {nL, each, each, none},
		"set_and_clear_every_XL":            {nXL, each, each, none},
		"set_and_clear_and_toggle_every_XL": {nXL, each, each, each},

		"set_every_2_and_clear_every_4_S":  {nS, 2, 4, none},
		"set_every_2_and_clear_every_4_M":  {nM, 2, 4, none},
		"set_every_2_and_clear_every_4_L":  {nL, 2, 4, none},
		"set_every_2_and_clear_every_4_XL": {nXL, 2, 4, none},

		"set_every_3_and_clear_every_4_M":                     {nM, 3, 4, none},
		"set_every_3_and_clear_every_4_L":                     {nL, 3, 4, none},
		"set_every_3_and_clear_every_4_XL":                    {nXL, 3, 4, none},
		"set_every_3_and_clear_every_4_and_toggle_every_5_XL": {nXL, 3, 4, 5},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			bm := New(tc.n)

			// setting
			for i := uint(0); i < bm.Len(); i++ {
				if i%tc.setEvery == 0 {
					bm.Set(i)
				}
			}

			// clearing
			for i := uint(0); i < bm.Len(); i++ {
				if i%tc.clearEvery == 0 {
					bm.Clear(i)
				}
			}

			// toggling
			for i := uint(0); i < bm.Len(); i++ {
				if i%tc.toggleEvery == 0 {
					bm.Toggle(i)
				}
			}

			// checking
			expectedSetButClear := []uint{}
			expectedClearButSet := []uint{}
			for i := uint(0); i < bm.Len(); i++ {

				wasSet := i%tc.setEvery == 0
				wasCleared := i%tc.clearEvery == 0
				wasToggled := i%tc.toggleEvery == 0
				isSetExpected := wasSet
				if wasCleared {
					isSetExpected = false
				}
				if wasToggled {
					isSetExpected = !isSetExpected
				}

				isSetActual := bm.IsSet(i)
				if isSetExpected && !isSetActual {
					expectedSetButClear = append(expectedSetButClear, i)
				}
				if !isSetExpected && isSetActual {
					expectedClearButSet = append(expectedClearButSet, i)
				}
			}
			assert.Equalf(
				t,
				0,
				len(expectedSetButClear),
				"in mask of size %v bits were expected to be set, but they are clear: %v",
				tc.n,
				expectedSetButClear)
			assert.Equalf(
				t,
				0,
				len(expectedClearButSet),
				"in mask of size %v bits were expected to be clear, but they are set: %v",
				tc.n,
				expectedClearButSet)
		})
	}
}

type copyTestCase struct {
	src         []uint
	srcLen      uint
	dst         []uint
	expectedDst []uint
	expectedN   uint
}

func TestCopy(t *testing.T) {

	tests := map[string]copyTestCase{
		"equal_size": {
			src:         []uint{0, 0},
			srcLen:      128,
			dst:         []uint{math.MaxUint, math.MaxUint},
			expectedDst: []uint{0, 0},
			expectedN:   uint(128),
		},
		"small_src": {
			src:         []uint{math.MaxUint, 1 << 63},
			srcLen:      65,
			dst:         []uint{0, math.MaxUint >> 1, math.MaxUint},
			expectedDst: []uint{math.MaxUint, math.MaxUint, math.MaxUint},
			expectedN:   uint(65),
		},
		"small_dst": {
			src:         []uint{math.MaxUint, math.MaxUint, math.MaxUint},
			srcLen:      192,
			dst:         []uint{0, 0},
			expectedDst: []uint{math.MaxUint, math.MaxUint},
			expectedN:   uint(128),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			srcProto := NewFromUint(tc.src...)
			src := New(tc.srcLen)
			Copy(src, srcProto)

			dst := NewFromUint(tc.dst...)
			expected := NewFromUint(tc.expectedDst...)

			// act
			n := Copy(dst, src)

			// assert
			assert.Equal(t, tc.expectedN, n)
			assert.Equal(t, expected, dst)
		})
	}

}

func genN() (nS uint, nM uint, nL uint, nXL uint) {
	sws := int(uintSize)
	nS, nM, nL, nXL =
		uint(1),
		uint(2),
		uint(3+rand.Intn(sws*2)),
		uint(sws*2+rand.Intn(sws*3))
	return
}
