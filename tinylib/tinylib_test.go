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
    port := r1.Intn(5000)
    fmt.Println("Using port :", 49152+port)
    fmt.Println("Note that this test uses randomly 4 consequent ports in the range 49152-54152.")
    go AESServer(key,49152+port,4)
    time.Sleep(500*time.Millisecond)

    fmt.Println("Continuing test with the client")

    iv := "f0f1f2f3f4f5f6f7f8f9fafbfcfdfeff"
    data := "6bc1bee22e409f96e93d7e117393172aae2d8a571e03ac9c9eb76fac45af8e5130c81c46a35ce411e5fbc1191a0a52eff69f2445df4f9b17ad2b417be66c3710"
    awaitedResult := "874d6191b620e3261bef6864990db6ce9806f66b7970fdff8617187bb9fffdff5ae4df3edbd5d35e5b4f09020db03eab1e031dda2fbe03d1792170a0f3009cee"

    ans, _ := AESCTR(data,"127.0.0.1",49152+port,iv)

    if strings.Join(ans,"") != strings.ToUpper(awaitedResult) {
		t.Error("Expected 874d6191b620e3261bef6864990db6ce9806f66b7970fdff8617187bb9fffdff5ae4df3edbd5d35e5b4f09020db03eab1e031dda2fbe03d1792170a0f3009cee, got ", ans)
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

    key := "636869636b656e207465726979616b69"

    SetCircuit(path,path+"/scd/netlists/aes_1cc.scd",1,false)
    s1 := rand.NewSource(time.Now().UnixNano())
    r1 := rand.New(s1)
    port := r1.Intn(1000)
    fmt.Println("Using port :",49152+port)
    fmt.Println("Note that this test assumes the localhost range 49152-50152 to be usable.")
    go AESServer(key,49152+port,3)
    time.Sleep(time.Millisecond*500)

    iv   := "00000000000000000000000000000000"
    data := "4920776f756c64206c696b65207468652047656e6572616c20476175277320436869636b656e2c20706c656173652c"
    awaitedResult := strings.ToUpper("97687268d6ecccc0c07b25e25ecfe584b3fffd940c16a18c1b5549d2f838029e39312523a78662d5be7fcbcc98ebf5")

    ans, _ := AESCBC(data,"127.0.0.1",49152+port,iv)
    if strings.Join(ans,"") != awaitedResult {
        t.Error("Expected", awaitedResult, "got", ans)
    }
}

func TestHamming1cc(t *testing.T) {
    fmt.Println("Testing with Hamming 32bits 1 clock cycles, first starting the server :")
    path :=os.Getenv("TINYGARBLE")
    if path == "" {
        t.Skip("skipping test; $TINYGARBLE not set")
    }

    SetCircuit(path,path+"/scd/netlists/hamming_32bit_1cc.scd",1,false)

    s1 := rand.NewSource(time.Now().UnixNano())
    r1 := rand.New(s1)
    port := r1.Intn(1000)
    fmt.Println("Using port :",49152+port)
    fmt.Println("Note that this test assumes the localhost range 49152-50152 to be usable.")
    go YaoServer("FF55AA77",49152+port)
    time.Sleep(time.Millisecond*500)

    ans := YaoClient("12345678", "127.0.0.1", 49152+port)
    if ans != "13\n" {
        t.Error("Expected 13, got", ans)
    }
}

func TestHamming8cc(t *testing.T) {
    fmt.Println("Testing with Hamming 32bits 8 clock cycles, first starting the server :")
    path :=os.Getenv("TINYGARBLE")
    if path == "" {
        t.Skip("skipping test; $TINYGARBLE not set")
    }

    SetCircuit(path,path+"/scd/netlists/hamming_32bit_8cc.scd",8,true)

    s1 := rand.NewSource(time.Now().UnixNano())
    r1 := rand.New(s1)
    port := r1.Intn(1000)
    fmt.Println("Using port :",49152+port)
    fmt.Println("Note that this test assumes the localhost range 49152-50152 to be usable.")
    go YaoServer("FF55AA77",49152+port)
    time.Sleep(time.Millisecond*500)

    ans := YaoClient("12345678", "127.0.0.1", 49152+port)
    if ans != "13" {
        t.Error("Expected 13, got", ans)
    }
}
