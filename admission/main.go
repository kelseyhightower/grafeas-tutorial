package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	keyRingFile   string
)

func main() {
	flag.StringVar(&keyRingFile, "keyring", "", "Path to keyring")
	flag.Parse()


	data, err := base64.StdEncoding.DecodeString(signatureBytes)
	if err != nil {
		log.Fatal(err)
	}
}
