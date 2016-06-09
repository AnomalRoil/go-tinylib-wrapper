package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"projet-pa-yao/code/tinylib"
	"strings"
)

var tinyPath string
var circuitPath string

func main() {

	nArgs := len(os.Args)

	if nArgs <= 1 {
		log.Fatal("No argument, please run as server Alice (-a) first and then as Bob (-b). Use -h for help.")
	}

	// To specify the location of the tinygarble executable, the circuit used and the clock cycles, as well as wether this circuit enforce the use of the --input flag if it has more than 1 clock cycles
	rootPtr := flag.String("r", os.Getenv("TINYGARBLE"), "the TinyGarble root directory path, default to $TINYGARBLE if var set, writes $TINYGARBLE if changed")
	circuitPtr := flag.String("n", "aes_1cc.scd", "name of the circuit file located in the circuit root directory")
	circuitPathPtr := flag.String("c", "$TINYGARBLE/scd/netlists", "location of the circuit root directory.")
	clockcyclesPtr := flag.Int("cc", 1, "number of clock cycles needed for this circuit, usually 1, usually indicated at the end of the circuit name, sha3_24cc needs 24 clock cycles for example")
	forceInputPtr := flag.Bool("input", false, "some circuits are using more than 1 clock cycles but don't use the init flag in TinyGarble. This allows to enforce the use of the --input flag instead of the --init one.")

	portsPtr := flag.Int("p", 1234, "Specify a starting port")
	addrPtr := flag.String("s", "127.0.0.1", "Specify a server address for Bob to connect.")

	alicePtr := flag.Bool("a", false, "run as server Alice")
	bobPtr := flag.Bool("b", false, "run as client Bob")
	ctrPtr := flag.Bool("ctr", false, "run using CTR mode and aes circuit in 1cc")
	cbcPtr := flag.Bool("cbc", false, "run using CBC mode and aes circuit in 1cc")
	customIv := flag.String("iv", "", "allows to specify a custom IV for the CTR mode, only for testing : using custom IV may be dangerous, since CTR is sensible to randomness reuses")
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
	tinylib.SetCircuit(tinyPath, circuitPath, *clockcyclesPtr, *forceInputPtr)

	// sanity check for the input
	if len(*initPtr) < 32 && (*ctrPtr || *cbcPtr) {
		log.Fatal("Please give an init value of length 32 at least until I've implemented padding.")
	}

	// we can continue, everything is initialized.
	switch {
	case (*cbcPtr || *ctrPtr) && *alicePtr:
		fmt.Println("Launching AES CTR server with key:", *initPtr)
		// Note the change of endianness for the data, since the AES_1cc uses little endian
		// Run for ever since -1 is decremented
		tinylib.RunServer(tinylib.ReverseEndianness(*initPtr), *portsPtr, -1)
		fmt.Println("AES Server terminated")
	case *ctrPtr && *bobPtr:
		cipher, ivUsed := tinylib.AESCTR(*initPtr, *addrPtr, *portsPtr, *customIv)
		fmt.Println("Data encrypted in CTR mode as:", cipher)
		fmt.Println("with", ivUsed, "as an iv.")
	case *cbcPtr && *bobPtr:
		cipher, ivUsed := tinylib.AESCBC(*initPtr, *addrPtr, *portsPtr, "")
		fmt.Println("Data encrypted in CBC mode as:", cipher)
		fmt.Println("with", ivUsed, "as an iv.")
	case *alicePtr:
		tinylib.YaoServer(*initPtr, *portsPtr)
	case *bobPtr:
            ret := tinylib.YaoClient(*initPtr, *addrPtr, *portsPtr)
            fmt.Println("Client's return value:", ret)
	default: // if running neither as Alice, nor as Bob, there is a misuse
		log.Fatal("Please run as server Alice (-a) first and then as Bob (-b). Use -h for help.")
	}
}
