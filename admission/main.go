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
	signatureFile string
)

var signatureBytes = `LS0tLS1CRUdJTiBQR1AgU0lHTkVEIE1FU1NBR0UtLS0tLQpIYXNoOiBTSEEyNTYKCnsKICAiaW1hZ2Vfc3VtbWFyeSI6IHsKICAgICJkaWdlc3QiOiAic2hhMjU2OjBkYjJiNTA1NDU2NTE0MjljNmYyZGE5M2NiN2IxMmIwMjZjOWQ1NTA1YjZlOTQ5YWJkOGMzYmVlYjI3YjAzODIiLAogICAgImZ1bGx5X3F1YWxpZmllZF9kaWdlc3QiOiAiZ2NyLmlvL2hpZ2h0b3dlcmxhYnMvamlyYUBzaGEyNTY6MGRiMmI1MDU0NTY1MTQyOWM2ZjJkYTkzY2I3YjEyYjAyNmM5ZDU1MDViNmU5NDlhYmQ4YzNiZWViMjdiMDM4MiIsCiAgICAicmVnaXN0cnkiOiAiZ2NyLmlvIiwKICAgICJyZXBvc2l0b3J5IjogImhpZ2h0b3dlcmxhYnMvamlyYSIKICB9Cn0KLS0tLS1CRUdJTiBQR1AgU0lHTkFUVVJFLS0tLS0KCmlRRkhCQUVCQ0FBeEZpRUVKVnVpY05KQ3JWd3MxMHluVGJsYlZiS3VTb1FGQWxuaVUvc1RIR2x0WVdkbGMwQmwKZUdGdGNHeGxMbU52YlFBS0NSQk51VnRWc3E1S2hQdldCLzBSY0dxdlR4QUpidlpkMmlTRkwwV21HRGNtYzNnYgpQQTBZMVNWbEZ5TGwwcFhzakZwazhWOURmV3kzQWdzY0tsOEdDVWs2TTFEQU9rMlM1NUM1dFFsN241RXprd3EvCkw2NTNScjVaS1ZQUjVMQU1TZEtmM0tQMnMzblRMeC8vSXBqWVpJQnpzQmR5eG14UzRvbjJUNWltRVR5NmhZaWgKTzcwbkFhMDFGZmc5TmE2cFN4SGlrbEZFOVpFaGYxVGtTcUd0aGNqNytJb085c1dHMDROb28wWG1STmpMemIzMQpYekRRZUpScGpQRlNFSzZmRTFqQUVFTWJteFdXamM2bEhyU3JIMFhIUlZHSmNQK1NKYXJtUG5ObVV5bXJHY3JXCk1JbC91ZWxYaC9YdEdKemxwV3VuL1luNjdJQ0pPV1Z6WjhURWVJdkU4SEJGZlZwQll4TEUwWlo0Cj05bkExCi0tLS0tRU5EIFBHUCBTSUdOQVRVUkUtLS0tLQo=`

func main() {
	flag.StringVar(&keyRingFile, "keyring", "", "Path to keyring")
	flag.Parse()

	kr, err := os.Open(keyRingFile)
	if err != nil {
		log.Fatal(err)
	}

	el, err := openpgp.ReadKeyRing(kr)
	if err != nil {
		log.Fatal(err)
	}

	data, err := base64.StdEncoding.DecodeString(signatureBytes)
	if err != nil {
		log.Fatal(err)
	}

	md, err := openpgp.ReadMessage(bytes.NewReader(data), el, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(md)
}
