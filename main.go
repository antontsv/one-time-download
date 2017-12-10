package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/antontsv/one-time-download/limfs"
)

const (
	defaultFileDir          = "files"
	defaultMaxDownloadTimes = 1
	defaultBindAddress      = "localhost:8080"
	bindAddressEnv          = "BIND_ADDRESS"
)

func getAddress(defaultAddress string) string {

	addrStr := os.Getenv(bindAddressEnv)
	parts := strings.Split(addrStr, ":")

	if len(parts) == 2 {
		port, err := strconv.Atoi(parts[1])
		if err == nil && port > 0 && port < 65535 {
			return addrStr
		}
	}
	return defaultAddress
}

func main() {

	bindAddr := flag.String("bind_address", defaultBindAddress,
		fmt.Sprintf(`address and port that server should be listening on. 
	Can be also set using %s env variable`, bindAddressEnv))
	fileDir := flag.String("file_dir", defaultFileDir, "path to directory with files to be served")
	maxTimes := flag.Int("max_times", defaultMaxDownloadTimes, "max times to allow download of a particular file")

	flag.Parse()

	limitedHandler := limfs.New(*fileDir, *maxTimes)
	limitedHandler.DisallowAccess("README.md")

	http.Handle("/", limitedHandler)
	address := getAddress(*bindAddr)
	log.Printf("Starting server on %s...\n", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatal(err)
	}
}
