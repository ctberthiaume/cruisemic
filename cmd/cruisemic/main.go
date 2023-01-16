package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ctberthiaume/cruisemic/parse"
	"github.com/ctberthiaume/cruisemic/storage"
)

var version = "v0.6.3"

var nameFlag = flag.String("name", "", "Cruise or experiment name (required)")
var noCleanFlag = flag.Bool("noclean", false, "Don't filter for whitelisted ASCII characters: Space to ~, TAB, LF, CR")
var rawFlag = flag.Bool("raw", false, "Save raw, unparsed, but possibly cleaned, input to storage")
var dirFlag = flag.String("dir", "", "Append received data to files in this directory (required)")
var intervalFlag = flag.Duration("interval", 0, "Per-feed throttling interval as duration parsed by time.ParseDuration, e.g. 300ms, 1s, 1m")
var parserFlag = flag.String("parser", "", "Parser to use, use -choices to see valid choices (required)")
var choicesFlag = flag.Bool("choices", false, "Print Parser choices and exit")
var udpFlag = flag.Bool("udp", false, "Read from UDP, not STDIN")
var hostFlag = flag.String("host", "0.0.0.0", "Interface IP to bind to for UDP")
var portFlag = flag.Uint("port", 1234, "UDP port to bind to")
var bufferFlag = flag.Uint("buffer", 1500, "Max UDP receive buffer size")
var quietFlag = flag.Bool("quiet", false, "Suppress UDP informational status on stderr")
var versionFlag = flag.Bool("version", false, "Print version and exit")
var flushFlag = flag.Bool("flush", false, "Flush data to disk after every parsed feed line")

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s\n", version)
		os.Exit(0)
	}
	if *choicesFlag {
		fmt.Printf("Choices for -parser option are:\n%v\n", parse.RegistryChoices())
		os.Exit(0)
	}
	if *nameFlag == "" {
		fmt.Println("-name is required")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *dirFlag == "" {
		fmt.Println("-dir is required")
	}

	parserFact, ok := parse.ParserRegistry[*parserFlag]
	if !ok {
		fmt.Println("-parser must be one of the choices listed by -choices")
		os.Exit(1)
	}
	parser := parserFact(*nameFlag, *intervalFlag, time.Now)
	outPrefix := *nameFlag + "-"
	outSuffix := ".tab"

	// Set header for parsed underway data file and raw data file
	feedHeaders := map[string]string{parse.UnderwayName: parser.Header()}
	if *rawFlag {
		feedHeaders[parse.RawName] = ""
	}

	storer, err := storage.NewDiskStorage(*dirFlag, outPrefix, outSuffix, feedHeaders, 0)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	if *flushFlag {
		err := storer.Flush()
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
	}

	// Handle sigint sigterm, make sure data is flushed, files are closed
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	var mut sync.Mutex // to guard close/flush race conditions
	go func() {
		<-sigs
		mut.Lock()
		if err = storer.Close(); err != nil {
			log.Printf("error: %v\n", err)
		}
		mut.Unlock()
		os.Exit(1)
	}()

	log.Printf("Writing to %q", *dirFlag)
	exitcode := 0
	if *udpFlag {
		if !*quietFlag {
			log.Printf("Starting cruisemic at %v:%v", *hostFlag, *portFlag)
		}
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%d", *hostFlag, *portFlag))
		if err != nil {
			log.Panic(err)
		}

		l, err := net.ListenUDP("udp", addr)
		if err != nil {
			log.Panic(err)
		}
		b := make([]byte, *bufferFlag)
		for {
			n, addr, err := l.ReadFromUDP(b)
			if err != nil {
				log.Printf("read from UDP failed, err: %v", err)
				exitcode = 1
				break
			}
			if !*quietFlag {
				log.Printf("Read from client(%v:%v), len: %v\n", addr.IP, addr.Port, n)
			}
			err = parse.ParseLines(parser, strings.NewReader(string(b[:n])), storer, *rawFlag, *flushFlag, *noCleanFlag)
			if err != nil {
				log.Println(err)
				exitcode = 1
				break
			}
		}
	} else {
		err := parse.ParseLines(parser, os.Stdin, storer, *rawFlag, *flushFlag, *noCleanFlag)
		if err != nil {
			log.Println(err)
			exitcode = 1
		}
	}

	// Exit code for non-signal-intiated exits
	mut.Lock()
	if err = storer.Close(); err != nil {
		log.Printf("error: %v\n", err)
		exitcode = 1
	}
	mut.Unlock()
	os.Exit(exitcode)
}
