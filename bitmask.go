package bitmask

import (
	"fmt"
	"strings"
)

const uintMax = ^uint(0)
const uintSize = 32 << (uintMax >> 63) // 32 or 64
const oneInBE = uint(1) << (uintSize - 1)

const maxStringedUints = 8

type BitMask struct {
	// uints are stored in BE representation
	store []uint
	// number of bits, doesn't include offset
	len uint
	// always < uintSize
	offset uint
}

func New(len uint) *BitMask {
	return &BitMask{store: make([]uint, (len+uintSize-1)/uintSize), len: len}
}

// Create a bitmask from uints, automatically reversing endianness, so that NewFromUint(1) produces the bitmask with the lowest bit set.
// Len() of the resulting bitmask will always be equal to len(values) * sizeof(uint).
func NewFromUint(values ...uint) *BitMask {
	store := make([]uint, len(values))
	for i, v := range values {
		store[i] = reverse(v)
	}
	// copy(store, values)
	return &BitMask{store: store, len: uintSize * uint(len(values))}
}

// Returns the legth of bitmask in bits. It will never be changed for the given receiver.
func (bm *BitMask) Len() uint {
	return bm.len
}

func (bm *BitMask) Set(bitIndex uint) {
	checkBounds(bm.len, bitIndex)
	bref, m := bm.getBit(bitIndex)
	*bref |= m
}

func (bm *BitMask) SetAll() {
	if bm.len == 0 {
		return
	}
	for i := 0; i < len(bm.store); i++ {
		bm.store[i] |= bm.getStoreWordMask(i)
	}
}

func (bm *BitMask) Clear(bitIndex uint) {
	checkBounds(bm.len, bitIndex)
	bref, m := bm.getBit(bitIndex)
	*bref &^= m
}

func (bm *BitMask) ClearAll() {
	if bm.len == 0 {
		return
	}
	for i := 0; i < len(bm.store); i++ {
		bm.store[i] &^= bm.getStoreWordMask(i)
	}
}

func (bm *BitMask) Toggle(bitIndex uint) {
	checkBounds(bm.len, bitIndex)
	bref, m := bm.getBit(bitIndex)
	*bref ^= m
}

func (bm *BitMask) ToggleAll() {
	if bm.len == 0 {
		return
	}
	for i := 0; i < len(bm.store); i++ {
		bm.store[i] ^= bm.getStoreWordMask(i)
	}
}

func (bm *BitMask) IsSet(bitIndex uint) bool {
	checkBounds(bm.len, bitIndex)
	bref, m := bm.getBit(bitIndex)
	return (*bref & m) != 0
}

// Copies bits from a source bit mask into a destination bit mask.
// Returns the number of bits copied, which will be the minimum of src.Len() and dst.Len().
func Copy(dst *BitMask, src *BitMask) uint {
	copyLen := minUint(src.len, dst.len)
	if copyLen == 0 {
		return 0
	}

	srcOffset := src.offset
	dstOffset := dst.offset

	if srcOffset == dstOffset {
		copyStartIndex := uint(0)
		copyEndIndex := ((copyLen - 1 + srcOffset) / uintSize) + 1
		remainderBitsN := (copyLen + srcOffset) % uintSize

		if srcOffset != 0 {
			// have to copy first uint manually
			copyUintPart(
				uintSize-srcOffset,
				src.store[copyStartIndex],
				srcOffset,
				&dst.store[copyStartIndex],
				dstOffset,
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
		currentSrcOffset := srcOffset
		currentDstOffset := dstOffset
		currentSrcIndex := 0
		currentDstIndex := 0
		for availableInSrc != 0 && availableInDst != 0 {
			srcRemainder := minUint(availableInSrc, uintSize-currentSrcOffset)
			dstRemainder := minUint(availableInDst, uintSize-currentDstOffset)
			partLen := minUint(srcRemainder, dstRemainder)

			copyUintPart(
				partLen,
				src.store[currentSrcIndex],
				currentSrcOffset,
				&dst.store[currentDstIndex],
				currentDstOffset,
			)

			availableInSrc -= partLen
			availableInDst -= partLen
			currentSrcOffset += partLen
			currentDstOffset += partLen

			if currentSrcOffset == uintSize {
				currentSrcOffset = 0
				currentSrcIndex++
			} else { // since "src.offset != dst.offset" here
				currentDstOffset = 0
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
		// to avoid empty BitMask with offset
		return New(0)
	}

	fromStoreIndex := bm.getStoreIndex(fromBit)
	toStoreIndex := bm.getStoreIndex(toBit-1) + 1

	return &BitMask{
		store:  bm.store[fromStoreIndex:toStoreIndex],
		offset: (fromBit + bm.offset) % uintSize,
		len:    toBit - fromBit,
	}
}

type BitIterator struct {
	Next func() (ok bool, isSet bool, index uint)
}

func (bm *BitMask) Iterator() BitIterator {
	index := uint(0)
	return BitIterator{
		Next: func() (bool, bool, uint) {
			if index == bm.len {
				return false, false, index
			}
			bref, m := bm.getBit(index)
			isSet := (*bref & m) != 0
			indexTmp := index
			index++
			return true, isSet, indexTmp
		},
	}
}

func (bm *BitMask) String() string {
	if bm.len == 0 {
		return "[0]{}"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("[%v]{", bm.len))

	storeLen := uint(len(bm.store))
	maxStringedWords := minUint(storeLen, maxStringedUints)
	firstSkippedIndex := maxStringedUints / uint(2)
	numSkipped := uint(maxInt(len(bm.store)-maxStringedUints, 0))

	for i := uint(0); i < maxStringedWords; i++ {
		wordIndex := i

		if numSkipped > 0 && i >= firstSkippedIndex {
			if i == firstSkippedIndex {
				b.WriteString("<more ")
				b.WriteString(fmt.Sprint(numSkipped * uintSize))
				b.WriteString(" bits> ")
			}
			wordIndex += numSkipped
		}

		word := bm.store[wordIndex]
		writeFromIndex := uint(0)
		writeToIndex := uint(uintSize)
		addSep := true
		if wordIndex == 0 {
			writeFromIndex = bm.offset
		}
		if wordIndex == storeLen-1 {
			writeToIndex = uintSize - bm.getTailLen()
			addSep = false
		}

		writeBits(&b, word, writeFromIndex, writeToIndex)

		if addSep {
			b.WriteString(" ")
		}
	}
	b.WriteString("}")
	return b.String()
}

func reverse(value uint) uint {
	// TODO compare with bits.Reverse()
	res := uint(0)
	for i := uint(0); i < uintSize; i++ {
		res |= (value & 1) << (uintSize - i - 1)
		value >>= 1
	}
	return res
}

func writeBits(b *strings.Builder, v uint, fromIndex uint, toIndex uint) {
	// write in store endianness (reversed)
	v <<= fromIndex
	i := fromIndex
	for i < toIndex {
		b.WriteByte('0' + byte((v&oneInBE)>>(uintSize-1)))
		v <<= 1
		i++
	}
}

func (bm *BitMask) getBitOffset(bitIndex uint) uint {
	return (bm.offset + bitIndex) % uintSize
}

func (bm *BitMask) getTailLen() uint {
	return (uintSize - bm.getBitOffset(bm.len)) % uintSize
}

func (bm *BitMask) getStoreIndex(bitIndex uint) uint {
	return (bm.offset + bitIndex) / uintSize
}

func (bm *BitMask) getBit(bitIndex uint) (*uint, uint) {
	storeIndex := bm.getStoreIndex(bitIndex)

	bitOffset := bm.getBitOffset(bitIndex)

	mask := oneInBE >> bitOffset
	return &bm.store[storeIndex], mask
}

func (bm *BitMask) getStoreWordMask(storeIndex int) uint {
	mask := uintMax
	if storeIndex == 0 {
		mask >>= bm.offset
	}
	if storeIndex == len(bm.store)-1 {
		dropTailLen := bm.getTailLen()
		mask = (mask >> dropTailLen) << dropTailLen
	}
	return mask
}

func minUint(a uint, b uint) uint {
	if a < b {
		return a
	}
	return b
}
func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func copyUintPart(len uint, src uint, srcOffset uint, dst *uint, dstOffset uint) {
	lenMask := uintMax << (uintSize - len)

	// read source bits
	srcMask := lenMask >> srcOffset
	bitsToCopy := (src & srcMask)

	// align bits to destination offset
	if dstOffset > srcOffset {
		bitsToCopy = bitsToCopy >> (dstOffset - srcOffset)
	} else {
		bitsToCopy = bitsToCopy << (srcOffset - dstOffset)
	}

	// clear target bits
	dstMask := lenMask >> dstOffset
	*dst &^= dstMask

	// copy
	*dst |= bitsToCopy
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
