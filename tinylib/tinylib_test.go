package tinylib

import (
	"encoding/hex"
    "fmt"
    "math/rand"
    "os"
    "strings"
    "time"
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

func TestAESServ(t *testing.T) {
    fmt.Println("Starting AES CTR mode test")
    path :=os.Getenv("TINYGARBLE")
    if path == "" {
        t.Skip("skipping test; $TINYGARBLE not set")
    }

    key := "2b7e151628aed2a6abf7158809cf4f3c"

    SetCircuit(path,path+"/scd/netlists/aes_1cc.scd",1,false)
    s1 := rand.NewSource(time.Now().UnixNano())
    r1 := rand.New(s1)
    port := r1.Intn(100)
    fmt.Println("Using port :", 1234+port)
    fmt.Println("Note that this test assumes the localhost range 1234-1334 to be usable.")
    go AESServer(key,1234+port,1)
    time.Sleep(500*time.Millisecond)

    fmt.Println("Continuing test with the client")

    iv := "f0f1f2f3f4f5f6f7f8f9fafbfcfdfeff"
    data := "6bc1bee22e409f96e93d7e117393172a"
    awaitedResult := "874d6191b620e3261bef6864990db6ce"

    ans, _ := AESCTR(data,"127.0.0.1",1234+port,iv)

    if ans[0] != strings.ToUpper(awaitedResult) {
		t.Error("Expected 874d6191b620e3261bef6864990db6ce, got ", ans)
    } else {
        fmt.Println("AES CTR Test passed")
    }
}

func TestAESCBC(t *testing.T) {
    fmt.Println("Testing the AES CBC mode, first starting the server :")
    path :=os.Getenv("TINYGARBLE")
    if path == "" {
        t.Skip("skipping test; $TINYGARBLE not set")
    }

    key := "2b7e151628aed2a6abf7158809cf4f3c"

    SetCircuit(path,path+"/scd/netlists/aes_1cc.scd",1,false)
    s1 := rand.NewSource(time.Now().UnixNano())
    r1 := rand.New(s1)
    port := r1.Intn(100)
    fmt.Println("Using port :", 1234+port)
    fmt.Println("Note that this test assumes the localhost range 1234-1334 to be usable.")
    go AESServer(key,1234+port,1)
    time.Sleep(time.Millisecond*500)

    iv := "7649ABAC8119B246CEE98E9B12E9197D"
    data := "ae2d8a571e03ac9c9eb76fac45af8e51"
    awaitedResult := strings.ToUpper("5086cb9b507219ee95db113a917678b2")

    ans, _ := AESCBC(data,"127.0.0.1", 1234+port,iv)
    if ans[0] != awaitedResult {
        t.Error("Expected", awaitedResult, "got", ans)
    }
}
