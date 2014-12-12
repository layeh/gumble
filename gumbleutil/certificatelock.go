package gumbleutil

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"

	"github.com/layeh/gumble/gumble"
)

// CertificateLockFile adds a new certificate lock on the given Client and
// Config that ensures that a server's certificate is signed by the same CA
// from connection-to-connection. This is helpful when connecting to servers
// with self-signed certificates.
//
// If filename does not exist, the server's certificate chain will be written
// to that file. If it does exist, certificates will be read from that file and
// added to RootCAs in config's TLSConfig.
//
// Example:
//
//  if firstConnectionToServer {
//      // Allow self-signed certificates to be accepted on the initial
//      // connection.
//      config.TLSConfig.InsecureSkipVerify = true
//  }
//  gumbleutil.CertificateLockFile(client, &config, filename)
//
//  if err := client.Connect(); err != nil {
//      panic(err)
//  }
func CertificateLockFile(client *gumble.Client, config *gumble.Config, filename string) (gumble.Detacher, error) {
	if file, err := os.Open(filename); err == nil {
		defer file.Close()
		if config.TLSConfig.RootCAs == nil {
			config.TLSConfig.RootCAs = x509.NewCertPool()
		}
		if data, err := ioutil.ReadAll(file); err == nil {
			config.TLSConfig.RootCAs.AppendCertsFromPEM(data)
		}
		return nil, nil
	}

	return client.Attach(Listener{
		Connect: func(e *gumble.ConnectEvent) {
			tlsClient, ok := e.Client.Conn().(*tls.Conn)
			if !ok {
				return
			}
			serverCerts := tlsClient.ConnectionState().PeerCertificates
			file, err := os.Create(filename)
			if err != nil {
				return
			}
			block := pem.Block{
				Type: "CERTIFICATE",
			}
			for _, cert := range serverCerts {
				block.Bytes = cert.Raw
				pem.Encode(file, &block)
			}
			file.Close()
		},
	}), nil
}
