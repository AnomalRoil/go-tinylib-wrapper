# Project PA
This is the repository for my semester project dedicated to Yao's garbled circuits. It just aimed at understanding the latters and using them in practice. I ended up creating a wrapper in Golang around the [TinyGarble](https://github.com/esonghori/TinyGarble) CLI tool, to allow easier usage of it.  

## TinyGarble Wrapper 
This wrapper consists in a library allowing to use the basic features of TinyGarble in your program through two methods:

    func YaoServer(data string, port int)

allows to use TinyGarble with server argument on the given port.

The client can be used with 

    func YaoClient(data string, addr string, port int)

Both methods requires that the TinyGarble path, circuit path and number of clock cycles needed by the circuit were first setted using 

    func SetCircuit(tiPath string, ciPath string, clCycles int)

### Other features
I also implemented some other features, which are not just wrapping around TinyGarble. For example if you want to, you can use the AES circuits provided with TinyGarble to perform AES CBC or AES CTR encryption using the following methods, for CBC mode:
    
    func AESCBC(data string, addr string, port int, iv string) ([]string, string)

where the data may be any hexadecimal string representing the data to be encrypted. As of now their must be at least 128 bits of data. The address and port should be the IP and port of the AESServer. This methods uses ciphertext stealing to avoid the need for padding.
And for CTR mode :

    func AESCTR(data string, addr string, port int, iv string) ([]string, string)

where the data may be any hexadecimal string of any length. CTR mode doesn't requires anypadding. The address and port should be those of the AESServer.
In order to be able to run those client function, a server should be listenning on the same ports at each step. The following function allows to run a server which will stop after a given number of *rounds*:

    func AESServer(key string, startingPort int, rounds int)

The starting port is incremented each time a new block is to be encrypted, so it should be in a range where the next ports are free. As of now, the wrapper is not yet able to establish a TCP session and then to dynamically choose which port he uses for the next block encryption. This may be a useful future extension.

## Example program
I also provided an exemple program mimicking the current TinyGarble CLI.
Usage :

    $ ./tinygarble -h
        Usage of ./tinygarble:
          -a=false: Run as server Alice
          -b=false: Run as client Bob
          -c="$TINYGARBLE/scd/netlists": location of the circuit root directory. Default : $TINYGARBLE/scd/netlists
          -cc=1: number of clock cycles needed for this circuit, usually 1, usually indicated at the end of the circuit name, sha3_24cc needs 24 clock cycles for example
          -ctr=false: Run using CTR mode and aes circuit in 1cc
          -d="00000000000000000000000000000000": Init data
          -n="aes_1cc.scd": name of the circuit file located in the circuit root directory
          -p=1234: Specify a starting port
          -r="/home/lery/Code/TinyGarble": TinyGarble root directory path, default to $TINYGARBLE if var set, writes $TINYGARBLE if changed
          -s="127.0.0.1": Specify a server address for Bob to connect.

So if you want to run the hamming distance circuit provided with TinyGarble, you may for instance use :

