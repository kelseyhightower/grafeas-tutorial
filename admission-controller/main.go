package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"

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
	flag.StringVar(&grafeasUrl, "grafeas", "http://127.0.0.1:8080", "The Grafeas server address")
	flag.StringVar(&tlsCertFile, "tls-cert", "/etc/admission-controller/tls/cert.pem", "TLS certificate file.")
	flag.StringVar(&tlsKeyFile, "tls-key", "/etc/admission-controller/tls/key.pem", "TLS key file.")
	flag.Parse()

	http.HandleFunc("/", admissionReviewHandler)
	log.Fatal(http.ListenAndServeTLS(":443", tlsCertFile, tlsKeyFile, nil))
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
		log.Println(container.Image)
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
