package main

import (
	"crypto/tls"
	"log"
	"net/http"

	"validatingwebhook/webhook/mutate"
	"validatingwebhook/webhook/validate"
)

func main() {
	log.Println("Webhook starting on port 9443")

	http.HandleFunc("/validate", validate.Handler)
	http.HandleFunc("/mutate", mutate.Handler)

	certReloader, err := NewCertReloader(
		certificatePath,
		privateKeyPath,
	)

	if err != nil {
		panic(err)
	}

	server := &http.Server{
		Addr: ":9443",
		TLSConfig: &tls.Config{
			GetCertificate: certReloader.GetCertificate,
		},
	}

	log.Fatal(
		server.ListenAndServeTLS("", ""),
	)
}
