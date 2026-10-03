package main

import (
	"crypto/tls"
	"log"
	"sync"
	"time"
)

const (
	certificatePath = "/certs/server.crt"
	privateKeyPath  = "/certs/server.key"
	// so periodically inspect the mounted certificate files.
	certificateReloadInterval = 30 * time.Second
)

// CertReloader owns the TLS certificate currently used by the server.
// certPath and keyPath point to files projected from the Kubernetes Secret.
// cert stores the certificate currently being served.
// mu protects cert from simultaneous reads and writes.
type CertReloader struct {
	certPath string
	keyPath  string
	cert     *tls.Certificate
	mu       sync.RWMutex
}

func NewCertReloader(
	certPath,
	keyPath string,
) (*CertReloader, error) {

	cr := &CertReloader{
		certPath: certPath,
		keyPath:  keyPath,
	}

	if err := cr.reload(); err != nil {
		return nil, err
	}

	go cr.watch()

	return cr, nil
}
func (cr *CertReloader) reload() error {
	cert, err := tls.LoadX509KeyPair(
		cr.certPath,
		cr.keyPath,
	)

	if err != nil {
		return err
	}

	cr.mu.Lock()
	cr.cert = &cert
	cr.mu.Unlock()

	log.Println("Certificate reloaded")

	return nil
}
func (cr *CertReloader) watch() {
	ticker := time.NewTicker(
		1 * time.Hour,
	)

	for range ticker.C {
		if err := cr.reload(); err != nil {
			log.Printf(
				"Failed to reload certificate: %v",
				err,
			)
		}
	}
}
func (cr *CertReloader) GetCertificate(
	_ *tls.ClientHelloInfo,
) (*tls.Certificate, error) {

	cr.mu.RLock()
	defer cr.mu.RUnlock()

	return cr.cert, nil
}
