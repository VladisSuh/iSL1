package customlib

import (
	"errors"
)

type FeistelCipher struct {
	keyExpander KeyExpander
	cipherFunc  CipherTransformation
	roundKeys   [][]byte
	numRounds   int
	blockSize   int
}

func NewFeistelCipher(keyExpander KeyExpander, cipherFunc CipherTransformation, numRounds int, blockSize int) (*FeistelCipher, error) {
	if keyExpander == nil || cipherFunc == nil {
		return nil, errors.New("keyExpander и cipherFunc не могут быть nil")
	}
	if numRounds <= 0 {
		return nil, errors.New("число раундов должно быть положительным")
	}
	if blockSize <= 0 || blockSize%2 != 0 {
		return nil, errors.New("размер блока должен быть положительным и чётным")
	}

	return &FeistelCipher{
		keyExpander: keyExpander,
		cipherFunc:  cipherFunc,
		numRounds:   numRounds,
		blockSize:   blockSize,
	}, nil
}

func (fc *FeistelCipher) SetKey(key []byte) error {
	roundKeys, err := fc.keyExpander.ExpandKey(key)
	if err != nil {
		return err
	}
	if len(roundKeys) != fc.numRounds {
		return errors.New("число раундовых ключей не соответствует числу раундов")
	}
	fc.roundKeys = roundKeys
	return nil
}

func (fc *FeistelCipher) EncryptBlock(block []byte) ([]byte, error) {
	if len(block) != fc.blockSize {
		return nil, errors.New("неверный размер блока")
	}

	left := make([]byte, fc.blockSize/2)
	right := make([]byte, fc.blockSize/2)
	copy(left, block[:fc.blockSize/2])
	copy(right, block[fc.blockSize/2:])

	for i := 0; i < fc.numRounds; i++ {
		fResult, err := fc.cipherFunc.EncryptBlock(right, fc.roundKeys[i])
		if err != nil {
			return nil, err
		}

		newRight := xorBytes(left, fResult)

		left = right
		right = newRight
	}

	result := append(left, right...)
	return result, nil
}

func (fc *FeistelCipher) DecryptBlock(block []byte) ([]byte, error) {
	if len(block) != fc.blockSize {
		return nil, errors.New("неверный размер блока")
	}

	left := make([]byte, fc.blockSize/2)
	right := make([]byte, fc.blockSize/2)
	copy(left, block[:fc.blockSize/2])
	copy(right, block[fc.blockSize/2:])

	for i := fc.numRounds - 1; i >= 0; i-- {
		fResult, err := fc.cipherFunc.EncryptBlock(left, fc.roundKeys[i])
		if err != nil {
			return nil, err
		}

		newLeft := xorBytes(right, fResult)

		right = left
		left = newLeft
	}

	result := append(left, right...)
	return result, nil
}

func xorBytes(a, b []byte) []byte {
	length := len(a)
	if len(b) < length {
		length = len(b)
	}
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = a[i] ^ b[i]
	}
	return result
}
