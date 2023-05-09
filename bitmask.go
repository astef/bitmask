package bitmask

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

const uintMax = ^uint(0)
const uintSize = 32 << (uintMax >> 63) // 32 or 64
const oneInBE = uint(1) << (uintSize - 1)

const maxStringedUints = 8

// Represents a fixed-size array of 0/1 bits.
type BitMask struct {
	// uints are stored in BE representation
	store []uint
	// number of bits, doesn't include offset
	len uint
	// always < uintSize
	offset uint
}

// Creates new BitMask of specified length (number of bits). All bits will be cleared.
func New(len uint) *BitMask {
	return &BitMask{store: make([]uint, (len+uintSize-1)/uintSize), len: len}
}

// Create a bitmask from uints, automatically reversing endianness, so that NewFromUint(1) produces the bitmask with the lowest (leftmost) bit set.
// Len() of the resulting bitmask will always be equal to len(values) * sizeof(uint).
func NewFromUint(values ...uint) *BitMask {
	store := make([]uint, len(values))
	for i, v := range values {
		store[i] = reverse(v)
	}
	return &BitMask{store: store, len: uintSize * uint(len(values))}
}

// Create a bitmask from uints, without endianness reversal, more effective version of NewFromUint
func NewFromUintRaw(values ...uint) *BitMask {
	store := make([]uint, len(values))
	copy(store, values)
	return &BitMask{store: store, len: uintSize * uint(len(values))}
}

// Returns the legth of bitmask in bits. It will never be changed for the given receiver.
func (bm *BitMask) Len() uint {
	return bm.len
}

// Returns the legth of bitmask in uints. Result is always positive and multiple of size(uint), e.g.: 32, 64, 96, 128, ...
func (bm *BitMask) LenUint() int {
	return len(bm.store)
}

// Returns uint by index, reversing the endianness, so that {1000...} bitmask is represented by uint(1)
func (bm *BitMask) Uint(index int) uint {
	return reverse(bm.store[index])
}

// Returns uint without reversing, more effective version of Uint method
func (bm *BitMask) UintRaw(index int) uint {
	return bm.store[index]
}

// Sets the bit by bitIndex to 1.
func (bm *BitMask) Set(bitIndex uint) {
	checkBounds(bm.len, bitIndex)
	bref, m := bm.getBit(bitIndex)
	*bref |= m
}

// Sets all bits to 1. Use in combination with Slice to set the range of bits.
func (bm *BitMask) SetAll() {
	if bm.len == 0 {
		return
	}
	for i := 0; i < len(bm.store); i++ {
		bm.store[i] |= bm.getStoreWordMask(i)
	}
}

// Clears the bit by bitIndex (sets it to 0).
func (bm *BitMask) Clear(bitIndex uint) {
	checkBounds(bm.len, bitIndex)
	bref, m := bm.getBit(bitIndex)
	*bref &^= m
}

// Clears all bits. Use in combination with Slice to clear the range of bits.
func (bm *BitMask) ClearAll() {
	if bm.len == 0 {
		return
	}
	for i := 0; i < len(bm.store); i++ {
		bm.store[i] &^= bm.getStoreWordMask(i)
	}
}

// Reverses the value of the bit by bitIndex.
func (bm *BitMask) Toggle(bitIndex uint) {
	checkBounds(bm.len, bitIndex)
	bref, m := bm.getBit(bitIndex)
	*bref ^= m
}

// Reverses the value of all bits. Use in combination with Slice to reverse the range of bits.
func (bm *BitMask) ToggleAll() {
	if bm.len == 0 {
		return
	}
	for i := 0; i < len(bm.store); i++ {
		bm.store[i] ^= bm.getStoreWordMask(i)
	}
}

// Checks, whether the bit by bitIndex is set or cleared. Returns true if bit is set, and false if it's cleared.
func (bm *BitMask) IsSet(bitIndex uint) bool {
	checkBounds(bm.len, bitIndex)
	bref, m := bm.getBit(bitIndex)
	return (*bref & m) != 0
}

func copyFirstUintSameOffset(totalCopyLen uint, src *BitMask, dst *BitMask) {
	copyUintPart(
		minUint(totalCopyLen, uintSize-src.offset),
		src.store[0],
		src.offset,
		&dst.store[0],
		dst.offset,
	)
}

func copyLastUintSameOffset(totalCopyLen uint, lastUintIndex uint, src *BitMask, dst *BitMask) {
	remainderBitsN := (totalCopyLen + src.offset) % uintSize
	copyUintPart(
		minUint(totalCopyLen, remainderBitsN),
		src.store[lastUintIndex],
		0,
		&dst.store[lastUintIndex],
		0,
	)
}

// Copies bits from a source bit mask into a destination bit mask.
// It's safe to copy overlapping bitmasks (which were created by slicing the original one).
// Returns the number of bits copied, which will be the minimum of src.Len() and dst.Len().
func Copy(dst *BitMask, src *BitMask) uint {
	// how many bits to copy
	copyLen := minUint(src.len, dst.len)
	if copyLen == 0 {
		return 0
	}

	// in what direction to copy (to handle overlapping data)
	srcDataPtr := (*reflect.SliceHeader)(unsafe.Pointer(&src.store)).Data
	dstDataPtr := (*reflect.SliceHeader)(unsafe.Pointer(&dst.store)).Data
	fwdDirection := false
	if dstDataPtr < srcDataPtr {
		fwdDirection = true
	}

	srcOffset := src.offset
	dstOffset := dst.offset

	if srcOffset == dstOffset {
		copyEndIndex := ((srcOffset + copyLen - 1) / uintSize) + 1
		bulkCopyStartIndex := uint(1)
		bulkCopyEndIndex := copyEndIndex - 1

		if fwdDirection {
			copyFirstUintSameOffset(copyLen, src, dst)
			if bulkCopyStartIndex < bulkCopyEndIndex {
				copy(dst.store[bulkCopyStartIndex:bulkCopyEndIndex], src.store[bulkCopyStartIndex:bulkCopyEndIndex])
			}
			if copyEndIndex > 1 {
				copyLastUintSameOffset(copyLen, copyEndIndex-1, src, dst)
			}
		} else {
			if copyEndIndex > 1 {
				copyLastUintSameOffset(copyLen, copyEndIndex-1, src, dst)
			}
			if bulkCopyStartIndex < bulkCopyEndIndex {
				copy(dst.store[bulkCopyStartIndex:bulkCopyEndIndex], src.store[bulkCopyStartIndex:bulkCopyEndIndex])
			}
			copyFirstUintSameOffset(copyLen, src, dst)
		}
	} else {
		// not optimized copy, only by uint parts
		availableInSrc := src.len
		availableInDst := dst.len

		var currentSrcCursor, currentDstCursor uint
		if fwdDirection {
			currentSrcCursor = srcOffset
			currentDstCursor = dstOffset
		} else {
			currentSrcCursor = (srcOffset + copyLen) % uintSize
			if currentSrcCursor == 0 {
				currentSrcCursor = uintSize
			}
			currentDstCursor = (dstOffset + copyLen) % uintSize
			if currentDstCursor == 0 {
				currentDstCursor = uintSize
			}
		}

		var currentSrcIndex, currentDstIndex uint
		if !fwdDirection {
			currentSrcIndex = (srcOffset + copyLen - 1) / uintSize
			currentDstIndex = (dstOffset + copyLen - 1) / uintSize
		}

		for availableInSrc != 0 && availableInDst != 0 {
			// length of the copied part
			var srcPartLen, dstPartLen uint
			if fwdDirection {
				srcPartLen = minUint(availableInSrc, uintSize-currentSrcCursor)
				dstPartLen = minUint(availableInDst, uintSize-currentDstCursor)
			} else {
				srcPartLen = minUint(availableInSrc, currentSrcCursor)
				dstPartLen = minUint(availableInDst, currentDstCursor)
			}
			partLen := minUint(srcPartLen, dstPartLen)

			// copy
			srcPartOffset := currentSrcCursor
			dstPartOffset := currentDstCursor
			if !fwdDirection {
				srcPartOffset = currentSrcCursor - partLen
				dstPartOffset = currentDstCursor - partLen
			}
			copyUintPart(
				partLen,
				src.store[currentSrcIndex],
				srcPartOffset,
				&dst.store[currentDstIndex],
				dstPartOffset,
			)

			// update variables
			availableInSrc -= partLen
			availableInDst -= partLen

			if fwdDirection {
				currentSrcCursor += partLen
				currentDstCursor += partLen
				if currentSrcCursor == uintSize {
					currentSrcCursor = 0
					currentSrcIndex++
				} else { // since "src.offset != dst.offset" here
					currentDstCursor = 0
					currentDstIndex++
				}
			} else {
				currentSrcCursor -= partLen
				currentDstCursor -= partLen
				if currentSrcCursor == 0 {
					currentSrcCursor = uintSize
					currentSrcIndex--
				} else { // since "src.offset != dst.offset" here
					currentDstCursor = uintSize
					currentDstIndex--
				}
			}
		}
	}

	return copyLen
}

// Effectively creates a new BitMask, without copying elements, just like regular slices work.
// As a side-effect, change to the sliced bitmask will be visible to original bitmask (and other way around),
// as well as to other "overlapping" slices.
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

// Stateful iterator.
// Example of usage:
//
//	it := bm.Iterator()
//	for {
//		ok, value, index := it.Next()
//		if !ok {
//			break;
//		}
//		// use the value
//	}
type BitIterator struct {
	// Atttempts to get the next item from the iterator.
	// If there're no more values left, ok will be false.
	Next func() (ok bool, isSet bool, index uint)

	// Resets iterator, so it can be reused
	Reset func()
}

// Creates stateful iterator to iterate through all the bits.
// See BitIterator doc for an example.
// It is an equivalent of just calling IsSet for each bit of a BitMask.
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
		Reset: func() {
			index = 0
		},
	}
}

// Returns string representation of a bitmask in the form "[length]{bits}".
// For example: [4]{0100}
// It is O(1) operation and it will skip bits after some amount of them.
// Should not be parsed, format is not fixed.
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
