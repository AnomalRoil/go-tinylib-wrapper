package main 

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var tinyPath string
var circuitPath string

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
	//if the data isn't in hex format, i.e. if it hasn't an even number of char, then the programmer made some mistake.

    for len(data) >= 2 {
		ans += data[len(data)-2:]
		//fmt.Printf(ans+"\n")
		data = data[:len(data)-2]
	}

	return ans
}

func YaoServer(data string, port string, clock_cycles int) {
	fmt.Printf("Alice here, running as server on port %s.\n", port)
	// Note the change of endianness for the data
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
	// Note the change of endianness for the data
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
	// We specify the --output_mode arg to be "last_clock" only, since otherwise it would output each clock cycle intermediate states.

	if err != nil {
		log.Fatal(err)
	}
	hexOut := ReverseEndianness(string(out))
	fmt.Printf("Client output:  %s\n", hexOut)

	return hexOut
}

// This function allows to use Tinygarble to encrypt more than 128 bit in a secure way through the use of CTR mode
func AESCTR(data string, addr string, startingPort string, iv string) {
	port, err := strconv.Atoi(startingPort)
	if err != nil {
		log.Fatal(err)
	}
	//lets splice our data into 32 char :
	var toCrypt []string
	var cipher []string
	for i, _ := range data {
		if i > 0 && (i+1)%32 == 0 {
			toCrypt = append(toCrypt, data[i-31:i+1])
		} else {
			if i == len(data) {
				toCrypt = append(toCrypt, data[len(data)-(len(data)+1)%32:len(data)])
			}
		}
	}
	// Counter generation:
    // counter size is 128 bits:
	ctrLength := 16
	counterByte := make([]byte, ctrLength)
	_, err = rand.Read(counterByte)
	if err != nil {
		log.Fatal(err)
	}
    // Test with custom iv :
	if iv != "" && len(iv)==32 {
        counterByte, _  = hex.DecodeString(iv)
    }
    // we split the counter and will increment only the last 64 bits so we can use the int64 type without needing to use big int: this is okay since we won't encrypt exabytes of data
	halfCounter := counterByte[8:]
	var count uint64
	buf := bytes.NewReader(halfCounter)
	err = binary.Read(buf, binary.BigEndian, &count)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	var counter []string
	for i := 0; i < len(toCrypt); i++ {
		//fmt.Println("counter:",i,count)
		bif := new(bytes.Buffer)
		err = binary.Write(bif, binary.BigEndian, count)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
		}
		halfCounter = append(counterByte[:8], bif.Bytes()...)
		counter = append(counter, hex.EncodeToString(halfCounter))
		//fmt.Println("in Bytes :",counter)
		count = count + uint64(1)
	}
    fmt.Println("Counter used:")
	// secure encryption of the counter :
	for i, r := range counter {
		fmt.Println("Sending :", r)
		ct := YaoClient(r, addr, strconv.Itoa(port+i), 1)
		cipher = append(cipher, ct)
	}

	cipherText := make([]string, len(cipher))
	for i, r := range toCrypt {
		//fmt.Println("cipher", cipher[i], "plain", r)
		cipherText[i] = xorStr(cipher[i], r)
	}
	fmt.Println("Ciphertext in CTR:", cipherText)

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
	for rounds >= 0 {
		YaoServer(key, strconv.Itoa(startingPort), 1)
		startingPort++
        rounds--
	}
}

func main() {

	nArgs := len(os.Args)

	if nArgs <= 1 {
		log.Fatal("No argument, please run as server Alice (-a) first and then as Bob (-b). Use -h for help.")
	}

	rootPtr := flag.String("r", os.Getenv("TINYGARBLE"), "TinyGarble root directory path, default to $TINYGARBLE if var set, writes $TINYGARBLE if changed")
	circuitPtr := flag.String("n", "aes_1cc.scd", "name of the circuit file located in the circuit root directory")
	circuitPathPtr := flag.String("c", "$TINYGARBLE/scd/netlists", "location of the circuit root directory. Default : $TINYGARBLE/scd/netlists")
	clockcyclesPtr := flag.Int("cc", 1, "number of clock cycles needed for this circuit, usually 1, usually indicated at the end of the circuit name, sha3_24cc needs 24 clock cycles for example")

	portsPtr := flag.String("p", "1234", "Specify a starting port")
	addrPtr := flag.String("s", "127.0.0.1", "Specify a server address for Bob to connect.")

	alicePtr := flag.Bool("a", false, "Run as server Alice")
	bobPtr := flag.Bool("b", false, "Run as client Bob")
	ctrPtr := flag.Bool("ctr", false, "Run using CTR mode and aes circuit in 1cc")

	initPtr := flag.String("d", "00000000000000000000000000000000", "Init data")

	flag.Parse()

	// Checking the remaining flag used : if there are unknown flags, we stop.
	if len(flag.Args()) > 0 {
		log.Fatal("Please use the good arguments. Use -h for help.")
	}

	// We now initialize the TinyGarble root path
	tinyPath = *rootPtr
	if tinyPath == "" {
		log.Fatal("$TINYGARBLE is not set. Please provide path to TinyGarble's root as argument or set $TINYGARBLE env var to its path.")
	}

	// We next initialize the circuit path based on the default value or on the user's input
	circuitPath = *circuitPathPtr + "/"
	// We convert $TINYGARBLE into tinyPath, mainly because of the default value includes it.
	if strings.Contains(circuitPath, "$TINYGARBLE") {
		circuitPath = strings.Replace(circuitPath, "$TINYGARBLE", tinyPath, -1)
	}

	circuitPath += *circuitPtr

	// sanity check for the input
	if len(*initPtr) < 32 {
		log.Fatal("Please give an init value of length 32 at least until I've implemented padding.")
	}

	// we can continue, everything is initialized.
	switch {
	case *ctrPtr && *alicePtr:
		CTRServer(*initPtr, *portsPtr,0)
	case *ctrPtr && *bobPtr:
		AESCTR(*initPtr, *addrPtr, *portsPtr, "")
	case *alicePtr:
		YaoServer(*initPtr, *portsPtr, *clockcyclesPtr)
	case *bobPtr:
		YaoClient(*initPtr, *addrPtr, *portsPtr, *clockcyclesPtr)
	default: // if running neither as Alice, nor as Bob, there is a misuse
		log.Fatal("Please run as server Alice (-a) first and then as Bob (-b). Use -h for help.")
	}
}
