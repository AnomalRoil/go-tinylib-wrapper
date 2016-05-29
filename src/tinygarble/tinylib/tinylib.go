package tinylib

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

var tinyPath string
var circuitPath string

// An utilitary function to set the path to the relevant component in order to be able to use TinyGarble
func SetTinyPaths(tiPath string, ciPath string) {
	tinyPath = tiPath
	circuitPath = ciPath
}

// A utilitary function to reverse endianness from little/big to big/little endian for a string of hex values
func ReverseEndianness(data string) string {
	//initalizing the return value as an empty string
	ans := ""

	//trimming since there are easily \n in cmd lines outputs.
	data = strings.TrimSpace(data)
	if len(data)%2 != 0 {
		log.Fatal("You can't change the endianness of a string whose length isn't a multiple of 2. Please check your data.")
		return data
	}
	//if the data isn't in hex format, e.g. if it hasn't an even number of char, then the programmer made some mistake. However this isn't checking it is actually hex

	for len(data) >= 2 {
		ans += data[len(data)-2:]
		data = data[:len(data)-2]
	}

	return ans
}

func YaoServer(data string, port string, clock_cycles int) {
	fmt.Printf("Alice here, running as server on port %s.\n", port)
	// Note the change of endianness for the data, since TinyGarble uses little endian
	data = ReverseEndianness(data)

	inputArg := "--input"
	if clock_cycles > 1 {
		inputArg = "--init"
	}

	fmt.Println(tinyPath, circuitPath, inputArg, data, port)
	out, err := exec.Command(
		tinyPath+"/bin/garbled_circuit/TinyGarble", "-a",
		"-i", circuitPath,
		inputArg, data[:32], "-p", port).Output()

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Alice output:  %s\n", out)
}

func YaoClient(data string, addr string, port string, clock_cycles int) string {
	fmt.Printf("Bob here, running as client on address %s and port %s.\n", addr, port)
	// Note the change of endianness for the data, since TinyGarble uses little endian
	data = ReverseEndianness(data)

	inputArg := "--input"
	if clock_cycles > 1 {
		inputArg = "--init"
	}

	out, err := exec.Command(
		tinyPath+"/bin/garbled_circuit/TinyGarble", "-b",
		"-i", circuitPath,
		inputArg, data[:32], // this should be handled more properly
		"-s", addr, "-p", port, "--output_mode", "2").Output()
	// We specify the --output_mode arg to be "last_clock" only, since otherwise it would output each clock cycle intermediate states when using multiple cycles circuits

	if err != nil {
		log.Fatal(err)
	}
	hexOut := ReverseEndianness(string(out))
	fmt.Printf("Client output:  %s\n", hexOut)

	return hexOut
}

// This function allows to use TinyGarble to encrypt more than 128 bits of data using CBC mode with ciphertext stealing (to avoid padding)
func AESCBC(data string, addr string, startingPort string, iv string) ([]string, string) {
	port, err := strconv.Atoi(startingPort)
	if err != nil {
		log.Fatal(err)
	}

	var toCrypt []string
	var cipher []string
	// splitting the data into 128 bits blocks or less for the last one.
	// We are using ciphertext stealing to avoid padding
	toCrypt = splitData(data)
	ivUsed := ivGeneration(iv)
	// We can use the IV to do ciphertext stealing in case of <128 bits data, but this will be implemented later
	if len(toCrypt) < 2 {
		log.Fatal("As of now, this CBC implementation needs more at least 128 bits of data to encrypt them")
	}
	// we set the IV as the first item used for xoring:
	xoring := hex.EncodeToString(ivUsed)
	// secure encryption of the data :
	for i, r := range toCrypt {
		dataLen := len(r)
		// if we use ciphertext stealing we need to pad the data with 0's
		if i == len(toCrypt) && dataLen < 32 {
			r += strings.Repeat("0", 32-dataLen)
			fmt.Println("Using ciphertext stealing on", xoring, r, dataLen)
		}
		plain := xorStr(r, xoring)
		fmt.Println("Sending :", plain)
		ct := YaoClient(plain, addr, strconv.Itoa(port+i), 1)
		cipher = append(cipher, ct)
		xoring = ct
		// ciphertext stealing in action:
		if i == len(toCrypt) && dataLen < 32 {
			stolenCiph := cipher[len(cipher)-1][:dataLen]
			cipher = append(cipher[:len(cipher)-2], ct, stolenCiph)
		}
	}

	fmt.Println("Ciphertext in CBC:", cipher)

	return cipher, iv
}

func ivGeneration(customIv string) []byte {
	// iv generation:
	// iv size is 128 bits:
	ivLength := 16
	ivByte := make([]byte, ivLength)
	_, err := rand.Read(ivByte)
	if err != nil {
		log.Fatal(err)
	}
	// To allow the use of a given  iv (mainly for testing purpose) :
	if customIv != "" && len(customIv) == 32 {
		fmt.Println("Be careful when using a custom iv as now: randomness reuses are dangerous")
		ivByte, err = hex.DecodeString(customIv)
		if err != nil {
			log.Fatal(err)
		}
	}

	return ivByte
}

func splitData(data string) []string {
	var toCrypt []string
	for i, _ := range data {
		if i > 0 && (i+1)%32 == 0 {
			toCrypt = append(toCrypt, data[i-31:i+1])
		} else {
			if i == len(data) {
				toCrypt = append(toCrypt, data[len(data)-(len(data)+1)%32:len(data)])
			}
		}
	}
	return toCrypt
}

// This function allows to use Tinygarble to encrypt more than 128 bit in a secure way through the use of CTR mode
func AESCTR(data string, addr string, startingPort string, iv string) ([]string, string) {
	fmt.Println("AES CTR started")
	port, err := strconv.Atoi(startingPort)
	if err != nil {
		log.Fatal(err)
	}
	//lets splice our data into 32 char :
	var toCrypt []string
	var cipher []string
	// we split the data into 128 bits or less for the last one. No padding needed for CTR mode
	toCrypt = splitData(data)

	// Counter generation:
	// counter size is 128 bits and is a nonce unless a custom one is used, i.e. iv!="":
	counterByte := ivGeneration(iv)
	// we split the counter and increment only the last 64 bits so we can use the int64 type without needing to use big int: this is okay since we won't encrypt exabytes of data
	halfCounter := counterByte[8:]
	var count uint64
	// we set count to be equal to the value stored as bytes in the halfCounter
	buf := bytes.NewReader(halfCounter)
	err = binary.Read(buf, binary.BigEndian, &count)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	var counter []string
	for i := 0; i < len(toCrypt); i++ {
		bif := new(bytes.Buffer)
		// we translate the counter into bytes (from int64)
		err = binary.Write(bif, binary.BigEndian, count)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
		}
		// we append the lower 64 bits of the counter with the upper 64 bits
		halfCounter = append(counterByte[:8], bif.Bytes()...)
		// we add it to the list we will enrypt later
		counter = append(counter, hex.EncodeToString(halfCounter))
		// we increment the counter
		count = count + uint64(1)
	}

	// secure encryption of the counter :
	for i, r := range counter {
		fmt.Println("Sending :", r)
		ct := YaoClient(r, addr, strconv.Itoa(port+i), 1)
		cipher = append(cipher, ct)
	}

	cipherText := make([]string, len(cipher))
	for i, r := range toCrypt {
		cipherText[i] = xorStr(cipher[i], r)
	}
	fmt.Println("Ciphertext in CTR:", cipherText)
	return cipherText, iv
}

// Helper method to xor (hexadecimal) strings together
func xorStr(str1 string, str2 string) string {
	s1, e1 := hex.DecodeString(str1)
	s2, e2 := hex.DecodeString(str2)
	if e1 != nil || e2 != nil {
		fmt.Println("Decoding from string failed:", e1, e2)
	}
	mini := len(s2)
	if len(s1) < len(s2) {
		mini = len(s1)
	}

	str := make([]byte, mini)
	for i := 0; i < mini; i++ {
		str[i] = s1[i] ^ s2[i]
	}

	return strings.ToUpper(hex.EncodeToString(str))

}

func CTRServer(key string, port string, rounds int) {
	startingPort, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal(err)
	}
	// TODO : find a good way to decide weither the server can stop or not
	// maybe establish myself a TCP connexion in order to communicate with
	// Bob to decide the next port to use and/or if it is finished ???
	// However it'll be certainly easier to just timeout
	for rounds > 0 || rounds < 0 { // This allows unending server cycles
		YaoServer(key, strconv.Itoa(startingPort), 1)
		startingPort++
		rounds--
		// This terminates when rounds == 0
	}
}
