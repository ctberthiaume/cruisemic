package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ctberthiaume/cruisemic/parse"
	"github.com/ctberthiaume/cruisemic/rawudp"
	"github.com/ctberthiaume/cruisemic/storage"
)

var version = "v0.9.1"

var nameFlag = flag.String("name", "", "Cruise or experiment name (required)")
var noCleanFlag = flag.Bool("noclean", false, "Don't filter for whitelisted ASCII characters: Space to ~, TAB, LF, CR")
var rawFlag = flag.Bool("raw", false, "Save raw, unparsed, but possibly cleaned, input to storage")
var dirFlag = flag.String("dir", "", "Append received data to files in this directory (required)")
var copyDirFlag = flag.String("copy", "", "Periodically (1m) copy parsed data to this directory")
var intervalFlag = flag.Duration("interval", 0, "Per-feed throttling interval as duration parsed by time.ParseDuration, e.g. 300ms, 1s, 1m")
var parserFlag = flag.String("parser", "", "Parser to use, use -choices to see valid choices (required)")
var choicesFlag = flag.Bool("choices", false, "Print Parser choices and exit")
var udpFlag = flag.Bool("udp", false, "Read from UDP, not STDIN")
var hostFlag = flag.String("host", "0.0.0.0", "Interface IP to bind to for UDP")
var portFlag = flag.String("port", "1234", "Comma-separated list of UDP ports to bind to")
var bufferFlag = flag.Uint("buffer", 1500, "Max UDP receive buffer size")
var quietFlag = flag.Bool("quiet", false, "Suppress UDP informational status on stderr")
var versionFlag = flag.Bool("version", false, "Print version and exit")
var flushFlag = flag.Bool("flush", false, "Flush data to disk after every parsed feed line")
var wrappedFlag = flag.Bool("wrapped", false, "STDIN UDP stream payloads are wrapped with RAWUDP headers")

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
	if *wrappedFlag && *udpFlag {
		fmt.Println("-wrapped and -udp cannot both be set")
		os.Exit(1)
	}

	parserFact, ok := parse.ParserRegistry[*parserFlag]
	if !ok {
		fmt.Println("-parser must be one of the choices listed by -choices")
		os.Exit(1)
	}
	parser := parserFact(*nameFlag, *intervalFlag, time.Now)
	outPrefix := *nameFlag + "-"
	outSuffix := ".tab"

	// Set header for parsed underway data file and raw data file. If not UDP,
	// then don't write raw data, assuming we are already reading a raw data
	// file.
	feedHeaders := map[string]string{parse.UnderwayName: parser.Header()}
	if *rawFlag && *udpFlag {
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
		log.Printf("Starting cruisemic, listening at %v on ports %v", *hostFlag, *portFlag)

		ports := strings.Split(*portFlag, ",")
		dataChan := make(chan []byte)
		var wg sync.WaitGroup

		// Read from UDP ports, write to channel
		for _, p := range ports {
			wg.Add(1)
			go func(port string) {
				defer wg.Done()
				addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%s", *hostFlag, port))
				if err != nil {
					log.Printf("Error resolving UDP address for port %v: %v", port, err)
					return
				}

				l, err := net.ListenUDP("udp", addr)
				if err != nil {
					log.Printf("Error listening on port %s: %v", port, err)
					return
				}
				defer l.Close()

				if !*quietFlag {
					log.Printf("Listening on UDP port %s", port)
				}

				b := make([]byte, *bufferFlag)
				for {
					n, addr, err := l.ReadFromUDP(b)
					if err != nil {
						log.Printf("read from UDP port %s failed, err: %v", port, err)
						break
					}
					if !*quietFlag {
						log.Printf("Read from client(%v:%v) on port %s, len: %v\n", addr.IP, addr.Port, port, n)
					}
					// Copy data to avoid race conditions on buffer 'b'
					data := make([]byte, n)
					copy(data, b[:n])
					dataChan <- data
				}
			}(strings.TrimSpace(p))
		}

		// Process data from channel
		go func() {
			for data := range dataChan {
				if *rawFlag {
					// Write UDP payload wrapped with RAWUDP header
					wrapped := rawudp.WrapUDPPayload(rawudp.RealTime{}, data)
					if err := storer.WriteString(parse.RawName, string(wrapped)); err != nil {
						log.Printf("error writing raw UDP: %v", err)
					}
				}

				err = parse.ParseLines(parser, strings.NewReader(string(data)), storer, *flushFlag, *noCleanFlag)
				if err != nil {
					log.Println(err)
					exitcode = 1
					break
				}
			}
		}()

		// Copy parsed data periodically to another directory if requested
		// TODO: pause storer during copy to avoid partial writes
		if *copyDirFlag != "" {
			log.Printf("Copying parsed data to %q every 1 minute", *copyDirFlag)
			go func() {
				ticker := time.NewTicker(1 * time.Minute)
				defer ticker.Stop()
				for range ticker.C {
					mut.Lock()
					for feed := range feedHeaders {
						srcPath := storer.FeedPath(feed)
						relPath, err := filepath.Rel(*dirFlag, srcPath)
						if err != nil {
							log.Printf("error getting relative path: %v", err)
							continue
						}
						dstPath := filepath.Join(*copyDirFlag, relPath)
						if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
							log.Printf("error creating directory for %q: %v", dstPath, err)
							continue
						}
						err = storage.CopyFile(srcPath, dstPath)
						if err != nil {
							log.Printf("error copying %q to %q: %v", srcPath, dstPath, err)
						}
					}
					mut.Unlock()
				}
			}()
		}
		// Wait for all UDP readers to finish (they run forever unless error)
		wg.Wait()
		close(dataChan)
	} else {
		var r io.Reader
		if *wrappedFlag {
			// Read from STDIN with RAWUDP-wrapped payloads
			r = rawudp.NewRawUDPReader(bufio.NewReader(os.Stdin))
		} else {
			r = bufio.NewReader(os.Stdin)
		}

		err := parse.ParseLines(parser, r, storer, *flushFlag, *noCleanFlag)
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
