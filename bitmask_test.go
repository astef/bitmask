package bitmask

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// in a lot of places, tests are relying on uintSize==64

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

func genN() (nS uint, nM uint, nL uint, nXL uint) {
	sws := int(uintSize)
	nS, nM, nL, nXL =
		uint(1),
		uint(2),
		uint(3+rand.Intn(sws*2)),
		uint(sws*2+rand.Intn(sws*3))
	return
}

type slice struct{ from, to uint }

type copyTestCase struct {
	base         *BitMask
	srcSlice     struct{ from, to uint }
	dstSlice     struct{ from, to uint }
	expectedBase string
}

func TestCopy(t *testing.T) {
	tests := map[string]copyTestCase{
		"empty": {
			base:         NewFromUint(uintMax),
			srcSlice:     slice{0, 0},
			dstSlice:     slice{0, uintSize},
			expectedBase: NewFromUint(uintMax).String(),
		},
		"equal_size": {
			base:         NewFromUint(uintMax, 0),
			srcSlice:     slice{0, uintSize},
			dstSlice:     slice{uintSize, 2 * uintSize},
			expectedBase: NewFromUint(uintMax, uintMax).String(),
		},
		"small_src": {
			base:         NewFromUint(1, 0),
			srcSlice:     slice{0, 1},
			dstSlice:     slice{uintSize, 2 * uintSize},
			expectedBase: NewFromUint(1, 1).String(),
		},
		"small_dst": {
			base:         NewFromUint(uintMax, 0),
			srcSlice:     slice{0, uintSize},
			dstSlice:     slice{uintSize, uintSize + 1},
			expectedBase: NewFromUint(uintMax, 1).String(),
		},

		"1w_overlap1_fw": {
			base:         NewFromUint(reverse(0b1000000000000000000000000000000000000000000000000000000000000010)),
			srcSlice:     slice{0, uintSize - 1},
			dstSlice:     slice{1, uintSize},
			expectedBase: NewFromUint(reverse(0b1100000000000000000000000000000000000000000000000000000000000001)).String(),
		},
		"1w_overlap2_fw": {
			base:         NewFromUint(reverse(0b1000000000000000000000000000000000000000000000000000000000000100)),
			srcSlice:     slice{0, uintSize - 2},
			dstSlice:     slice{2, uintSize},
			expectedBase: NewFromUint(reverse(0b1010000000000000000000000000000000000000000000000000000000000001)).String(),
		},

		"1w_overlap1_fw_inverse": {
			base:         NewFromUint(reverse(0b0111111111111111111111111111111111111111111111111111111111111101)),
			srcSlice:     slice{0, uintSize - 1},
			dstSlice:     slice{1, uintSize},
			expectedBase: NewFromUint(reverse(0b0011111111111111111111111111111111111111111111111111111111111110)).String(),
		},
		"1w_overlap2_fw_inverse": {
			base:         NewFromUint(reverse(0b0111111111111111111111111111111111111111111111111111111111111011)),
			srcSlice:     slice{0, uintSize - 2},
			dstSlice:     slice{2, uintSize},
			expectedBase: NewFromUint(reverse(0b0101111111111111111111111111111111111111111111111111111111111110)).String(),
		},

		"1w_overlap1_bw": {
			base:         NewFromUint(reverse(0b0100000000000000000000000000000000000000000000000000000000000001)),
			srcSlice:     slice{1, uintSize},
			dstSlice:     slice{0, uintSize - 1},
			expectedBase: NewFromUint(reverse(0b1000000000000000000000000000000000000000000000000000000000000011)).String(),
		},
		"1w_overlap2_bw": {
			base:         NewFromUint(reverse(0b0010000000000000000000000000000000000000000000000000000000000001)),
			srcSlice:     slice{2, uintSize},
			dstSlice:     slice{0, uintSize - 2},
			expectedBase: NewFromUint(reverse(0b1000000000000000000000000000000000000000000000000000000000000101)).String(),
		},

		"2w_overlap1l_fw": {
			base:         NewFromUint(0, uintMax),
			srcSlice:     slice{0, uintSize},
			dstSlice:     slice{uintSize - 1, 2 * uintSize},
			expectedBase: NewFromUint(0, ^(uintMax >> 1)).String(),
		},
		"2w_overlap1r_fw": {
			base:         NewFromUint(uintMax, 0, uintMax),
			srcSlice:     slice{0, uintSize + 1},
			dstSlice:     slice{uintSize, 2 * uintSize},
			expectedBase: NewFromUint(uintMax, uintMax, uintMax).String(),
		},

		"2w_overlap32r_fw": {
			base:         NewFromUint(0, uintMax),
			srcSlice:     slice{uintSize / 2, uintSize + uintSize/2},
			dstSlice:     slice{uintSize, 2 * uintSize},
			expectedBase: NewFromUint(0, reverse(uintMax>>(uintSize/2))).String(),
		},
		"2w_overlap32r_bw": {
			base:         NewFromUint(0, uintMax),
			srcSlice:     slice{uintSize, 2 * uintSize},
			dstSlice:     slice{uintSize / 2, uintSize + uintSize/2},
			expectedBase: NewFromUint(reverse(uintMax>>(uintSize/2)), uintMax).String(),
		},

		"3w_overlap_same_offset0_fw": {
			base:         NewFromUint(uintMax, 0, uintMax),
			srcSlice:     slice{0, uintSize + 2},
			dstSlice:     slice{uintSize, 2*uintSize + 2},
			expectedBase: NewFromUint(uintMax, uintMax, reverse(uintMax>>2)).String(),
		},
		"3w_overlap_same_offset0_bw": {
			base:         NewFromUint(uintMax, 0, uintMax),
			srcSlice:     slice{uintSize, 2*uintSize + 2},
			dstSlice:     slice{0, uintSize + 2},
			expectedBase: NewFromUint(0, reverse(0b1100000000000000000000000000000000000000000000000000000000000000), uintMax).String(),
		},

		"3w_overlap_same_offset1_fw": {
			base:         NewFromUint(uintMax, 0, uintMax),
			srcSlice:     slice{1, uintSize + 2},
			dstSlice:     slice{uintSize + 1, 2*uintSize + 2},
			expectedBase: NewFromUint(uintMax, reverse(uintMax>>1), reverse(uintMax>>2)).String(),
		},
		"3w_overlap_same_offset1_bw": {
			base:         NewFromUint(uintMax, 0, uintMax),
			srcSlice:     slice{uintSize + 1, 2*uintSize + 2},
			dstSlice:     slice{1, uintSize + 2},
			expectedBase: NewFromUint(1, reverse(0b1100000000000000000000000000000000000000000000000000000000000000), uintMax).String(),
		},

		"2w_small_dst_same_offset0": {
			base:         NewFromUint(uintMax, 0),
			srcSlice:     slice{0, uintSize},
			dstSlice:     slice{uintSize, uintSize + 2},
			expectedBase: NewFromUint(uintMax, reverse(0b1100000000000000000000000000000000000000000000000000000000000000)).String(),
		},
		"3w_small_dst_same_offset0": {
			base:         NewFromUint(uintMax, 0, 0),
			srcSlice:     slice{0, uintSize},
			dstSlice:     slice{uintSize * 2, uintSize*2 + 2},
			expectedBase: NewFromUint(uintMax, 0, reverse(0b1100000000000000000000000000000000000000000000000000000000000000)).String(),
		},

		"2w_small_dst_same_offset2": {
			base:         NewFromUint(uintMax, 0),
			srcSlice:     slice{2, uintSize},
			dstSlice:     slice{uintSize + 2, uintSize + 4},
			expectedBase: NewFromUint(uintMax, reverse(0b0011000000000000000000000000000000000000000000000000000000000000)).String(),
		},
		"3w_small_dst_same_offset2": {
			base:         NewFromUint(uintMax, 0, 0),
			srcSlice:     slice{2, uintSize},
			dstSlice:     slice{uintSize*2 + 2, uintSize*2 + 4},
			expectedBase: NewFromUint(uintMax, 0, reverse(0b0011000000000000000000000000000000000000000000000000000000000000)).String(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			src := tc.base.Slice(tc.srcSlice.from, tc.srcSlice.to)
			dst := tc.base.Slice(tc.dstSlice.from, tc.dstSlice.to)

			// act
			n := Copy(dst, src)

			// assert
			assert.Equal(
				t,
				int(minUint(tc.srcSlice.to-tc.srcSlice.from, tc.dstSlice.to-tc.dstSlice.from)),
				int(n),
				"copy returned wrong number of bits")
			assert.Equal(t, tc.expectedBase, tc.base.String())
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
	// TODO rewrite without impl details
	store := []uint{0, uintMax, 0}
	storeLen := uint(len(store))
	bm := BitMask{
		store:  store,
		len:    storeLen * uintSize,
		offset: 0,
	}
	minOffsetBm := BitMask{
		store:  store,
		len:    storeLen*uintSize - 1,
		offset: 1,
	}
	min2OffsetBm := BitMask{
		store:  store,
		len:    storeLen*uintSize - 2,
		offset: 2,
	}
	maxOffsetBm := BitMask{
		store:  store,
		len:    (storeLen-1)*uintSize + 1,
		offset: uintSize - 1,
	}
	withoutFirstBm := BitMask{
		store:  store[1:],
		len:    (storeLen - 1) * uintSize,
		offset: 0,
	}
	twoBitsBetweenFirstAndSecondBm := BitMask{
		store:  store[:2],
		len:    2,
		offset: uintSize - 1,
	}
	tests := map[string]sliceTestCase{
		"full_no_offset": {
			source:   bm,
			from:     0,
			to:       bm.len,
			expected: bm,
		},
		"full_min_offset": {
			source:   minOffsetBm,
			from:     0,
			to:       minOffsetBm.len,
			expected: minOffsetBm,
		},
		"full_max_offset": {
			source:   maxOffsetBm,
			from:     0,
			to:       maxOffsetBm.len,
			expected: maxOffsetBm,
		},
		"from1_no_offset": {
			source:   bm,
			from:     1,
			to:       bm.len,
			expected: minOffsetBm,
		},
		"from1_min_offset": {
			source:   minOffsetBm,
			from:     1,
			to:       minOffsetBm.len,
			expected: min2OffsetBm,
		},
		"from1_max_offset": {
			source:   maxOffsetBm,
			from:     1,
			to:       maxOffsetBm.len,
			expected: withoutFirstBm,
		},
		"from0_to2_maxoffset": {
			source:   maxOffsetBm,
			from:     0,
			to:       2,
			expected: twoBitsBetweenFirstAndSecondBm,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual := tc.source.Slice(tc.from, tc.to)

			// assert
			assert.Equal(t, tc.expected, *actual)
		})
	}
}

func TestIterator(t *testing.T) {
	bm := NewFromUint(uintMax, 0).Slice(62, 66)
	it := bm.Iterator()

	for i := uint(0); i < 4; i++ {
		ok, value, index := it.Next()
		assert.Equal(t, i, index)
		assert.Equal(t, true, ok)
		assert.Equal(t, i < 2, value)
	}
	ok, _, _ := it.Next()
	assert.Equal(t, false, ok)
}

func TestSliceWithString(t *testing.T) {
	tests := map[string]struct {
		source   *BitMask
		from     uint
		to       uint
		expected string
	}{
		"0w_empty": {New(0), 0, 0, "[0]{}"},
		"1w_full":  {NewFromUint(0), 0, 64, "[64]{0000000000000000000000000000000000000000000000000000000000000000}"},
		"1w1_full": {NewFromUint(1), 0, 64, "[64]{1000000000000000000000000000000000000000000000000000000000000000}"},
		"1w_left": {
			NewFromUint(0b0000000000000000000000000000000000000000000000000000001000000101),
			0,
			10,
			"[10]{1010000001}",
		},
		"1w_right": {
			NewFromUint(0b1110000000000000000000000000000000000000000000000000000000001001),
			uintSize - 4,
			uintSize,
			"[4]{0111}",
		},
		"1w_middle": {
			NewFromUint(0b0000000000000000000010010110000000000000000000000000000000000000),
			37,
			44,
			"[7]{1101001}",
		},

		"2w_lleft": {
			NewFromUint(0b0000000000000000000000000000000000000000000000000000001111111101, uintMax),
			0,
			10,
			"[10]{1011111111}",
		},
		"2w_lright": {
			NewFromUint(0b1110000000000000000000000000000000000000000000000000000000001001, uintMax),
			uintSize - 4,
			uintSize,
			"[4]{0111}",
		},
		"2w_lmiddle": {
			NewFromUint(0b0000000000000000000010010110000000000000000000000000000000000000, uintMax),
			37,
			44,
			"[7]{1101001}",
		},

		"2w_rleft": {
			NewFromUint(uintMax, 0b0000000000000000000000000000000000000000000000000000001011111111),
			uintSize,
			uintSize + 10,
			"[10]{1111111101}",
		},
		"2w_rright": {
			NewFromUint(uintMax, 0b1110000000000000000000000000000000000000000000000000000000001001),
			uintSize + uintSize - 4,
			uintSize + uintSize,
			"[4]{0111}",
		},
		"2w_rmiddle": {
			NewFromUint(uintMax, 0b0000000000000000000010010110000000000000000000000000000000000000),
			uintSize + 37,
			uintSize + 44,
			"[7]{1101001}",
		},

		"2w_middle": {
			NewFromUint(uintMax, 0),
			uintSize - 3,
			uintSize + 2,
			"[5]{111 00}",
		},

		"3w_middle": {
			NewFromUint(uintMax, 0, uintMax),
			uintSize,
			uintSize * 2,
			"[64]{0000000000000000000000000000000000000000000000000000000000000000}",
		},
		"3w_middle_outer": {
			NewFromUint(uintMax, 0, uintMax),
			uintSize - 1,
			uintSize*2 + 1,
			"[66]{1 0000000000000000000000000000000000000000000000000000000000000000 1}",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual := tc.source.Slice(tc.from, tc.to)

			// assert
			assert.Equal(t, tc.expected, actual.String())
		})
	}
}

type rangeOps int

const (
	SetAll rangeOps = iota
	ClearAll
	ToggleAll
)

func TestRangeOps(t *testing.T) {
	tests := map[string]struct {
		source   *BitMask
		ops      []rangeOps
		expected string
	}{
		"zerolen": {
			New(0),
			[]rangeOps{ToggleAll, ClearAll, SetAll},
			"[0]{}",
		},
		"1w_cleared_clear": {
			NewFromUint(0),
			[]rangeOps{ClearAll},
			"[64]{0000000000000000000000000000000000000000000000000000000000000000}",
		},
		"1w_set_clear": {
			NewFromUint(uintMax),
			[]rangeOps{ClearAll},
			"[64]{0000000000000000000000000000000000000000000000000000000000000000}",
		},

		"1w_cleared_set": {
			NewFromUint(0),
			[]rangeOps{SetAll},
			"[64]{1111111111111111111111111111111111111111111111111111111111111111}",
		},
		"1w_set_set": {
			NewFromUint(uintMax),
			[]rangeOps{SetAll},
			"[64]{1111111111111111111111111111111111111111111111111111111111111111}",
		},

		"1w_cleared_toggle": {
			NewFromUint(0),
			[]rangeOps{ToggleAll},
			"[64]{1111111111111111111111111111111111111111111111111111111111111111}",
		},
		"1w_set_toggle": {
			NewFromUint(uintMax),
			[]rangeOps{ToggleAll},
			"[64]{0000000000000000000000000000000000000000000000000000000000000000}",
		},

		"65b_mixed_clear": {
			NewFromUint(0, 1).Slice(0, 65),
			[]rangeOps{ClearAll},
			"[65]{0000000000000000000000000000000000000000000000000000000000000000 0}",
		},
		"65b_mixed_set": {
			NewFromUint(uintMax, 0).Slice(0, 65),
			[]rangeOps{SetAll},
			"[65]{1111111111111111111111111111111111111111111111111111111111111111 1}",
		},
		"65b_mixed_toggle": {
			NewFromUint(uintMax, 0).Slice(0, 65),
			[]rangeOps{ToggleAll},
			"[65]{0000000000000000000000000000000000000000000000000000000000000000 1}",
		},
		"65b_mixed_and_sliced_toggle": {
			NewFromUint(uintMax, 0).Slice(5, 65),
			[]rangeOps{ToggleAll},
			"[60]{00000000000000000000000000000000000000000000000000000000000 1}",
		},

		"3w_mixed_and_2xsliced_toggle": {
			NewFromUint(uintMax, 0, uintMax).Slice(1, 3*uintSize-1).Slice(uintSize-2, 2*uintSize),
			[]rangeOps{ToggleAll},
			"[66]{0 1111111111111111111111111111111111111111111111111111111111111111 0}",
		},
		"3w_mixed_and_2xsliced_2xtoggle": {
			NewFromUint(uintMax, 0, uintMax).Slice(1, 3*uintSize-1).Slice(uintSize-2, 2*uintSize),
			[]rangeOps{ToggleAll, ToggleAll},
			"[66]{1 0000000000000000000000000000000000000000000000000000000000000000 1}",
		},
		"3w_mixed_and_2xsliced_3xtoggle": {
			NewFromUint(uintMax, 0, uintMax).Slice(1, 3*uintSize-1).Slice(uintSize-2, 2*uintSize),
			[]rangeOps{ToggleAll, ToggleAll, ToggleAll},
			"[66]{0 1111111111111111111111111111111111111111111111111111111111111111 0}",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual := tc.source
			for _, op := range tc.ops {
				if op == SetAll {
					actual.SetAll()
				} else if op == ClearAll {
					actual.ClearAll()
				} else {
					actual.ToggleAll()
				}
			}

			// assert
			assert.Equal(t, tc.expected, actual.String())
		})
	}
}

func zeros(n uint) string {
	return fmt.Sprintf("%0"+fmt.Sprint(n)+"b", 0)
}

func zerosWords(n uint) string {
	var b strings.Builder
	for i := uint(0); i < n; i++ {
		if i != 0 {
			b.WriteString(" ")
		}
		b.WriteString(zeros(uintSize))
	}
	return b.String()
}

func TestStringSkips(t *testing.T) {
	tests := map[string]struct {
		source   *BitMask
		expected string
	}{
		"512_not_skipped": {New(512), "[512]{" + zerosWords(8) + "}"},
		"513":             {New(513), "[513]{" + zerosWords(4) + " <more 64 bits> " + zerosWords(3) + " 0}"},
		"576":             {New(576), "[576]{" + zerosWords(4) + " <more 64 bits> " + zerosWords(4) + "}"},
		"640":             {New(640), "[640]{" + zerosWords(4) + " <more 128 bits> " + zerosWords(4) + "}"},
		"641":             {New(641), "[641]{" + zerosWords(4) + " <more 192 bits> " + zerosWords(3) + " 0}"},
		"639":             {New(639), "[639]{" + zerosWords(4) + " <more 128 bits> " + zerosWords(3) + " " + zeros(63) + "}"},
		"960":             {New(960), "[960]{" + zerosWords(4) + " <more 448 bits> " + zerosWords(4) + "}"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.source.String())
		})
	}
}

func TestNewFromUint(t *testing.T) {
	bm := NewFromUint(1)
	bm.Set(1)
	bm.Toggle(2)
	bm.Toggle(3)
	bm.Clear(3)
	assert.Equal(t, "[64]{1110000000000000000000000000000000000000000000000000000000000000}", bm.String())
}

func TestDocExample(t *testing.T) {
	// [4]{0000}
	bm := New(4)

	// [4]{0001}
	bm.Set(3)

	// [4]{0010}
	bm.Slice(2, 4).ToggleAll()

	// [4]{0100}
	bm.Slice(1, 3).ToggleAll()

	// [4]{1000}
	bm.Slice(0, 2).ToggleAll()

	assert.Equal(t, "[4]{1000}", bm.String())
}
