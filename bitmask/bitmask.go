package bitmask

import (
	"fmt"
)

const uintMax = ^uint(0)
const uintSize = 32 << (uintMax >> 63) // 32 or 64

var uintFormat = fmt.Sprintf("%%0%vb", uintSize)

const maxStringedUints = 8

type BitMask struct {
	store []uint
	len   uint
	// always in range [0 ; uintSize)
	shift uint
}

func New(len uint) *BitMask {
	return &BitMask{store: make([]uint, (len+uintSize-1)/uintSize), len: len}
}

func NewFromUint(values ...uint) *BitMask {
	store := make([]uint, len(values))
	copy(store, values)
	return &BitMask{store: store, len: uintSize * uint(len(values))}
}

// Returns the legth of bitmask in bits. It will never be changed for the given receiver.
func (bm *BitMask) Len() uint {
	return bm.len
}

func (bm *BitMask) Set(index uint) {
	checkBounds(bm.len, index)
	bref, m := bm.getBit(index)
	*bref |= m
}

// func (bm BitMask) SetRange(fromIndex uint, toIndex uint) {
// 	checkBounds(bm.len, fromIndex)
// 	checkBounds(bm.len, toIndex)

// 	bref, m := findBit(bm.store, index)
// 	*bref |= m
// }

func (bm *BitMask) Clear(index uint) {
	checkBounds(bm.len, index)
	bref, m := bm.getBit(index)
	*bref &^= m
}

func (bm *BitMask) Toggle(index uint) {
	checkBounds(bm.len, index)
	bref, m := bm.getBit(index)
	*bref ^= m
}

func (bm *BitMask) IsSet(index uint) bool {
	checkBounds(bm.len, index)
	bref, m := bm.getBit(index)
	return (*bref & m) != 0
}

// Copies bits from a source bit mask into a destination bit mask.
// Returns the number of bits copied, which will be the minimum of src.Len() and dst.Len().
func Copy(dst *BitMask, src *BitMask) uint {
	copyLen := minUint(src.len, dst.len)
	if copyLen == 0 {
		return 0
	}

	srcShift := src.shift
	dstShift := dst.shift

	if srcShift == dstShift {
		var copyStartIndex uint = 0
		copyEndIndex := ((copyLen - 1 + srcShift) / uintSize) + 1
		remainderBitsN := (copyLen + srcShift) % uintSize

		if srcShift != 0 {
			// have to copy first uint manually
			copyUintPart(
				uintSize-srcShift,
				src.store[copyStartIndex],
				srcShift,
				&dst.store[copyStartIndex],
				dstShift,
			)
			copyStartIndex++
		}

		if remainderBitsN != 0 {
			// have to copy last uint manually
			copyEndIndex--
			copyUintPart(
				remainderBitsN,
				src.store[copyEndIndex],
				0,
				&dst.store[copyEndIndex],
				0,
			)
		}

		if copyStartIndex != copyEndIndex {
			// copy whole part
			copy(dst.store[copyStartIndex:copyEndIndex], src.store[copyStartIndex:copyEndIndex])
		}
	} else {
		// not optimized copy, by uint parts
		availableInSrc := src.len
		availableInDst := dst.len
		currentSrcShift := srcShift
		currentDstShift := dstShift
		currentSrcIndex := 0
		currentDstIndex := 0
		for availableInSrc != 0 && availableInDst != 0 {
			srcRemainder := minUint(availableInSrc, uintSize-currentSrcShift)
			dstRemainder := minUint(availableInDst, uintSize-currentDstShift)
			copyLen = minUint(srcRemainder, dstRemainder)

			copyUintPart(
				copyLen,
				src.store[currentSrcIndex],
				currentSrcShift,
				&dst.store[currentDstIndex],
				currentDstShift,
			)

			availableInSrc -= copyLen
			availableInDst -= copyLen
			currentSrcShift += copyLen
			currentDstShift += copyLen

			if currentSrcShift == uintSize {
				currentSrcShift = 0
				currentSrcIndex++
			} else { // optimizing, using the fact of "src.shift != dst.shift" here
				currentDstShift = 0
				currentDstIndex++
			}
		}
	}

	return copyLen
}

// Selects a half-open range which includes the "from" bit, but excludes the "to" one.
func (bm *BitMask) Slice(fromBit uint, toBit uint) *BitMask {
	checkSliceBounds(fromBit, toBit, bm.len)
	if fromBit == toBit {
		// to avoid empty BitMask with shift
		return New(0)
	}

	fromStoreIndex := bm.getStoreIndex(fromBit)
	toStoreIndex := bm.getStoreIndex(toBit) + 1

	return &BitMask{
		store: bm.store[fromStoreIndex:toStoreIndex],
		shift: (fromBit + bm.shift) % uintSize,
		len:   toBit - fromBit,
	}
}

// Returns a closure function, which may be called many times to iterate
// through all bits and get their value as a bool.
// Closure will start returning nil after reaching the end.
func (bm *BitMask) Iterator() func() *bool {
	index := uint(0)
	return func() *bool {
		if index == bm.len {
			return nil
		}
		bref, m := bm.getBit(index)
		index++
		isSet := (*bref & m) != 0
		return &isSet
	}
}

// func (bm *BitMask) String() string {
// 	if bm.len == 0 {
// 		return "[0]{}"
// 	}
// 	var b strings.Builder
// 	b.WriteString(fmt.Sprintf("[%v]{", bm.len))

// 	// first word
// 	first := fmt.Sprintf(uintFormat, bm.store[0])[bm.shift:]

// 	// last word
// 	last := fmt.Sprintf(uintFormat, )

// 	storeLen := uint(len(bm.store))
// 	end := minUint(storeLen, maxStringedUints)

// 	for n, word := range bm.store[:end] {
// 		if n != 0 {
// 			b.WriteString(" ")
// 		}
// 		b.WriteString(fmt.Sprintf(uintFormat, word))
// 	}

// 	skipped := storeLen - end
// 	if skipped > 0 {
// 		b.WriteString(fmt.Sprintf(" ...(more %v values)", skipped))
// 	}

// 	b.WriteString("}")
// 	return b.String()
// }

func (bm *BitMask) getRemainderLen(bitIndex uint) uint {
	return (bm.shift + bitIndex) % uintSize
}

func (bm *BitMask) getStoreIndex(bitIndex uint) uint {
	return (bm.shift + bitIndex) / uintSize
}

func (bm *BitMask) getBit(bitIndex uint) (*uint, uint) {
	storeIndex := bm.getStoreIndex(bitIndex)
	remainderLen := bm.getRemainderLen(bitIndex)
	mask := uint(1) << remainderLen
	return &bm.store[storeIndex], mask
}

func minUint(a uint, b uint) uint {
	if a < b {
		return a
	}
	return b
}

func maxUint(a uint, b uint) uint {
	if a > b {
		return a
	}
	return b
}

func copyUintPart(len uint, src uint, srcShift uint, dst *uint, dstShift uint) {
	m := (uintMax << (uintSize - len))
	srcMask := m >> srcShift
	tmp := (src & srcMask)
	if dstShift > srcShift {
		tmp = tmp >> (dstShift - srcShift)
	} else {
		tmp = tmp << (srcShift - dstShift)
	}
	*dst |= tmp
}

func checkSliceBounds(fromIndex uint, toIndex uint, capacity uint) {
	if fromIndex > toIndex {
		panic(fmt.Sprintf("slice bounds out of range [%v:%v]", fromIndex, toIndex))
	}
	if toIndex > capacity {
		panic(fmt.Sprintf("slice bounds out of range [:%v] with capacity %v", toIndex, capacity))
	}
}

func checkBounds(len uint, index uint) {
	if index >= len {
		panic(fmt.Sprintf("index out of range [%v] with length %v", index, len))
	}
}
