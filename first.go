package main

import (
	"fmt"
)

func GetBit(value []byte, bitIndex int, indexFromLSB bool) int {
	byteIndex := bitIndex / 8
	bitOffset := bitIndex % 8

	if indexFromLSB {
		return int((value[byteIndex] >> bitOffset) & 1)
	} else {
		return int((value[byteIndex] >> (7 - bitOffset)) & 1)
	}
}

func SetBit(value []byte, bitIndex int, bit int, indexFromLSB bool) {
	byteIndex := bitIndex / 8
	bitOffset := bitIndex % 8

	if indexFromLSB {
		if bit == 1 {
			value[byteIndex] |= (1 << bitOffset)
		} else {
			value[byteIndex] &^= (1 << bitOffset)
		}
	} else {
		if bit == 1 {
			value[byteIndex] |= (1 << (7 - bitOffset))
		} else {
			value[byteIndex] &^= (1 << (7 - bitOffset))
		}
	}
}

func PermuteBits(value []byte, pBlock []int, indexFromLSB bool, startBitNumber int) []byte {
	outputBitsLen := len(pBlock)
	outputBytesLen := (outputBitsLen + 7) / 8
	output := make([]byte, outputBytesLen)

	for i, p := range pBlock {
		inputBitIndex := p - startBitNumber

		bit := GetBit(value, inputBitIndex, indexFromLSB)

		SetBit(output, i, bit, indexFromLSB)
	}

	return output
}

func main() {
	value := []byte{0b11001100, 0b10101010}
	pBlock := []int{8, 7, 6, 5, 4, 3, 2, 1, 16, 15, 14, 13, 12, 11, 10, 9}
	indexFromLSB := true
	startBitNumber := 1

	result := PermuteBits(value, pBlock, indexFromLSB, startBitNumber)

	fmt.Printf("Результат перестановки: %08b %08b\n", result[0], result[1])
}
