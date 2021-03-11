package transport

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type tcpclient struct {
	URI     string
	timeout int
}

func newTCPProvider(s Subscriber) (*tcpclient, error) {
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
	if strings.ToLower(u.Scheme) != "tcp" && strings.ToLower(u.Scheme) != "udp" {
		lg.Error("Unknown scheme ")
		return nil, err
	}
	if u.Host == "" {
		lg.Error("No host specified")
		return nil, errors.New("No  host specified")

	}
	var tcpTimeoutProperty string = prefix + strings.ToUpper(u.Scheme) + ".Timeout"
	if u.Port() == "" {
		lg.Error("No port specified")
		return nil, errors.New("No  port specified")
	}
	timeout := defaultTimeout
	if s.Properties != nil {
		for key, property := range s.Properties {

			if key != "" && strings.HasPrefix(key, prefix+strings.ToUpper(u.Scheme)) {
				switch key {
				case tcpTimeoutProperty:
					timeout, err = strconv.Atoi(property)
					if err != nil {
						lg.Error("Invalid timeout value '" + property + "'")
						return nil, errors.New("Invalid timeout value '" + property + "'")

					}

					if timeout < 0 {
						lg.Error("Invalid timeout value '" + property + "'")
						return nil, errors.New("Invalid timeout value '" + property + "'")

					}

				default:
					lg.Error("Unknown property key '" + key + "'")
					return nil, errors.New("Unknown property key '" + key + "'")

				}
			}
		}
	}

	return &tcpclient{s.URI, timeout}, nil
}
func (c *tcpclient) Publish(topic string, message interface{}) {
	if c == nil {
		return
	}

	str := fmt.Sprintf("%v", message)
	u, _ := url.Parse(c.URI)
	servAddr := u.Host

	d := net.Dialer{Timeout: time.Duration(c.timeout/1000) * time.Second}
	conn, err := d.Dial(strings.ToLower(u.Scheme), servAddr)
	if err != nil {
		lg.Error("Dial failed:", err.Error())
		return
	}

	_, err = conn.Write([]byte(str))
	if err != nil {
		lg.Error("Write to server failed:", err.Error())
		return
	}
	conn.Close()
}
func (c *tcpclient) Shutdown() {}
func (c *tcpclient) SetTLS(id string) error {
	return nil
}
