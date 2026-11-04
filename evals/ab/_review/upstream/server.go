package upstream

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"sync"
)

// Revocations holds the serial numbers of revoked client certificates. The
// security team revokes a certificate through the admin API; it must stop
// working on the next connection.
type Revocations struct {
	mu      sync.RWMutex
	serials map[string]bool
}

// Revoke marks the certificate with serial as revoked.
func (r *Revocations) Revoke(serial string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.serials == nil {
		r.serials = map[string]bool{}
	}
	r.serials[serial] = true
}

func (r *Revocations) revoked(serial string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.serials[serial]
}

// ServerTLS returns the mTLS configuration for the admin listener: clients
// present a certificate from clientCAs that is not revoked.
func ServerTLS(cert tls.Certificate, clientCAs *x509.CertPool, rev *Revocations) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCAs,
		VerifyPeerCertificate: func(_ [][]byte, chains [][]*x509.Certificate) error {
			if len(chains) == 0 || len(chains[0]) == 0 {
				return errors.New("upstream: no verified chain")
			}
			if rev.revoked(chains[0][0].SerialNumber.String()) {
				return errors.New("upstream: client certificate revoked")
			}
			return nil
		},
	}
}
