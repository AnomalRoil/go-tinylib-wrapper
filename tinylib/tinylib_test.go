package tinylib

import (
	"encoding/hex"
	"testing"
)

func TestIvGeneration(t *testing.T) {
	var test []byte
	test = ivGeneration(string("12345678901234567890123456789012"))
	ans := hex.EncodeToString(test)
	if ans != "12345678901234567890123456789012" {
		t.Error("Expected 12345678901234567890123456789012, got ", ans)
	}
	// Further testing of the IV generation without custom iv is not necessary: the random generator used should be tested by their creator, not here.
}

// Basic test to try out the conversion from Little/Big to Big/Little Endian
func TestReverseEndianness(t *testing.T) {
	var test string
	test = ReverseEndianness(string("DEC0ADDE00"))
	if test != "00DEADC0DE" {
		t.Error("Expected 00DEADC0DE, got ", test)
	}

	test = ReverseEndianness(string("A0B70708"))
	if test != "0807B7A0" {
		t.Error("Expected 0807B7A0, got ", test)
	}
}

// Basic tests to see if the xoring of strings works
func TestXorStr(t *testing.T) {
	var str1 string
	str1 = xorStr(string("DEADC0DE"), string("DEADC0DE"))
	if str1 != "00000000" {
		t.Error("Expected 00000000, got ", str1)
	}

	str1 = xorStr(string("ee2eee1ff1"), string("fffffffffe"))
	if str1 != "11D111E00F" {
		t.Error("Expected 11D111E00, got ", str1)
	}

	str1 = xorStr(string("00000000"), string("DEADC0DE"))
	if str1 != "DEADC0DE" {
		t.Error("Expected DEADC0DE, got ", str1)
	}
}

// Test vector for AES-CTR 128:
// IV : f0f1f2f3f4f5f6f7f8f9fafbfcfdfeff
// Encryption Key : 2b7e151628aed2a6abf7158809cf4f3c
//
// test : 6bc1bee22e409f96e93d7e117393172a --> 874d6191b620e3261bef6864990db6ce
