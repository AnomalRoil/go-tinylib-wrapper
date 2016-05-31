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

	fmt.Println("Starting TinyGarble main example program")

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
	tinylib.SetTinyPaths(tinyPath, circuitPath)

	// sanity check for the input
	if len(*initPtr) < 32 {
		log.Fatal("Please give an init value of length 32 at least until I've implemented padding.")
	}

	// we can continue, everything is initialized.
	switch {
	case *ctrPtr && *alicePtr:
		fmt.Println("Launching AES CTR server with key:", *initPtr)
		tinylib.CTRServer(*initPtr, *portsPtr, -1)
		fmt.Println("AES Server terminated")
	case *ctrPtr && *bobPtr:
		_, ivUsed := tinylib.AESCTR(*initPtr, *addrPtr, *portsPtr, "")
		fmt.Println("Terminated after using as iv:", ivUsed)
	case *alicePtr:
		tinylib.YaoServer(*initPtr, *portsPtr, *clockcyclesPtr)
	case *bobPtr:
		tinylib.YaoClient(*initPtr, *addrPtr, *portsPtr, *clockcyclesPtr)
	default: // if running neither as Alice, nor as Bob, there is a misuse
		log.Fatal("Please run as server Alice (-a) first and then as Bob (-b). Use -h for help.")
	}
}
