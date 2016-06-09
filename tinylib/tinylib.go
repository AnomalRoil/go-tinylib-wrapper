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
var clockCycles int
var forceInput bool

// Be careful, you have to first set the TinyGarble Path and the Circuit Path to the AES-128 circuit, in order to use this
// This function allows to use TinyGarble to encrypt more than 128 bits of data using CBC mode with ciphertext stealing (to avoid padding)
func AESCBC(data string, addr string, port int, o_iv ...string) ([]string, string) {
	fmt.Println("\tAES CBC started")

	var toCrypt []string
	var cipher []string
	// splitting the data into 128 bits blocks or less for the last one.
	toCrypt = SplitData(data,32)
	// We are using ciphertext stealing to avoid to have to use padding on the data
	
	//IV
	iv := ""
	//check wheter an optional iv is specified, since we just work with AES 128, we conditionnally check if it has the good length to be an iv
	if len(o_iv) > 0 && len(o_iv[0]) == 32 {
		iv = o_iv[0]
		//Note that we should also test wheter it is an hexadecimanl string for completness, but we don't as of now.
	}
	ivUsed := ivGeneration(iv)
	// We can use the IV to do ciphertext stealing in case of <128 bits data, but this will be implemented later
	if len(toCrypt[0]) < 32 {
		log.Fatal("As of now, this CBC implementation needs at least 128 bits of data to encrypt them")
	}
	// we set the IV as the first item used for xoring:
	xoring := hex.EncodeToString(ivUsed)
	// secure encryption of the data :
	for i, r := range toCrypt {
		dataLen := len(r)
		// if we use ciphertext stealing we need to pad the data with 0's
		if i == len(toCrypt)-1 && dataLen < 32 {
			r += strings.Repeat("0", 32-dataLen)
		}
		// We have to reverse endianness since TinyGrable AES uses little endian
		plain := ReverseEndianness(xorStr(r, xoring))

		// We have to reverse endianness from the output to work again in big endian
		ct := ReverseEndianness(YaoClient(plain, addr, port+i))
		cipher = append(cipher, ct)
		// ciphertext stealing in action:
		if i == len(toCrypt)-1 && dataLen < 32 {
			stolenCiph := xoring[:dataLen]
			cipher = append(cipher[:len(cipher)-2], ct, stolenCiph)
		}
		xoring = ct
	}

	return cipher, hex.EncodeToString(ivUsed)
}

// Be careful, you have to first set the TinyGarble Path and the Circuit Path to the AES-128 circuit, in order to use this
// This function allows to use Tinygarble to encrypt more than 128 bit in a secure way through the use of CTR mode
func AESCTR(data string, addr string, port int, o_iv ...string) ([]string, string) {
	fmt.Println("\tAES CTR started")

	//lets splice our data into 32 char :
	var toCrypt []string
	var cipher []string
	// we split the data into 128 bits or less for the last one. No padding needed for CTR mode
	toCrypt = SplitData(data,32)

	// Counter generation:
	iv := ""
	//check wheter an optional iv is specified, since we just work with AES 128, we conditionnally check if it has the good length to be an iv
	if len(o_iv) > 0 && len(o_iv[0]) == 32 {
		iv = o_iv[0]
		//Note that we should also test wheter it is an hexadecimanl string for completness, but we don't as of now.
	}
	// counter size is 128 bits and is a random nonce unless a custom one is used, i.e. iv!="":
	counterByte := ivGeneration(iv)
	// we split the counter and increment only the last 64 bits so we can use the int64 type without needing to use big int: this is okay since we won't encrypt exabytes of data and since the probability for being almost at the end of the counter is too low to be worrysome. However it may be good, later, to ensure the counter doesn't reach its max value, since this is still a (low probability) bug.
	halfCounter := counterByte[8:]
	var count uint64
	// we set count to be equal to the value stored as bytes in the halfCounter
	buf := bytes.NewReader(halfCounter)
	err := binary.Read(buf, binary.BigEndian, &count)
	if err != nil {
		log.Fatal("binary.Read failed:", err)
	}
	var counter []string
	for i := 0; i < len(toCrypt); i++ {
		bif := new(bytes.Buffer)
		// we translate the counter into bytes (from int64)
		err = binary.Write(bif, binary.BigEndian, count)
		if err != nil {
			log.Fatal("binary.Write failed:", err)
		}
		// we append the lower 64 bits of the counter with the upper 64 bits
		halfCounter = append(counterByte[:8], bif.Bytes()...)
		// we add it to the list we will enrypt later
		counter = append(counter, hex.EncodeToString(halfCounter))
		// we increment the counter
		count = count + uint64(1)
		// Note that a unint64 won't overflow but wrap around in golang
	}

	// secure encryption of the counter :
	for i, r := range counter {
		ct := YaoClient(ReverseEndianness(r), addr, port+i)
		cipher = append(cipher, ReverseEndianness(ct))
	}

	cipherText := make([]string, len(cipher))
	for i, r := range toCrypt {
		cipherText[i] = xorStr(cipher[i], r)
	}

	return cipherText, hex.EncodeToString(counterByte)
}

// A method allowing one to generate a random iv in a byte slice or to set this iv to the given string (assuming a big endian representation in hexadecimal) and using the secure PRNG from "crypto/rand"
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
		fmt.Println("\tBe careful when using a custom iv as now: randomness reuses are dangerous")
		ivByte, err = hex.DecodeString(customIv)
		if err != nil {
			log.Fatal(err)
		}
	}

	return ivByte
}

// An utilitary function to reverse endianness from little/big to big/little endian for a string of hex values
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

// Be careful, you have to first set the TinyGarble Path and the Circuit Path to the AES-128 circuit, in order to use this
// This function allows to run an server a given number of time "rounds", incrementing the port number each time to avoid problems with the TIME_WAIT
func RunServer(key string, startingPort int, rounds int) {
	// TODO : find a good way to decide weither the server can stop or not
	// maybe establish a TCP connexion in order to communicate with
	// Bob to decide the next port to use and/or if it is finished?
	// However it'll be certainly easier to just timeout. As of now fixed number of rounds:
	for rounds != 0 { // This allows unending server cycles
		YaoServer(key, startingPort)
		// Note that this will crash sometimes if the next port isn't available
		startingPort++
		rounds--
		// This terminates when rounds == 0
	}
}

// An utilitary function to set the path to the relevant component in order to be able to use TinyGarble
func SetCircuit(tiPath string, ciPath string, clCycles int, uInput bool) {
	tinyPath = tiPath
	circuitPath = ciPath
	clockCycles = clCycles
	forceInput = uInput
}

// An utilitary function to easily split the input data into a slice of char blocks of variable sizes as string (or less for the last block)
func SplitData(data string, length int) []string {
	var toCrypt []string
	for i, _ := range data {
		if i > 0 && (i+1)%length == 0 {
			toCrypt = append(toCrypt, data[i-length-1:i+1])
		} else {
			// using len-1 because the index begin at 0, not at 1
			if i == len(data)-1 {
				toCrypt = append(toCrypt, data[len(data)-(len(data))%length:len(data)])
			}
		}
	}
	return toCrypt
}

// Helper method to xor (hexadecimal) strings together
func xorStr(str1 string, str2 string) string {
	s1, e1 := hex.DecodeString(str1)
	s2, e2 := hex.DecodeString(str2)
	if e1 != nil || e2 != nil {
		log.Fatal("Decoding from string failed:", e1, e2)
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

// The wrapper function for the TinyGarble client option
func YaoClient(data string, addr string, port int) string {
	fmt.Printf("\tClient running on address %s and port %d.\n", addr, port)

	// we will use the following arguments when we run the client :
	var yaoArgs []string
	yaoArgs = []string{"-b", "-i", circuitPath,
		"-s", addr, "-p", strconv.Itoa(port),
		"--output_mode", "2"}
	// We specify the --output_mode arg to be "last_clock", aka 2, only, since otherwise it would output each clock cycle intermediate states when using multiple cycles circuits

	var inputArg []string
	if clockCycles > 1 {
		inputArg = []string{"--clock_cycle", strconv.Itoa(clockCycles)}
	} else {
		forceInput = true
	}

	if forceInput {
		inputArg = append(inputArg, "--input", data)
	} else {
		inputArg = append(inputArg, "--init", data)
	}

	yaoArgs = append(yaoArgs, inputArg...)

	//log.Println("Arguments used to run TinyGarble:",yaoArgs)
	cmd := exec.Command(tinyPath+"/bin/garbled_circuit/TinyGarble", yaoArgs...)

	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	return string(out)
}

// A wrapper function for the TinyGarble with server (alice) argument set
func YaoServer(data string, port int) {
	fmt.Printf("\tServer running on port %d.\n", port)

	var yaoArgs []string
	yaoArgs = []string{"-a", "-i", circuitPath,
		"-p", strconv.Itoa(port)}

	var inputArg []string
	if clockCycles > 1 {
		inputArg = []string{"--clock_cycle", strconv.Itoa(clockCycles)}
	} else {
		forceInput = true
	}

	if forceInput {
		inputArg = append(inputArg, "--input", data)
	} else {
		inputArg = append(inputArg, "--init", data)
	}

	yaoArgs = append(yaoArgs, inputArg...)

	_, err := exec.Command(
		tinyPath+"/bin/garbled_circuit/TinyGarble", yaoArgs...).Output()
	if err != nil {
		log.Fatal(err)
	}
}
