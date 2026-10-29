package tls_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newKeyPEM(t *testing.T) (*ecdsa.PrivateKey, []byte) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return key, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

func newPair(t *testing.T) (certPEM, keyPEM []byte) {
	key, keyPEM := newKeyPEM(t)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), keyPEM
}

func TestX509KeyPair(t *testing.T) {
	certPEM, keyPEM := newPair(t)
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}
	if cert.Leaf == nil || len(cert.Certificate) != 1 {
		t.Errorf("got %d certificates and Leaf %v", len(cert.Certificate), cert.Leaf)
	}

	_, otherPEM := newKeyPEM(t)
	if _, err := tls.X509KeyPair(certPEM, otherPEM); err == nil {
		t.Error("a mismatched key was accepted")
	}
}

func TestLoadX509KeyPair(t *testing.T) {
	dir, err := os.MkdirTemp("", "keypair")
	if err != nil {
		t.Skip(err)
	}
	defer os.RemoveAll(dir)
	certPEM, keyPEM := newPair(t)
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certFile, certPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := tls.LoadX509KeyPair(certFile, keyFile); err != nil {
		t.Error(err)
	}
}

func TestListenNotImplemented(t *testing.T) {
	if _, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{}); err == nil {
		t.Error("Listen returned no error")
	}
}
