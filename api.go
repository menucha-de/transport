package transport

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"software.sslmate.com/src/go-pkcs12"

	guuid "github.com/google/uuid"
	"github.com/menucha-de/utils"
)

const certFolder = "./conf/transport/certs"
const trustFileName = "ca"
const certFileName = "cert"
const keystoreFileName = "key"

var secKeys map[string]string

func getSubscribers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	subscribers := []*Subscriber{}
	for _, value := range config.Subscribers {
		subscribers = append(subscribers, value)
	}
	var err = json.NewEncoder(w).Encode(subscribers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func hasTrusted(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "plain/text; charset=UTF-8")
	vars := mux.Vars(r)
	id := vars["id"]
	fileExist := utils.FileExists(certFolder + "/" + id + "/" + trustFileName)
	n, err := io.WriteString(w, strconv.FormatBool(fileExist))
	if err != nil && n > 0 {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func deleteTrusted(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	fileExist := utils.FileExists(certFolder + "/" + id + "/" + trustFileName)
	if fileExist {
		err := os.Remove(certFolder + "/" + id + "/" + trustFileName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
func setTrusted(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if _, err := os.Stat(certFolder + "/" + id); os.IsNotExist(err) {
		os.MkdirAll(certFolder+"/"+id, 0700)
	}
	out, err := os.Create(certFolder + "/" + id + "/" + trustFileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Write the body to file

	_, err = io.Copy(out, r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
func hasKeyStore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	w.Header().Set("Content-Type", "plain/text; charset=UTF-8")
	fileExist := utils.FileExists(certFolder + "/" + id + "/" + keystoreFileName)
	n, err := io.WriteString(w, strconv.FormatBool(fileExist))
	if err != nil && n > 0 {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func setKeyStore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	u, _ := url.Parse(r.URL.String())
	values, _ := url.ParseQuery(u.RawQuery)

	idd := values.Get("secKey")
	passphrase, ok := secKeys[idd]
	if !ok {
		http.Error(w, "No passphrase specified", http.StatusInternalServerError)
		return
	}
	delete(secKeys, idd)
	if _, err := os.Stat(certFolder + "/" + id); os.IsNotExist(err) {
		os.MkdirAll(certFolder+"/"+id, 0700)
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	privateKey, certificate, err := pkcs12.Decode(body, passphrase)
	if err != nil {
		lg.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := verify(certificate); err != nil {
		lg.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//write private key as pem
	priv, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		err = errors.New("expected RSA private key type")
		lg.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	keyFile, err := os.Create(certFolder + "/" + id + "/" + keystoreFileName)
	if err != nil {
		lg.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer keyFile.Close()
	err = pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if err != nil {
		lg.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	certFile, err := os.Create(certFolder + "/" + id + "/" + certFileName)
	if err != nil {
		lg.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer certFile.Close()
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw})
	if err != nil {
		lg.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = config.update(id)
	if err != nil {
		lg.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}
func deleteKeyStore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	fileExist := utils.FileExists(certFolder + "/" + id + "/" + keystoreFileName)
	if fileExist {
		err := os.Remove(certFolder + "/" + id + "/" + keystoreFileName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	fileExist = utils.FileExists(certFolder + "/" + id + "/" + certFileName)
	if fileExist {
		err := os.Remove(certFolder + "/" + id + "/" + certFileName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
func setPassphrase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "plain/text; charset=UTF-8")
	var b bytes.Buffer
	n, err := b.ReadFrom(r.Body)
	if err != nil || n == 0 {
		http.Error(w, "Could not read passphrase value", http.StatusBadRequest)
		return
	}
	secKey := guuid.New().String()
	secKeys[secKey] = b.String()
	nn, err := io.WriteString(w, secKey)
	if err != nil && nn > 0 {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func addSubscriber(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "plain/text; charset=UTF-8")
	var subscriber *Subscriber

	err := utils.DecodeJSONBody(w, r, &subscriber)
	if err != nil {
		var mr *utils.MalformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			lg.WithError(err).Error("Failed to get subscriber")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	id, err := config.add(*subscriber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	subscriber.ID = id

	n, err := io.WriteString(w, id)
	if err != nil && n > 0 {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func getSubscriber(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if strings.TrimSpace(id) == "" {
		http.Error(w, "ID must not be null", http.StatusBadRequest)
		return
	}
	var s *Subscriber
	s, ok := config.Subscribers[id]
	if !ok {
		http.Error(w, "Subscriber with ID "+id+" does not exist", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var err = json.NewEncoder(w).Encode(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

//todo don't update when locked
func setSubscriber(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "plain/text; charset=UTF-8")
	var subscriber *Subscriber
	vars := mux.Vars(r)
	err := utils.DecodeJSONBody(w, r, &subscriber)
	if err != nil {
		var mr *utils.MalformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			lg.WithError(err).Error("Failed to get subscriber")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	id := vars["id"]
	if strings.TrimSpace(id) == "" {
		http.Error(w, "ID must not be null", http.StatusBadRequest)
		return
	}
	if subscriber.ID != id {
		http.Error(w, "ID of subscriber does not match ", http.StatusBadRequest)
		return
	}
	err = config.set(*subscriber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

//todo don't delete when used
func deleteSubscriber(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "plain/text; charset=UTF-8")

	vars := mux.Vars(r)
	id := vars["id"]
	if strings.TrimSpace(id) == "" {
		http.Error(w, "ID must not be null", http.StatusBadRequest)
		return
	}
	err := config.delete(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	w.WriteHeader(http.StatusNoContent)
}
func verify(cert *x509.Certificate) error {
	_, err := cert.Verify(x509.VerifyOptions{})
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case x509.CertificateInvalidError:
		switch e.Reason {
		case x509.Expired:
			return errors.New("certificate has expired or is not yet valid")
		default:
			return err
		}
	case x509.UnknownAuthorityError:
		// Apple cert isn't in the cert pool
		// ignoring this error
		return nil
	default:
		return err
	}
}
