package bitmask

import (
	"fmt"
)

const uintSize = 32 << (^uint(0) >> 63) // 32 or 64

type BitMask struct {
	store []uint
	len   uint
}

func New(len uint) BitMask {
	return BitMask{store: make([]uint, (len+uintSize-1)/uintSize), len: len}
}

func NewFromUint(values ...uint) BitMask {
	store := make([]uint, len(values))
	copy(store, values)
	return BitMask{store: store, len: uintSize * uint(len(values))}
}

func (bm BitMask) Len() uint {
	return bm.len
}

func (bm BitMask) Set(index uint) {
	checkBounds(bm.len, index)
	bref, m := findBit(bm.store, index)
	*bref |= m
}

// func (bm BitMask) SetRange(fromIndex uint, toIndex uint) {
// 	checkBounds(bm.len, fromIndex)
// 	checkBounds(bm.len, toIndex)

// 	bref, m := findBit(bm.store, index)
// 	*bref |= m
// }

func (bm BitMask) Clear(index uint) {
	checkBounds(bm.len, index)
	bref, m := findBit(bm.store, index)
	*bref &^= m
}

func (bm BitMask) Toggle(index uint) {
	checkBounds(bm.len, index)
	bref, m := findBit(bm.store, index)
	*bref ^= m
}

func (bm BitMask) IsSet(index uint) bool {
	checkBounds(bm.len, index)
	bref, m := findBit(bm.store, index)
	return (*bref & m) != 0
}

// Copies bits from a source bit mask into a destination bit mask.
// Returns the number of bits copied, which will be the minimum of src.Len() and dst.Len().
func Copy(dst BitMask, src BitMask) uint {
	bitsN := dst.len
	if src.len < dst.len {
		bitsN = src.len
	}

	// copy whole part
	wordsN := bitsN / uintSize
	copy(dst.store[:wordsN], src.store[:wordsN])

	// copy remainder
	remainderBitsN := bitsN % uintSize
	if remainderBitsN == 0 {
		return bitsN
	}

	shift := uintSize - remainderBitsN
	mask := (^uint(0) >> shift) << shift
	dst.store[wordsN] |= src.store[wordsN] | mask

	return bitsN
}

// // Selects a half-open range which includes the "from" bit, but excludes the "to" one.
// func (bm BitMask) Slice(fromIndex uint64, toIndex uint64) BitMask {
// 	fromWordIndex := fromIndex / STORE_WORD_SIZE

// 	return BitMask{
// 		store: bm.store[fromIndex:toIndex],
// 		len:   toIndex - fromIndex,
// 	}
// }

// func (bm BitMask) String() string {
// 	if bm.len <= 128 {
// 		return fmt.Sprintf("%08b", bm.store)
// 	}
// 	left := bm.store[:64]
// 	right := bm.store[bm.len-64:]
// 	return fmt.Sprintf("%08b <skipped  items> %08b", bm.sli)

// }

func checkBounds(len uint, index uint) {
	if index >= len {
		panic(fmt.Sprintf("index out of range (len=%v, index=%v)", len, index))
	}
}

func findBit(store []uint, index uint) (*uint, uint) {
	storeIndex := index / uintSize
	byteMask := uint(1) << (index % uintSize)
	return &store[storeIndex], byteMask
}
