package gumbleutil

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"

	"github.com/layeh/gumble/gumble"
)

// CertificateLockFile adds a new certificate lock on the given Client that
// ensures that a server's certificate chain is the same from
// connection-to-connection. This is helpful when connecting to servers with
// self-signed certificates.
//
// If filename does not exist, the server's certificate chain will be written
// to that file. If it does exist, certificates will be read from the file and
// checked against the server's certificate chain upon connection.
//
// Example:
//
//  if allowSelfSignedCertificates {
//      config.TLSConfig.InsecureSkipVerify = true
//  }
//  gumbleutil.CertificateLockFile(client, filename)
//
//  if err := client.Connect(); err != nil {
//      panic(err)
//  }
func CertificateLockFile(client *gumble.Client, filename string) {
	client.Config.TLSVerify = func(state *tls.ConnectionState) error {
		if file, err := os.Open(filename); err == nil {
			defer file.Close()
			data, err := ioutil.ReadAll(file)
			if err != nil {
				return err
			}
			i := 0
			for block, data := pem.Decode(data); block != nil; block, data = pem.Decode(data) {
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					return err
				}
				if i >= len(state.PeerCertificates) {
					return errors.New("gumbleutil: invalid certificate chain length")
				}
				if !cert.Equal(state.PeerCertificates[i]) {
					return errors.New("gumbleutil: certificate verification failure")
				}
				i++
			}
			if i != len(state.PeerCertificates) {
				return errors.New("gumbleutil: invalid certificate chain length")
			}
			return nil
		}

		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		block := pem.Block{
			Type: "CERTIFICATE",
		}
		for _, cert := range state.PeerCertificates {
			block.Bytes = cert.Raw
			if err := pem.Encode(file, &block); err != nil {
				return err
			}
		}
		return nil
	}
}
