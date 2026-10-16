package main

// A real TLS handshake over an in-memory pipe. The stub crypto/tls has a
// handshake that does nothing, so it cannot pass this.

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"math/big"
	"net"
	"time"
)

func main() {
	cert, pool := selfSigned()

	// A client that trusts the certificate completes the handshake and
	// exchanges data.
	client, server := net.Pipe()
	go serve(server, cert)
	conn := tls.Client(client, &tls.Config{RootCAs: pool, ServerName: "tinygo.test"})
	if err := conn.Handshake(); err != nil {
		println("handshake failed:", err.Error())
		return
	}
	if v := conn.ConnectionState().Version; v < tls.VersionTLS12 {
		println("negotiated an unexpected version:", v)
		return
	}
	if _, err := conn.Write([]byte("ping")); err != nil {
		println("write failed:", err.Error())
		return
	}
	buf := make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err != nil {
		println("read failed:", err.Error())
		return
	}
	println("got:", string(buf))
	conn.Close()

	// A client that does not trust the certificate must refuse it.
	client, server = net.Pipe()
	go serve(server, cert)
	conn = tls.Client(client, &tls.Config{ServerName: "tinygo.test"})
	if err := conn.Handshake(); err == nil {
		println("an unknown certificate was accepted")
		return
	}
	conn.Close()
	println("unknown certificate refused")
}

func serve(conn net.Conn, cert tls.Certificate) {
	server := tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{cert}})
	if err := server.Handshake(); err != nil {
		conn.Close()
		return
	}
	buf := make([]byte, 4)
	if _, err := io.ReadFull(server, buf); err != nil {
		server.Close()
		return
	}
	server.Write([]byte("pong"))
}

func selfSigned() (tls.Certificate, *x509.CertPool) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "tinygo.test"},
		DNSNames:              []string{"tinygo.test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		panic(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(leaf)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: leaf}, pool
}
