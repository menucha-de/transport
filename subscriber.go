package transport

import (
	"errors"
	"net/url"
)

//Provider provider interface
type Provider interface {
	Publish(topic string, m interface{})
	Shutdown()
	SetTLS(id string) error
}

//Subscriber subscriber structure
type Subscriber struct {
	ID         string            `json:"id,omitempty"`
	Enable     bool              `json:"enable"`
	URI        string            `json:"uri,omitempty"`
	Properties map[string]string `json:"properties"`
	Provider   Provider          `json:"-"`
}

func (s *Subscriber) newProvider() (bool, error) {
	u, err := url.Parse(s.URI)
	if err != nil {
		lg.Error(err.Error())
		return false, err
	}
	switch u.Scheme {
	case "mqtts":
		p, err := newMqttProvider(*s)
		s.Provider = p
		if err != nil {
			lg.Error(err.Error())
			return false, err
		}
		return true, nil
	case "mqtt":
		p, err := newMqttProvider(*s)
		s.Provider = p
		if err != nil {
			lg.Error(err.Error())
			return false, err
		}
		return false, nil
	case "https":
		p, err := newHTTPProvider(*s)
		if err != nil {
			lg.Error(err.Error())
			return false, err
		}
		s.Provider = p
		return true, nil
	case "http":
		p, err := newHTTPProvider(*s)
		if err != nil {
			lg.Error(err.Error())
			return false, err
		}
		s.Provider = p
		return false, nil
	case "udp":
		fallthrough
	case "tcp":
		p, err := newTCPProvider(*s)
		if err != nil {
			lg.Error(err.Error())
			return false, err
		}
		s.Provider = p
		return false, nil

	case "azure":
		p, err := newAzureProvider(*s)
		if err != nil {
			lg.Error(err.Error())
			return false, err
		}
		s.Provider = p
		return false, nil

	default:
		lg.Error("Unsuported scheme " + u.Scheme)
		return false, errors.New("Unsuported scheme " + u.Scheme)
	}

}
