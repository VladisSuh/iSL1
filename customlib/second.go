package customlib

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"errors"
	"os"
	"sync"
)

type CipherMode int

const (
	ModeECB CipherMode = iota
	ModeCBC
	ModePCBC
	ModeCFB
	ModeOFB
	ModeCTR
	ModeRandomDelta
)

type PaddingMode int

const (
	PaddingZeros PaddingMode = iota
	PaddingANSIX923
	PaddingPKCS7
	PaddingISO10126
)

type KeyExpander interface {
	ExpandKey(key []byte) ([][]byte, error)
}

type CipherTransformation interface {
	EncryptBlock(inputBlock []byte, roundKey []byte) ([]byte, error)
	DecryptBlock(inputBlock []byte, roundKey []byte) ([]byte, error)
}

type SymmetricCipher interface {
	SetKey(key []byte) error
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

type CryptoContext struct {
	key         []byte
	roundKeys   [][]byte
	mode        CipherMode
	padding     PaddingMode
	iv          []byte
	transform   CipherTransformation
	expander    KeyExpander
	extraParams map[string]interface{}
}

func NewCryptoContext(key []byte, mode CipherMode, padding PaddingMode, iv []byte, transform CipherTransformation, expander KeyExpander) (*CryptoContext, error) {
	ctx := &CryptoContext{
		key:         key,
		mode:        mode,
		padding:     padding,
		iv:          iv,
		transform:   transform,
		expander:    expander,
		extraParams: make(map[string]interface{}),
	}

	err := ctx.SetKey(key)
	if err != nil {
		return nil, err
	}

	return ctx, nil
}

func (ctx *CryptoContext) SetKey(key []byte) error {
	if ctx.expander == nil {
		return errors.New("KeyExpander не инициализирован")
	}

	roundKeys, err := ctx.expander.ExpandKey(key)
	if err != nil {
		return err
	}

	ctx.roundKeys = roundKeys
	return nil
}

func (ctx *CryptoContext) applyPadding(data []byte) ([]byte, error) {
	blockSize := len(ctx.roundKeys[0])
	paddingNeeded := blockSize - (len(data) % blockSize)
	if paddingNeeded == 0 {
		paddingNeeded = blockSize
	}

	switch ctx.padding {
	case PaddingZeros:
		return applyZerosPadding(data, paddingNeeded), nil
	case PaddingANSIX923:
		return applyANSIX923Padding(data, paddingNeeded, blockSize)
	case PaddingPKCS7:
		return applyPKCS7Padding(data, paddingNeeded), nil
	case PaddingISO10126:
		return applyISO10126Padding(data, paddingNeeded)
	default:
		return nil, errors.New("неподдерживаемый режим набивки")
	}
}

func (ctx *CryptoContext) removePadding(data []byte) ([]byte, error) {
	switch ctx.padding {
	case PaddingZeros:
		return removeZerosPadding(data), nil
	case PaddingANSIX923:
		return removeANSIX923Padding(data)
	case PaddingPKCS7:
		return removePKCS7Padding(data)
	case PaddingISO10126:
		return removeANSIX923Padding(data)
	default:
		return nil, errors.New("неподдерживаемый режим набивки")
	}
}

func applyZerosPadding(data []byte, paddingNeeded int) []byte {
	padding := bytes.Repeat([]byte{0}, paddingNeeded)
	return append(data, padding...)
}

func removeZerosPadding(data []byte) []byte {
	return bytes.TrimRight(data, "\x00")
}

func applyANSIX923Padding(data []byte, paddingNeeded int, blockSize int) ([]byte, error) {
	padding := make([]byte, paddingNeeded)
	padding[paddingNeeded-1] = byte(paddingNeeded)
	return append(data, padding...), nil
}

func removeANSIX923Padding(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("данные пусты")
	}
	paddingLength := int(data[len(data)-1])
	if paddingLength > len(data) {
		return nil, errors.New("неверная длина набивки")
	}
	return data[:len(data)-paddingLength], nil
}

func applyPKCS7Padding(data []byte, paddingNeeded int) []byte {
	padding := bytes.Repeat([]byte{byte(paddingNeeded)}, paddingNeeded)
	return append(data, padding...)
}

func removePKCS7Padding(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("данные пусты")
	}
	paddingLength := int(data[len(data)-1])
	if paddingLength > len(data) || paddingLength == 0 {
		return nil, errors.New("неверная длина набивки")
	}
	for i := 1; i <= paddingLength; i++ {
		if data[len(data)-i] != byte(paddingLength) {
			return nil, errors.New("неверные данные набивки")
		}
	}
	return data[:len(data)-paddingLength], nil
}

func applyISO10126Padding(data []byte, blockSize int) ([]byte, error) {
	paddingNeeded := blockSize - (len(data) % blockSize)
	if paddingNeeded == 0 {
		paddingNeeded = blockSize
	}
	padding := make([]byte, paddingNeeded)
	_, err := rand.Read(padding[:paddingNeeded-1])
	if err != nil {
		return nil, err
	}
	padding[paddingNeeded-1] = byte(paddingNeeded)
	return append(data, padding...), nil
}

func (ctx *CryptoContext) Encrypt(data []byte) ([]byte, error) {
	dataWithPadding, err := ctx.applyPadding(data)
	if err != nil {
		return nil, err
	}

	switch ctx.mode {
	case ModeECB:
		return ctx.encryptECB(dataWithPadding)
	case ModeCBC:
		return ctx.encryptCBC(dataWithPadding)
	case ModePCBC:
		return ctx.encryptPCBC(dataWithPadding)
	case ModeCFB:
		return ctx.encryptCFB(dataWithPadding)
	case ModeOFB:
		return ctx.encryptOFB(dataWithPadding)
	case ModeCTR:
		return ctx.encryptCTR(dataWithPadding)
	case ModeRandomDelta:
		return ctx.encryptRandomDelta(dataWithPadding)
	default:
		return nil, errors.New("неподдерживаемый режим шифрования")
	}
}

func (ctx *CryptoContext) Decrypt(data []byte) ([]byte, error) {
	var decryptedData []byte
	var err error

	switch ctx.mode {
	case ModeECB:
		decryptedData, err = ctx.decryptECB(data)
	case ModeCBC:
		decryptedData, err = ctx.decryptCBC(data)
	case ModePCBC:
		decryptedData, err = ctx.decryptPCBC(data)
	case ModeCFB:
		decryptedData, err = ctx.decryptCFB(data)
	case ModeOFB:
		decryptedData, err = ctx.decryptOFB(data)
	case ModeCTR:
		decryptedData, err = ctx.decryptCTR(data)
	case ModeRandomDelta:
		decryptedData, err = ctx.decryptRandomDelta(data)
	default:
		return nil, errors.New("неподдерживаемый режим шифрования")
	}

	if err != nil {
		return nil, err
	}

	decryptedData, err = ctx.removePadding(decryptedData)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}

func (ctx *CryptoContext) encryptECB(data []byte) ([]byte, error) {
	blockSize := len(ctx.roundKeys[0])
	if len(data)%blockSize != 0 {
		return nil, errors.New("данные не кратны размеру блока")
	}

	result := make([]byte, len(data))

	ch := make(chan error)
	for i := 0; i < len(data); i += blockSize {
		go func(start int) {
			block := data[start : start+blockSize]
			encryptedBlock := block
			var err error

			for _, roundKey := range ctx.roundKeys {
				encryptedBlock, err = ctx.transform.EncryptBlock(encryptedBlock, roundKey)
				if err != nil {
					ch <- err
					return
				}
			}

			copy(result[start:start+blockSize], encryptedBlock)
			ch <- nil
		}(i)
	}

	for i := 0; i < len(data)/blockSize; i++ {
		if err := <-ch; err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (ctx *CryptoContext) decryptECB(data []byte) ([]byte, error) {
	blockSize := len(ctx.roundKeys[0])
	if len(data)%blockSize != 0 {
		return nil, errors.New("данные не кратны размеру блока")
	}

	result := make([]byte, len(data))

	ch := make(chan error)
	for i := 0; i < len(data); i += blockSize {
		go func(start int) {
			block := data[start : start+blockSize]
			decryptedBlock := block
			var err error

			for idx := len(ctx.roundKeys) - 1; idx >= 0; idx-- {
				roundKey := ctx.roundKeys[idx]
				decryptedBlock, err = ctx.transform.DecryptBlock(decryptedBlock, roundKey)
				if err != nil {
					ch <- err
					return
				}
			}

			copy(result[start:start+blockSize], decryptedBlock)
			ch <- nil
		}(i)
	}

	for i := 0; i < len(data)/blockSize; i++ {
		if err := <-ch; err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (ctx *CryptoContext) encryptCBC(data []byte) ([]byte, error) {
	blockSize := len(ctx.roundKeys[0])
	if len(data)%blockSize != 0 {
		return nil, errors.New("данные не кратны размеру блока")
	}

	if ctx.iv == nil || len(ctx.iv) != blockSize {
		return nil, errors.New("неверный вектор инициализации (IV)")
	}

	result := make([]byte, len(data))
	prevBlock := ctx.iv

	for i := 0; i < len(data); i += blockSize {
		block := xorBlocks(data[i:i+blockSize], prevBlock)

		encryptedBlock, err := ctx.encryptBlock(block)
		if err != nil {
			return nil, err
		}

		copy(result[i:i+blockSize], encryptedBlock)
		prevBlock = encryptedBlock
	}

	return result, nil
}

func (ctx *CryptoContext) decryptCBC(data []byte) ([]byte, error) {
	blockSize := len(ctx.roundKeys[0])
	if len(data)%blockSize != 0 {
		return nil, errors.New("данные не кратны размеру блока")
	}

	if ctx.iv == nil || len(ctx.iv) != blockSize {
		return nil, errors.New("неверный вектор инициализации (IV)")
	}

	result := make([]byte, len(data))
	prevBlock := ctx.iv

	for i := 0; i < len(data); i += blockSize {
		encryptedBlock := data[i : i+blockSize]

		decryptedBlock, err := ctx.decryptBlock(encryptedBlock)
		if err != nil {
			return nil, err
		}

		block := xorBlocks(decryptedBlock, prevBlock)

		copy(result[i:i+blockSize], block)
		prevBlock = encryptedBlock
	}

	return result, nil
}

func (ctx *CryptoContext) encryptPCBC(data []byte) ([]byte, error) {
	blockSize := len(ctx.roundKeys[0])
	if len(data)%blockSize != 0 {
		return nil, errors.New("данные не кратны размеру блока")
	}

	if ctx.iv == nil || len(ctx.iv) != blockSize {
		return nil, errors.New("неверный вектор инициализации (IV)")
	}

	result := make([]byte, len(data))
	prevPlainBlock := ctx.iv
	prevCipherBlock := ctx.iv

	for i := 0; i < len(data); i += blockSize {
		plainBlock := data[i : i+blockSize]

		block := xorBlocks(plainBlock, xorBlocks(prevPlainBlock, prevCipherBlock))

		encryptedBlock, err := ctx.encryptBlock(block)
		if err != nil {
			return nil, err
		}

		copy(result[i:i+blockSize], encryptedBlock)

		prevPlainBlock = plainBlock
		prevCipherBlock = encryptedBlock
	}

	return result, nil
}

func (ctx *CryptoContext) decryptPCBC(data []byte) ([]byte, error) {
	blockSize := len(ctx.roundKeys[0])
	if len(data)%blockSize != 0 {
		return nil, errors.New("данные не кратны размеру блока")
	}

	if ctx.iv == nil || len(ctx.iv) != blockSize {
		return nil, errors.New("неверный вектор инициализации (IV)")
	}

	result := make([]byte, len(data))
	prevPlainBlock := ctx.iv
	prevCipherBlock := ctx.iv

	for i := 0; i < len(data); i += blockSize {
		cipherBlock := data[i : i+blockSize]

		decryptedBlock, err := ctx.decryptBlock(cipherBlock)
		if err != nil {
			return nil, err
		}

		plainBlock := xorBlocks(decryptedBlock, xorBlocks(prevPlainBlock, prevCipherBlock))

		copy(result[i:i+blockSize], plainBlock)

		prevPlainBlock = plainBlock
		prevCipherBlock = cipherBlock
	}

	return result, nil
}

func (ctx *CryptoContext) encryptCFB(data []byte) ([]byte, error) {
	blockSize := len(ctx.roundKeys[0])

	if ctx.iv == nil || len(ctx.iv) != blockSize {
		return nil, errors.New("неверный вектор инициализации (IV)")
	}

	result := make([]byte, len(data))
	stream := ctx.iv

	for i := 0; i < len(data); i += blockSize {
		encryptedStream, err := ctx.encryptBlock(stream)
		if err != nil {
			return nil, err
		}

		blockSizeToUse := min(blockSize, len(data)-i)
		cipherBlock := xorBlocks(data[i:i+blockSizeToUse], encryptedStream[:blockSizeToUse])

		copy(result[i:i+blockSizeToUse], cipherBlock)

		stream = append(stream[blockSizeToUse:], cipherBlock...)
		if len(stream) > blockSize {
			stream = stream[:blockSize]
		}
	}

	return result, nil
}

func (ctx *CryptoContext) decryptCFB(data []byte) ([]byte, error) {
	blockSize := len(ctx.roundKeys[0])

	if ctx.iv == nil || len(ctx.iv) != blockSize {
		return nil, errors.New("неверный вектор инициализации (IV)")
	}

	result := make([]byte, len(data))
	stream := ctx.iv

	for i := 0; i < len(data); i += blockSize {
		encryptedStream, err := ctx.encryptBlock(stream)
		if err != nil {
			return nil, err
		}

		blockSizeToUse := min(blockSize, len(data)-i)
		plainBlock := xorBlocks(data[i:i+blockSizeToUse], encryptedStream[:blockSizeToUse])

		copy(result[i:i+blockSizeToUse], plainBlock)

		stream = append(stream[blockSizeToUse:], data[i:i+blockSizeToUse]...)
		if len(stream) > blockSize {
			stream = stream[:blockSize]
		}
	}

	return result, nil
}

func (ctx *CryptoContext) encryptOFB(data []byte) ([]byte, error) {
	return ctx.processOFB(data)
}

func (ctx *CryptoContext) decryptOFB(data []byte) ([]byte, error) {
	return ctx.processOFB(data)
}

func (ctx *CryptoContext) processOFB(data []byte) ([]byte, error) {
	blockSize := len(ctx.roundKeys[0])

	if ctx.iv == nil || len(ctx.iv) != blockSize {
		return nil, errors.New("неверный вектор инициализации (IV)")
	}

	result := make([]byte, len(data))
	stream := ctx.iv

	for i := 0; i < len(data); i += blockSize {
		encryptedStream, err := ctx.encryptBlock(stream)
		if err != nil {
			return nil, err
		}

		blockSizeToUse := min(blockSize, len(data)-i)
		outputBlock := xorBlocks(data[i:i+blockSizeToUse], encryptedStream[:blockSizeToUse])

		copy(result[i:i+blockSizeToUse], outputBlock)

		stream = encryptedStream
	}

	return result, nil
}

func (ctx *CryptoContext) encryptCTR(data []byte) ([]byte, error) {
	return ctx.processCTR(data)
}

func (ctx *CryptoContext) decryptCTR(data []byte) ([]byte, error) {
	return ctx.processCTR(data)
}

func (ctx *CryptoContext) processCTR(data []byte) ([]byte, error) {
	blockSize := len(ctx.roundKeys[0])

	if ctx.iv == nil || len(ctx.iv) != blockSize {
		return nil, errors.New("неверный вектор инициализации (IV)")
	}

	result := make([]byte, len(data))
	counter := make([]byte, blockSize)
	copy(counter, ctx.iv)

	for i := 0; i < len(data); i += blockSize {
		encryptedCounter, err := ctx.encryptBlock(counter)
		if err != nil {
			return nil, err
		}

		blockSizeToUse := min(blockSize, len(data)-i)
		outputBlock := xorBlocks(data[i:i+blockSizeToUse], encryptedCounter[:blockSizeToUse])

		copy(result[i:i+blockSizeToUse], outputBlock)

		incrementCounter(counter)
	}

	return result, nil
}

func (ctx *CryptoContext) encryptRandomDelta(data []byte) ([]byte, error) {
	blockSize := len(ctx.roundKeys[0])

	result := make([]byte, len(data))

	for i := 0; i < len(data); i += blockSize {
		delta := make([]byte, blockSize)
		_, err := rand.Read(delta)
		if err != nil {
			return nil, err
		}

		blockSizeToUse := min(blockSize, len(data)-i)
		intermediateBlock := xorBlocks(data[i:i+blockSizeToUse], delta[:blockSizeToUse])

		encryptedBlock, err := ctx.encryptBlock(intermediateBlock)
		if err != nil {
			return nil, err
		}

		copy(result[i:i+blockSizeToUse], encryptedBlock)
	}

	return result, nil
}

func (ctx *CryptoContext) decryptRandomDelta(data []byte) ([]byte, error) {
	return nil, errors.New("режим Random Delta не поддерживает дешифрование без дополнительных данных")
}

func (ctx *CryptoContext) encryptBlock(block []byte) ([]byte, error) {
	var err error
	encryptedBlock := make([]byte, len(block))
	copy(encryptedBlock, block)

	for _, roundKey := range ctx.roundKeys {
		encryptedBlock, err = ctx.transform.EncryptBlock(encryptedBlock, roundKey)
		if err != nil {
			return nil, err
		}
	}

	return encryptedBlock, nil
}

func (ctx *CryptoContext) decryptBlock(block []byte) ([]byte, error) {
	var err error
	decryptedBlock := make([]byte, len(block))
	copy(decryptedBlock, block)

	for i := len(ctx.roundKeys) - 1; i >= 0; i-- {
		roundKey := ctx.roundKeys[i]
		decryptedBlock, err = ctx.transform.DecryptBlock(decryptedBlock, roundKey)
		if err != nil {
			return nil, err
		}
	}

	return decryptedBlock, nil
}

func xorBlocks(a, b []byte) []byte {
	result := make([]byte, min(len(a), len(b)))
	for i := range result {
		result[i] = a[i] ^ b[i]
	}
	return result
}

func incrementCounter(counter []byte) {
	for i := len(counter) - 1; i >= 0; i-- {
		counter[i]++
		if counter[i] != 0 {
			break
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (ctx *CryptoContext) EncryptFile(inputPath string, outputPath string) error {
	return ctx.processFile(inputPath, outputPath, ctx.Encrypt)
}

func (ctx *CryptoContext) DecryptFile(inputPath string, outputPath string) error {
	return ctx.processFile(inputPath, outputPath, ctx.Decrypt)
}

func (ctx *CryptoContext) processFile(inputPath string, outputPath string, processFunc func([]byte) ([]byte, error)) error {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	blockSize := len(ctx.roundKeys[0]) * 1024

	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	doneChan := make(chan struct{})

	reader := bufio.NewReader(inputFile)
	writer := bufio.NewWriter(outputFile)
	defer writer.Flush()

	dataChan := make(chan []byte, 10)

	go func() {
		for {
			buffer := make([]byte, blockSize)
			n, err := reader.Read(buffer)
			if err != nil {
				if errors.Is(err, os.ErrClosed) || n == 0 {
					break
				}
				errChan <- err
				return
			}

			buffer = buffer[:n]

			dataChan <- buffer
		}
		close(dataChan)
	}()

	go func() {
		for data := range dataChan {
			wg.Add(1)
			go func(dataBlock []byte) {
				defer wg.Done()
				processedData, err := processFunc(dataBlock)
				if err != nil {
					errChan <- err
					return
				}

				writerMutex := &sync.Mutex{}
				writerMutex.Lock()
				_, err = writer.Write(processedData)
				writerMutex.Unlock()
				if err != nil {
					errChan <- err
					return
				}
			}(data)
		}
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		return nil
	case err := <-errChan:
		return err
	}
}
