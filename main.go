package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	SPECIAL_CHARACTERS    = "!@#$%^&*+-="
	UPPER_CASE_CHARACTERS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LOWER_CASE_CHARACTERS = "abcdefghijklmnopqrstuvwxyz"
	NUMBERS               = "0123456789"
)

func bufferToHex(buffer []byte) string {
	str := ""
	for i := 0; i < len(buffer); i++ {
		str += fmt.Sprintf("%02x", buffer[i])
		// fmt.Println("b is", buffer[i])
	}
	return str
}

func getCharFromHash(hash string, index int, allCharacters string) byte {
	start := index * 2
	end := start + 2
	if end > len(hash) {
		end = len(hash)
	}
	slice := hash[start:end]
	num, err := hex.DecodeString(slice)
	if err != nil || len(num) == 0 {
		return ' '
	}
	charIndex := int(num[0]) % len(allCharacters)
	return allCharacters[charIndex]
}

func getSecret(enrollmentID string) (string, error) {
	// Hash the enrollmentID using SHA-256
	enrollmentBytes := []byte(enrollmentID)
	hash := sha256.Sum256(enrollmentBytes)
	hashString := bufferToHex(hash[:])

	// Ensure at least one character from each set is included
	uniqueString := ""
	uniqueString += string(UPPER_CASE_CHARACTERS[int(hash[0])%len(UPPER_CASE_CHARACTERS)])
	uniqueString += string(LOWER_CASE_CHARACTERS[int(hash[1])%len(LOWER_CASE_CHARACTERS)])
	uniqueString += string(SPECIAL_CHARACTERS[int(hash[2])%len(SPECIAL_CHARACTERS)])
	uniqueString += string(NUMBERS[int(hash[3])%len(NUMBERS)])

	// Fill the rest of the string with characters from the hash
	allCharacters := UPPER_CASE_CHARACTERS + LOWER_CASE_CHARACTERS + SPECIAL_CHARACTERS + NUMBERS
	for i := 4; i < 16; i++ {
		uniqueString += string(getCharFromHash(hashString, i, allCharacters))
	}

	return uniqueString, nil
}

func main() {
	enrollmentID := "5d7ba51aa50c181a91530c2c8c5e5256404d7fa4"
	secret, err := getSecret(enrollmentID)
	if err != nil {
		fmt.Printf("Error generating secret: %v\n", err)
		return
	}
	fmt.Printf("Generated Secret: %s\n", secret)

	// fmt.Println(bufferToHex([]byte{0x12, 0x34, 0xb}))
}
