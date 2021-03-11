package transport

import (
	"bytes"
	"context"
	tls "crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type httpclient struct {
	mclient  *http.Client
	method   string
	URI      string
	mimeType string
	user     *url.Userinfo
	timeout  int
	err      int32
}

var httpTimeoutProperty string = prefix + "HTTP.Timeout"

var httpMethodProperty string = prefix + "HTTP.Method"
var httpsBypassSSlVerificationProperty = prefix + "HTTPS.BypassSSLVerification"
var mimeTypeProperty string = "MimeType"

var defaultmimeType string = "application/json"
var defaultbypassSslVerification bool = false
var defaultmethod string = "POST"

func newHTTPProvider(s Subscriber) (*httpclient, error) {
	defaultTimeout = 1000
	if s.URI == "" {
		lg.Error("URI Can't be null")
		return nil, errors.New("URI must not be null")
	}
	u, err := url.Parse(s.URI)
	if err != nil {
		lg.Error("Can't parse subscriber URI")
		return nil, err
	}

	if u.Host == "" {
		lg.Error("No host specified")
		return nil, errors.New("No  host specified")

	}
	method := defaultmethod
	timeout := defaultTimeout
	bypassSslVerification := defaultbypassSslVerification
	mimeType := defaultmimeType
	if s.Properties != nil {
		for key, property := range s.Properties {

			if key != "" && strings.HasPrefix(key, prefix+"HTTP") {
				switch key {
				case httpTimeoutProperty:
					timeout, err := strconv.Atoi(property)
					if err != nil {
						lg.Error("Invalid timeout value '" + property + "'")
						return nil, errors.New("Invalid timeout value '" + property + "'")

					}

					if timeout < 0 {
						lg.Error("Invalid timeout value '" + property + "'")
						return nil, errors.New("Invalid timeout value '" + property + "'")

					}
				case httpMethodProperty:
					method = strings.ToUpper(property)
					if method == "" {
						lg.Error("HTTP method not specified")
						return nil, errors.New("HTTP method not specified")
					}
					switch method {
					case http.MethodDelete:
					case http.MethodGet:
					case http.MethodHead:
					case http.MethodOptions:
					case http.MethodPost:
					case http.MethodPut:

					default:
						lg.Error("Invalid HTTP method value '" + property + "'")
						return nil, errors.New("Invalid HTTP method value '" + property + "'")
					}

				case httpsBypassSSlVerificationProperty:
					b, err := strconv.ParseBool(property)
					if err != nil {
						lg.Error("Invalid " + httpsBypassSSlVerificationProperty + " '" + property + "'")
						return nil, errors.New("Invalid " + httpsBypassSSlVerificationProperty + " '" + property + "'")

					}
					bypassSslVerification = b

				default:
					lg.Error("Unknown property key '" + key + "'")
					return nil, errors.New("Unknown property key '" + key + "'")
				}
			} else if key == mimeTypeProperty {
				mimeType = property
			}
		}
	}
	cookieJar, _ := cookiejar.New(nil)
	var tlsConfig *tls.Config

	//tlsConfig := &tls.Config{}
	if bypassSslVerification {
		if tlsConfig == nil {
			tlsConfig = &tls.Config{}
		}
		tlsConfig.InsecureSkipVerify = true
	}

	mclient := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		Jar:     cookieJar,
	}
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    time.Duration(timeout) * time.Millisecond,
		DisableCompression: true,
		TLSClientConfig:    tlsConfig,
	}
	mclient.Transport = tr
	return &httpclient{mclient, method, s.URI, mimeType, u.User, timeout, 0}, nil

}
func (c *httpclient) Publish(topic string, message interface{}) {
	//data := url.Values{}
	str := []byte(fmt.Sprintf("%v", message))
	//data.Set("report", str)

	req, err := http.NewRequest(c.method, c.URI, bytes.NewBuffer(str))
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Duration(c.timeout)*time.Millisecond))
	defer cancel()
	if err != nil {
		lg.Error(err.Error())
		return
	}
	req.Header.Set("Content-Type", c.mimeType)
	if c.user != nil {
		a := c.user.Username()
		b, _ := c.user.Password()
		if a != "" && b != "" {
			req.SetBasicAuth(a, b)
		}
	}
	req = req.WithContext(ctx)

	resp, err := c.mclient.Do(req)
	if err != nil {
		if atomic.LoadInt32(&c.err) == 0 {
			atomic.StoreInt32(&c.err, 1)
			lg.Error(err.Error())
		}
		return
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusOK+100 {
		err = errors.New("HTTP " + strconv.Itoa(resp.StatusCode) + ": " + resp.Status)
		if atomic.LoadInt32(&c.err) == 0 {
			atomic.StoreInt32(&c.err, 1)
			lg.Error(err.Error())
		}
		return
	}
	atomic.StoreInt32(&c.err, 0)
}
func (c *httpclient) Shutdown() {}
func (c *httpclient) SetTLS(id string) error {

	tlsConfig := newTLSConfig(dirname + "/certs/" + id)
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    time.Duration(c.timeout) * time.Second,
		DisableCompression: true,
		TLSClientConfig:    tlsConfig,
	}
	c.mclient.Transport = tr
	return nil
}
