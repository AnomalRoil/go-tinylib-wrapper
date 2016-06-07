package main

import (
	"crypto/aes"
	"encoding/hex"
	"fmt"
	"log"
	"projet-pa-yao/code/tinylib"
)

func main() {

	key, err := hex.DecodeString("754620676E754B20796D207374616854")
	plaintext, err1 := hex.DecodeString("6F775420656E694E20656E4F206F7754")
	if err != nil && err1 != nil {
		log.Fatal("Error while decoding the hex strings into bytes")
	}

	ciphertext := make([]byte, aes.BlockSize)

	cipher, errc := aes.NewCipher(key)
	if errc != nil {
		log.Fatal("Error while setting the key for AES-128")
	}
	cipher.Encrypt(ciphertext, plaintext)

	fmt.Println("Ciphertext:\t", hex.EncodeToString(ciphertext))

	key2, err2 := hex.DecodeString(tinylib.ReverseEndianness("754620676E754B20796D207374616854"))
	plaintext, err1 = hex.DecodeString(tinylib.ReverseEndianness("6F775420656E694E20656E4F206F7754"))
	if err2 != nil && err1 != nil {
		log.Fatal("Error while decoding the reversed hex strings into bytes")
	}
	cipher, errc = aes.NewCipher(key2)
	if errc != nil {
		log.Fatal("Error while setting the key for AES-128")
	}

	cipher.Encrypt(ciphertext, plaintext)

	fmt.Println("Ciphertext:\t", hex.EncodeToString(ciphertext))
	fmt.Println("once reversed:\t", tinylib.ReverseEndianness(hex.EncodeToString(ciphertext)))
}
