package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"

	"golang.org/x/crypto/openpgp"
	grafeas "github.com/Grafeas/client-go/v1alpha1"

	"k8s.io/api/admission/v1alpha1"
	"k8s.io/api/core/v1"
)

var (
	grafeasUrl  string
	tlsCertFile string
	tlsKeyFile  string
)

var (
	notesPath       = "/v1alpha1/projects/image-signing/notes"
	occurrencesPath = "/v1alpha1/projects/image-signing/occurrences"
)

func main() {
	flag.StringVar(&grafeasUrl, "grafeas", "http://grafeas:8080", "The Grafeas server address")
	flag.StringVar(&tlsCertFile, "tls-cert", "/etc/admission-controller/tls/cert.pem", "TLS certificate file.")
	flag.StringVar(&tlsKeyFile, "tls-key", "/etc/admission-controller/tls/key.pem", "TLS key file.")
	flag.Parse()

	http.HandleFunc("/", admissionReviewHandler)
	s := http.Server{
		Addr: ":443",
		TLSConfig: &tls.Config{
			ClientAuth: tls.NoClientCert,
		},
	}
	log.Fatal(s.ListenAndServeTLS(tlsCertFile, tlsKeyFile))
}

func admissionReviewHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ar := v1alpha1.AdmissionReview{}
	if err := json.Unmarshal(data, &ar); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pod := v1.Pod{}
	if err := json.Unmarshal(ar.Spec.Object.Raw, &pod); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	admissionReviewStatus := v1alpha1.AdmissionReviewStatus{Allowed: true}
	for _, container := range pod.Spec.Containers {
		u := fmt.Sprintf("%s/%s", grafeasUrl, occurrencesPath)
		resp, err := http.Get(u)
		if err != nil {
			log.Println(err)
			continue
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			resp.Body.Close()
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Printf("non 200 status code: %d", resp.StatusCode)
			continue
		}

		occurrencesResponse := grafeas.ListOccurrencesResponse{}
		if err := json.Unmarshal(data, &occurrencesResponse); err != nil {
			log.Println(err)
			continue
		}

		for _, occurrence := range occurrencesResponse.Occurrences {
			signature := occurrence.Attestation.PgpSignedAttestation.Signature
			keyId := occurrence.Attestation.PgpSignedAttestation.PgpKeyId

			log.Println(container.Image)
			log.Println(signature)
			log.Println(keyId)

			s, err := base64.StdEncoding.DecodeString(signature)
			if err != nil {
				log.Println(err)
				continue
			}

			outFile, err := ioutil.TempFile("", "output")
			if err != nil {
				log.Println(err)
				continue
			}
			outFile.Close()

			tmpfile, err := ioutil.TempFile("", "signature")
			if err != nil {
				log.Println(err)
				continue
			}

			// defer os.Remove(tmpfile.Name())
			// defer os.Remove(outFile.Name())
			log.Println(tmpfile.Name())
			log.Println(outFile.Name())

			if _, err := tmpfile.Write(s); err != nil {
				log.Println(err)
				continue
			}

			if err := tmpfile.Close(); err != nil {
				log.Println(err)
				continue
			}

			keyring := fmt.Sprintf("/etc/admission-controller/keyrings/%s.pub", keyId)

			c := exec.Command("gpg", "--import", keyring)
			o, err := c.CombinedOutput()
			if err != nil {
				log.Println(err)
				log.Println(string(o))
				continue
			}

			cmd := exec.Command("gpg", "--trust-model", "always", "--output", outFile.Name(), tmpfile.Name())

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err = cmd.Run()
			if err != nil {
				log.Println(stderr.String())
				log.Println(err)
				continue
			}
			log.Println(stderr.String())

			out, err := ioutil.ReadFile(outFile.Name())
			if err != nil {
				log.Println(err)
				continue
			}

			log.Println(string(out))
		}
	}

	ar = v1alpha1.AdmissionReview{
		Status: admissionReviewStatus,
	}

	data, err = json.Marshal(ar)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
