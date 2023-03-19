package bitmask

import (
	"fmt"
)

const STORE_WORD_SIZE = uint64(64)

type BitMask struct {
	store []uint64
	len   uint64
}

func New(len uint64) BitMask {
	return BitMask{store: make([]uint64, (len+STORE_WORD_SIZE-1)/STORE_WORD_SIZE), len: len}
}

func NewFromUInt64(values ...uint64) BitMask {
	store := make([]uint64, len(values))
	copy(store, values)
	wordSize := uint64(64)
	return BitMask{store: store, len: wordSize * uint64(len(values))}
}

func (bm BitMask) Len() uint64 {
	return bm.len
}

func (bm BitMask) Set(index uint64) {
	checkBounds(bm.len, index)
	bref, m := findBit(bm.store, index)
	*bref |= m
}

// func (bm BitMask) SetRange(fromIndex uint64, toIndex uint64) {
// 	checkBounds(bm.len, fromIndex)
// 	checkBounds(bm.len, toIndex)

// 	bref, m := findBit(bm.store, index)
// 	*bref |= m
// }

func (bm BitMask) Clear(index uint64) {
	checkBounds(bm.len, index)
	bref, m := findBit(bm.store, index)
	*bref &^= m
}

func (bm BitMask) Toggle(index uint64) {
	checkBounds(bm.len, index)
	bref, m := findBit(bm.store, index)
	*bref ^= m
}

func (bm BitMask) IsSet(index uint64) bool {
	checkBounds(bm.len, index)
	bref, m := findBit(bm.store, index)
	return (*bref & m) != 0
}

// Copies bits from a source bit mask into a destination bit mask.
// Returns the number of bits copied, which will be the minimum of src.Len() and dst.Len().
func Copy(dst BitMask, src BitMask) uint64 {
	bitsN := dst.len
	if src.len < dst.len {
		bitsN = src.len
	}

	// copy whole part
	wordsN := bitsN / STORE_WORD_SIZE
	copy(dst.store[:wordsN], src.store[:wordsN])

	// copy remainder
	remainderBitsN := bitsN % STORE_WORD_SIZE
	if remainderBitsN == 0 {
		return bitsN
	}

	shift := STORE_WORD_SIZE - remainderBitsN
	mask := (^uint64(0) >> shift) << shift
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

func checkBounds(len uint64, index uint64) {
	if index >= len {
		panic(fmt.Sprintf("index out of range (len=%v, index=%v)", len, index))
	}
}

func findBit(store []uint64, index uint64) (*uint64, uint64) {
	storeIndex := index / STORE_WORD_SIZE
	byteMask := uint64(1) << (index % STORE_WORD_SIZE)
	return &store[storeIndex], byteMask
}
