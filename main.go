package main

import (
	"fmt"
	"iSL1/des"
)

func main() {
	key := []byte{0x13, 0x34, 0x57, 0x79, 0x9B, 0xBC, 0xDF, 0xF1}
	plaintext := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}

	desCipher, err := des.NewDES()
	if err != nil {
		fmt.Println("Ошибка при создании DES:", err)
		return
	}

	err = desCipher.SetKey(key)
	if err != nil {
		fmt.Println("Ошибка при установке ключа:", err)
		return
	}

	ciphertext, err := desCipher.EncryptBlock(plaintext)
	if err != nil {
		fmt.Println("Ошибка при шифровании:", err)
		return
	}

	fmt.Printf("Зашифрованные данные: % X\n", ciphertext)

	decryptedText, err := desCipher.DecryptBlock(ciphertext)
	if err != nil {
		fmt.Println("Ошибка при дешифровании:", err)
		return
	}

	fmt.Printf("Расшифрованные данные: % X\n", decryptedText)
}
