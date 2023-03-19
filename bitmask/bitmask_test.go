package bitmask

import (
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
	none := uintMax

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
			srcLen:      uintSize * 2,
			dst:         []uint{uintMax, uintMax},
			expectedDst: []uint{0, 0},
			expectedN:   uintSize * 2,
		},
		"small_src": {
			src:         []uint{uintMax, 1 << (uintSize - 1)},
			srcLen:      uintSize + 1,
			dst:         []uint{0, uintMax >> 1, uintMax},
			expectedDst: []uint{uintMax, uintMax, uintMax},
			expectedN:   uintSize + 1,
		},
		"small_dst": {
			src:         []uint{uintMax, uintMax, uintMax},
			srcLen:      uintSize * 3,
			dst:         []uint{0, 0},
			expectedDst: []uint{uintMax, uintMax},
			expectedN:   uintSize * 2,
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

type sliceTestCase struct {
	source BitMask

	from uint
	to   uint

	expected BitMask
}

func TestSlice(t *testing.T) {
	store := []uint{0, uintMax, 0}
	storeLen := uint(len(store))
	bm := BitMask{
		store: store,
		len:   storeLen * uintSize,
		shift: 0,
	}
	minShiftBm := BitMask{
		store: store,
		len:   storeLen*uintSize - 1,
		shift: 1,
	}
	min2ShiftBm := BitMask{
		store: store,
		len:   storeLen*uintSize - 2,
		shift: 2,
	}
	maxShiftBm := BitMask{
		store: store,
		len:   (storeLen-1)*uintSize + 1,
		shift: uintSize - 1,
	}
	withoutFirstBm := BitMask{
		store: store[1:],
		len:   (storeLen - 1) * uintSize,
		shift: 0,
	}
	twoBitsBetweenFirstAndSecondBm := BitMask{
		store: store[:2],
		len:   2,
		shift: uintSize - 1,
	}
	tests := map[string]sliceTestCase{
		"full_no_shift": {
			source:   bm,
			from:     0,
			to:       bm.len,
			expected: bm,
		},
		"full_min_shift": {
			source:   minShiftBm,
			from:     0,
			to:       minShiftBm.len,
			expected: minShiftBm,
		},
		"full_max_shift": {
			source:   maxShiftBm,
			from:     0,
			to:       maxShiftBm.len,
			expected: maxShiftBm,
		},
		"from1_no_shift": {
			source:   bm,
			from:     1,
			to:       bm.len,
			expected: minShiftBm,
		},
		"from1_min_shift": {
			source:   minShiftBm,
			from:     1,
			to:       minShiftBm.len,
			expected: min2ShiftBm,
		},
		"from1_max_shift": {
			source:   maxShiftBm,
			from:     1,
			to:       maxShiftBm.len,
			expected: withoutFirstBm,
		},
		"from0_to2_maxshift": {
			source:   maxShiftBm,
			from:     0,
			to:       2,
			expected: twoBitsBetweenFirstAndSecondBm,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if name != "from0_to2_maxshift" {
				return
			}

			actual := tc.source.Slice(tc.from, tc.to)

			// assert
			assert.Equal(t, tc.expected, *actual)
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
