package tinygarble

import (
	"testing"
)

func TestReverseEndianness(t *testing.T) {
	var test string
	test = ReverseEndianness(string("DEC0ADDE00"))
	if test != "00DEADC0DE" {
		t.Error("Expected DEADC0DE, got ", test)
	}

	test = ReverseEndianness(string("A0B70708"))
	if test != "0807B7A0" {
		t.Error("Expected 0807B7A0, got ", test)
	}
}

func TestxorStr(t *testing.T) {
	var str1 string
	str1 = xorStr(string("DEADC0DE"), string("DEADC0DE"))
	if str1 != "00000000" {
		t.Error("Expected 00000000, got ", str1)
	}

	str1 = xorStr(string("ee2eee1ff1"), string("fffffffffe"))
	if str1 != "11d111e00f" {
		t.Error("Expected 11d111e00f, got ", str1)
	}

	str1 = xorStr(string("00000000"), string("DEADC0DE"))
	if str1 != "DEADC0DE" {
		t.Error("Expected DEADC0DE, got ", str1)
	}
}
