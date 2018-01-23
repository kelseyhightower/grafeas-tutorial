// Copyright 2017 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/clearsign"
	"golang.org/x/crypto/openpgp/packet"

	grafeas "github.com/Grafeas/client-go/v1alpha1"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	log.Println(string(data))

	ar := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(data, &ar); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pod := v1.Pod{}
	if err := json.Unmarshal(ar.Request.Object.Raw, &pod); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	admissionResponse := v1beta1.AdmissionResponse{Allowed: false}
	for _, container := range pod.Spec.Containers {
		// Retrieve all occurrences.
		// This call should be replaced by a filtered called based on
		// the container image under review.
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

		// Find a valid signature for the given container image.
		match := false
		for _, occurrence := range occurrencesResponse.Occurrences {
			resourceUrl := occurrence.ResourceUrl
			signature := occurrence.Attestation.PgpSignedAttestation.Signature
			keyId := occurrence.Attestation.PgpSignedAttestation.PgpKeyId

			log.Printf("Container Image: %s", container.Image)
			log.Printf("ResourceUrl: %s", resourceUrl)
			log.Printf("Signature: %s", signature)
			log.Printf("KeyId: %s", keyId)

			if container.Image != strings.TrimPrefix(resourceUrl, "https://") {
				continue
			}

			match = true

			s, err := base64.StdEncoding.DecodeString(signature)
			if err != nil {
				log.Println(err)
				continue
			}

			publicKey := fmt.Sprintf("/etc/admission-controller/pubkeys/%s.pub", keyId)
			log.Printf("Using public key: %s", publicKey)

			f, err := os.Open(publicKey)
			if err != nil {
				log.Println(err)
				continue
			}

			block, err := armor.Decode(f)
			if err != nil {
				log.Println(err)
				continue
			}

			if block.Type != openpgp.PublicKeyType {
				log.Println("Not public key")
				continue
			}

			reader := packet.NewReader(block.Body)
			pkt, err := reader.Next()
			if err != nil {
				log.Println(err)
				continue
			}

			key, ok := pkt.(*packet.PublicKey)
			if !ok {
				log.Println("Not public key")
				continue
			}

			b, _ := clearsign.Decode(s)

			reader = packet.NewReader(b.ArmoredSignature.Body)
			pkt, err = reader.Next()
			if err != nil {
				log.Println(err)
				continue
			}

			sig, ok := pkt.(*packet.Signature)
			if !ok {
				log.Println("Not signature")
				continue
			}

			hash := sig.Hash.New()
			io.Copy(hash, bytes.NewReader(b.Bytes))

			err = key.VerifySignature(hash, sig)
			if err != nil {
				log.Println(err)

				message := fmt.Sprintf("Signature verification failed for container image: %s", container.Image)
				log.Printf(message)

				admissionResponse.Allowed = false
				admissionResponse.Result = &metav1.Status{
					Reason: metav1.StatusReasonInvalid,
					Details: &metav1.StatusDetails{
						Causes: []metav1.StatusCause{
							{Message: message},
						},
					},
				}
				goto done
			}

			log.Printf("Signature verified for container image: %s", container.Image)
			admissionResponse.Allowed = true
		}

		if !match {
			message := fmt.Sprintf("No matched signatures for container image: %s", container.Image)
			log.Printf(message)
			admissionResponse.Allowed = false
			admissionResponse.Result = &metav1.Status{
				Reason: metav1.StatusReasonInvalid,
				Details: &metav1.StatusDetails{
					Causes: []metav1.StatusCause{
						{Message: message},
					},
				},
			}
			goto done
		}
	}

done:
	ar = v1beta1.AdmissionReview{
		Response: &admissionResponse,
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
