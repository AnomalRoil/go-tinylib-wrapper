package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var tinyPath string
var circuitPath string

// just an example on how to use Stdin, for reference. Coming straight from the Go Docs
func ExampleCommand() {
	cmd := exec.Command("tr", "a-z", "A-Z")
	cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("in all caps: %q\n", out.String())
}

func ReverseEndianness(data string) string {
	//initalizing the return value as an empty string
	ans := ""

	//trimming since there are easily \n in cmd lines outputs.
	data = strings.TrimSpace(data)
	if len(data)%2 != 0 {
		log.Printf("You can't change the endianness of a string whose length isn't a multiple of 2. Please check your data.")
		return data
	}

	//if the data isn't in hex format, i.e. if it hasn't an even number of char, then the programmer made some mistake... Don't need to check if %2==0
	for len(data) >= 2 {
		ans += data[len(data)-2:]
		//fmt.Printf(ans+"\n")
		data = data[:len(data)-2]
	}

	return ans
}

func alice(data string, port string, clock_cycles int) {
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

func bob(data string, addr string, port string, clock_cycles int) {
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
		inputArg, data[:32], "-s", addr, "-p", port, "--output_mode", "2").Output()
	// We specify the --output_mode arg to be "last_clock" only, since otherwise it would output each clock cycle intermediate states.
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Bob output:  %s\n", endian(string(out)))
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

	initPtr := flag.String("d", "00000000000000000000000000000000", "Init data")

	flag.Parse()

	// Checking the remaining flag used : if there are unknown flags, we stop.
	if len(flag.Args()) > 0 {
		log.Fatal("Please use the good arguments. Use -h for help.")
	}

	// We now initialize the TinyGarble root path
	tinyPath = *rootPtr
	if tinyPath == "" {
		fmt.Printf("$TINYGARBLE not set. Assuming TinyGarble is located in $HOME/Code/TinyGarble\n")
		home := os.Getenv("HOME")
		if home != "" {
			tinyPath = home + "/Code/TinyGarble"
			// TODO: check if path exists (exec.Command does it, but it may be better to perform the check here)
		} else {
			log.Fatal("$HOME is not set. Please provide path to TinyGarble's root as argument or set $TINYGARBLE env var to its path.")
		}
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
	case *alicePtr:
		alice(*initPtr, *portsPtr, *clockcyclesPtr)
	case *bobPtr:
		bob(*initPtr, *addrPtr, *portsPtr, *clockcyclesPtr)
	default: // if running neither as Alice, nor as Bob, there is a misuse
		log.Fatal("Please run as server Alice (-a) first and then as Bob (-b). Use -h for help.")
	}
}
