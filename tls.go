package transport

import (
	tls "crypto/tls"
	"crypto/x509"
	"io/ioutil"

	utils "github.com/peramic/utils"
)

func newTLSConfig(dir string) *tls.Config {
	// Import trusted certificates from CAfile.pem.
	certpool := x509.NewCertPool()

	pemCerts, err := ioutil.ReadFile(dir + "/ca")
	if err == nil {
		certpool.AppendCertsFromPEM(pemCerts)
	}
	var cert tls.Certificate
	// Import client certificate/key pair
	if utils.FileExists(dir + "/key") {
		cert, err = tls.LoadX509KeyPair(dir+"/cert", dir+"/key")
		if err != nil {
			lg.Error(err.Error())
		}
	}

	// Create tls.Config with desired tls properties
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs: certpool,
		// ClientAuth = whether to request cert from server.
		// Since the server is set up for SSL, this happens
		// anyways.
		ClientAuth: tls.NoClientCert,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: false,
		// Certificates = list of certs client sends to server.
		Certificates: []tls.Certificate{cert},
	}
}
