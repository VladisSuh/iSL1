package des

import (
	"errors"
	"iSL1/customlib"
)

var (
	initialPermutation = []int{
		58, 50, 42, 34, 26, 18, 10, 2,
		60, 52, 44, 36, 28, 20, 12, 4,
		62, 54, 46, 38, 30, 22, 14, 6,
		64, 56, 48, 40, 32, 24, 16, 8,
		57, 49, 41, 33, 25, 17, 9, 1,
		59, 51, 43, 35, 27, 19, 11, 3,
		61, 53, 45, 37, 29, 21, 13, 5,
		63, 55, 47, 39, 31, 23, 15, 7,
	}

	finalPermutation = []int{
		40, 8, 48, 16, 56, 24, 64, 32,
		39, 7, 47, 15, 55, 23, 63, 31,
		38, 6, 46, 14, 54, 22, 62, 30,
		37, 5, 45, 13, 53, 21, 61, 29,
		36, 4, 44, 12, 52, 20, 60, 28,
		35, 3, 43, 11, 51, 19, 59, 27,
		34, 2, 42, 10, 50, 18, 58, 26,
		33, 1, 41, 9, 49, 17, 57, 25,
	}

	expansionPermutation = []int{
		32, 1, 2, 3, 4, 5,
		4, 5, 6, 7, 8, 9,
		8, 9, 10, 11, 12, 13,
		12, 13, 14, 15, 16, 17,
		16, 17, 18, 19, 20, 21,
		20, 21, 22, 23, 24, 25,
		24, 25, 26, 27, 28, 29,
		28, 29, 30, 31, 32, 1,
	}

	permutationP = []int{
		16, 7, 20, 21,
		29, 12, 28, 17,
		1, 15, 23, 26,
		5, 18, 31, 10,
		2, 8, 24, 14,
		32, 27, 3, 9,
		19, 13, 30, 6,
		22, 11, 4, 25,
	}

	sBoxes = [8][4][16]int{
		{
			{14, 4, 13, 1, 2, 15, 11, 8, 3, 10, 6, 12, 5, 9, 0, 7},
			{0, 15, 7, 4, 14, 2, 13, 1, 10, 6, 12, 11, 9, 5, 3, 8},
			{4, 1, 14, 8, 13, 6, 2, 11, 15, 12, 9, 7, 3, 10, 5, 0},
			{15, 12, 8, 2, 4, 9, 1, 7, 5, 11, 3, 14, 10, 0, 6, 13},
		},
		{
			{15, 1, 8, 14, 6, 11, 3, 4, 9, 7, 2, 13, 12, 0, 5, 10},
			{3, 13, 4, 7, 15, 2, 8, 14, 12, 0, 1, 10, 6, 9, 11, 5},
			{0, 14, 7, 11, 10, 4, 13, 1, 5, 8, 12, 6, 9, 3, 2, 15},
			{13, 8, 10, 1, 3, 15, 4, 2, 11, 6, 7, 12, 0, 5, 14, 9},
		},
		{
			{10, 0, 9, 14, 6, 3, 15, 5, 1, 13, 12, 7, 11, 4, 2, 8},
			{13, 7, 0, 9, 3, 4, 6, 10, 2, 8, 5, 14, 12, 11, 15, 1},
			{13, 6, 4, 9, 8, 15, 3, 0, 11, 1, 2, 12, 5, 10, 14, 7},
			{1, 10, 13, 0, 6, 9, 8, 7, 4, 15, 14, 3, 11, 5, 2, 12},
		},
		{
			{7, 13, 14, 3, 0, 6, 9, 10, 1, 2, 8, 5, 11, 12, 4, 15},
			{13, 8, 11, 5, 6, 15, 0, 3, 4, 7, 2, 12, 1, 10, 14, 9},
			{10, 6, 9, 0, 12, 11, 7, 13, 15, 1, 3, 14, 5, 2, 8, 4},
			{3, 15, 0, 6, 10, 1, 13, 8, 9, 4, 5, 11, 12, 7, 2, 14},
		},
		{
			{2, 12, 4, 1, 7, 10, 11, 6, 8, 5, 3, 15, 13, 0, 14, 9},
			{14, 11, 2, 12, 4, 7, 13, 1, 5, 0, 15, 10, 3, 9, 8, 6},
			{4, 2, 1, 11, 10, 13, 7, 8, 15, 9, 12, 5, 6, 3, 0, 14},
			{11, 8, 12, 7, 1, 14, 2, 13, 6, 15, 0, 9, 10, 4, 5, 3},
		},
		{
			{12, 1, 10, 15, 9, 2, 6, 8, 0, 13, 3, 4, 14, 7, 5, 11},
			{10, 15, 4, 2, 7, 12, 9, 5, 6, 1, 13, 14, 0, 11, 3, 8},
			{9, 14, 15, 5, 2, 8, 12, 3, 7, 0, 4, 10, 1, 13, 11, 6},
			{4, 3, 2, 12, 9, 5, 15, 10, 11, 14, 1, 7, 6, 0, 8, 13},
		},
		{
			{4, 11, 2, 14, 15, 0, 8, 13, 3, 12, 9, 7, 5, 10, 6, 1},
			{13, 0, 11, 7, 4, 9, 1, 10, 14, 3, 5, 12, 2, 15, 8, 6},
			{1, 4, 11, 13, 12, 3, 7, 14, 10, 15, 6, 8, 0, 5, 9, 2},
			{6, 11, 13, 8, 1, 4, 10, 7, 9, 5, 0, 15, 14, 2, 3, 12},
		},
		{
			{13, 2, 8, 4, 6, 15, 11, 1, 10, 9, 3, 14, 5, 0, 12, 7},
			{1, 15, 13, 8, 10, 3, 7, 4, 12, 5, 6, 11, 0, 14, 9, 2},
			{7, 11, 4, 1, 9, 12, 14, 2, 0, 6, 10, 13, 15, 3, 5, 8},
			{2, 1, 14, 7, 4, 10, 8, 13, 15, 12, 9, 0, 3, 5, 6, 11},
		},
	}

	pc1 = []int{
		57, 49, 41, 33, 25, 17, 9,
		1, 58, 50, 42, 34, 26, 18,
		10, 2, 59, 51, 43, 35, 27,
		19, 11, 3, 60, 52, 44, 36,
		63, 55, 47, 39, 31, 23, 15,
		7, 62, 54, 46, 38, 30, 22,
		14, 6, 61, 53, 45, 37, 29,
		21, 13, 5, 28, 20, 12, 4,
	}

	pc2 = []int{
		14, 17, 11, 24, 1, 5,
		3, 28, 15, 6, 21, 10,
		23, 19, 12, 4, 26, 8,
		16, 7, 27, 20, 13, 2,
		41, 52, 31, 37, 47, 55,
		30, 40, 51, 45, 33, 48,
		44, 49, 39, 56, 34, 53,
		46, 42, 50, 36, 29, 32,
	}

	keyShifts = []int{
		1, 1, 2, 2, 2, 2, 2, 2,
		1, 2, 2, 2, 2, 2, 2, 1,
	}
)

func leftShift(bits []byte, shift int) []byte {
	bitLen := len(bits) * 8
	totalShifts := shift % bitLen
	result := make([]byte, len(bits))
	for i := 0; i < bitLen; i++ {
		fromIndex := (i + totalShifts) % bitLen
		bit := getBit(bits, fromIndex)
		setBit(result, i, bit)
	}
	return result
}

func getBit(data []byte, index int) byte {
	byteIndex := index / 8
	bitOffset := 7 - (index % 8)
	return (data[byteIndex] >> bitOffset) & 1
}

func setBit(data []byte, index int, value byte) {
	byteIndex := index / 8
	bitOffset := 7 - (index % 8)
	if value == 1 {
		data[byteIndex] |= (1 << bitOffset)
	} else {
		data[byteIndex] &^= (1 << bitOffset)
	}
}

func permuteBits(value []byte, pBlock []int) []byte {
	outputBitsLen := len(pBlock)
	outputBytesLen := (outputBitsLen + 7) / 8
	output := make([]byte, outputBytesLen)

	for i, p := range pBlock {
		inputBitIndex := p - 1
		bit := getBit(value, inputBitIndex)
		setBit(output, i, bit)
	}

	return output
}

type DESKeyExpander struct{}

func (ke *DESKeyExpander) ExpandKey(key []byte) ([][]byte, error) {
	if len(key) != 8 {
		return nil, errors.New("ключ должен быть длиной 8 байт")
	}

	key56 := permuteBits(key, pc1)

	c := key56[:len(key56)/2]
	d := key56[len(key56)/2:]

	roundKeys := make([][]byte, 16)
	for i := 0; i < 16; i++ {
		c = leftShift(c, keyShifts[i])
		d = leftShift(d, keyShifts[i])

		cd := append(c, d...)
		roundKey := permuteBits(cd, pc2)
		roundKeys[i] = roundKey
	}

	return roundKeys, nil
}

type DESCipherTransformation struct{}

func (ct *DESCipherTransformation) EncryptBlock(inputBlock []byte, roundKey []byte) ([]byte, error) {
	expandedBlock := permuteBits(inputBlock, expansionPermutation)

	for i := 0; i < len(expandedBlock); i++ {
		expandedBlock[i] ^= roundKey[i]
	}

	sBoxOutput := make([]byte, 4)
	for i := 0; i < 8; i++ {
		offset := i * 6

		b1 := getBit(expandedBlock, offset)
		b2 := getBit(expandedBlock, offset+1)
		b3 := getBit(expandedBlock, offset+2)
		b4 := getBit(expandedBlock, offset+3)
		b5 := getBit(expandedBlock, offset+4)
		b6 := getBit(expandedBlock, offset+5)

		row := (b1 << 1) | b6
		column := (b2 << 3) | (b3 << 2) | (b4 << 1) | b5

		value := byte(sBoxes[i][row][column])

		setBits(sBoxOutput, i*4, 4, value)
	}

	pBlock := permuteBits(sBoxOutput, permutationP)

	return pBlock, nil
}

func (ct *DESCipherTransformation) DecryptBlock(inputBlock []byte, roundKey []byte) ([]byte, error) {
	return ct.EncryptBlock(inputBlock, roundKey)
}

func getBits(data []byte, offset int, length int) []byte {
	result := make([]byte, (length+7)/8)
	for i := 0; i < length; i++ {
		bit := getBit(data, offset+i)
		setBit(result, i, bit)
	}
	return result
}

func setBits(data []byte, offset int, length int, value byte) {
	for i := 0; i < length; i++ {
		bit := (value >> (length - i - 1)) & 1
		setBit(data, offset+i, bit)
	}
}

type DES struct {
	feistelCipher *customlib.FeistelCipher
}

func NewDES() (*DES, error) {
	keyExpander := &DESKeyExpander{}
	cipherFunc := &DESCipherTransformation{}
	blockSize := 8
	numRounds := 16

	feistelCipher, err := customlib.NewFeistelCipher(keyExpander, cipherFunc, numRounds, blockSize)
	if err != nil {
		return nil, err
	}

	return &DES{
		feistelCipher: feistelCipher,
	}, nil
}

func (des *DES) SetKey(key []byte) error {
	return des.feistelCipher.SetKey(key)
}

func (des *DES) EncryptBlock(block []byte) ([]byte, error) {
	if len(block) != 8 {
		return nil, errors.New("блок должен быть длиной 8 байт")
	}

	block = permuteBits(block, initialPermutation)

	block, err := des.feistelCipher.EncryptBlock(block)
	if err != nil {
		return nil, err
	}

	block = permuteBits(block, finalPermutation)

	return block, nil
}

func (des *DES) DecryptBlock(block []byte) ([]byte, error) {
	if len(block) != 8 {
		return nil, errors.New("блок должен быть длиной 8 байт")
	}

	block = permuteBits(block, initialPermutation)

	block, err := des.feistelCipher.DecryptBlock(block)
	if err != nil {
		return nil, err
	}

	block = permuteBits(block, finalPermutation)

	return block, nil
}
