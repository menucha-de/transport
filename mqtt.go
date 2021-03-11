package transport

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type client struct {
	mclient     mqtt.Client
	topic       string
	qos         int
	isConnected bool
	opts        *mqtt.ClientOptions
	timeout     int
}

var defaultTimeout int = 10000
var defaultPort string = "1883"
var defaultQueueSize int = 10
var mqttTimeoutProperty string = prefix + "MQTT.Timeout"

func newMqttProvider(s Subscriber) (*client, error) {
	if s.URI == "" {
		lg.Error("URI Can't be null")
		return nil, errors.New("URI must not be null")
	}
	u, err := url.Parse(s.URI)
	if err != nil {
		lg.Error("Can't parse subscriber URI")
		return nil, err
	}
	if u.Path == "" {
		lg.Error("MQTT topic must be specified using the path of the URI")
		return nil, errors.New("MQTT topic must be specified using the path of the URI")
	}
	topic := u.Path

	if strings.Contains(topic, "#") || strings.Contains(topic, "+") {
		lg.Error("MQTT topic should not contain '#' or '+'")
		return nil, errors.New("MQTT topic should not contain '#' or '+'")

	}
	if u.Host == "" {
		lg.Error("No MQTT host specified")
		return nil, errors.New("No MQTT host specified")

	}
	if u.Port() == "" || u.Port() == "0" {
		u.Host = u.Host + ":" + defaultPort
	}
	values, _ := url.ParseQuery(u.RawQuery)

	id := values.Get("clientid")
	if id == "" {
		lg.Error("clientid must be set as URI query parameter for MQTT transporter")
		return nil, errors.New("clientid must be set as URI query parameter for MQTT transporter")
	}
	runes := []rune(topic)
	topic = string(runes[1:])
	qos := values.Get("qos")
	var nr int
	if qos != "" {
		nr, err = strconv.Atoi(qos)
		if err != nil {
			lg.Error("Invalid MQTT qos value '" + qos + "'")
			return nil, errors.New("Invalid MQTT qos value '" + qos + "'")
		}
	}
	timeout := defaultTimeout
	queueSizeprop := defaultQueueSize
	if s.Properties != nil {
		for key, property := range s.Properties {

			if key != "" {

				if strings.HasPrefix(key, prefix+"MQTT") {
					switch key {
					case mqttTimeoutProperty:
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
				} else if strings.HasPrefix(key, prefix) {
					switch key {
					case queueSize:
						queueSizeprop, err = strconv.Atoi(property)
						if err != nil {
							lg.Error("Invalid queuesize value '" + property + "'")
							return nil, errors.New("Invalid queuesize value '" + property + "'")

						}

						if queueSizeprop <= 0 {
							lg.Error("Invalid queuesize value '" + property + "'")
							return nil, errors.New("Invalid queuesize value '" + property + "'")

						}

					default:
						lg.Error("Unknown property key '" + key + "'")
						return nil, errors.New("Unknown property key '" + key + "'")
					}
				} else {
					lg.Error("Unknown property key '" + key + "'")
					return nil, errors.New("Unknown property key '" + key + "'")

				}
			}
		}
	}
	opts := mqtt.NewClientOptions()
	switch u.Scheme {
	case "mqtt":
		opts.AddBroker("tcp://" + u.Host)
	case "mqtts":
		opts.AddBroker("ssl://" + u.Host)

	default:
		lg.Error("Unknown scheme '" + u.Scheme + "'")
		return nil, errors.New("Unknown scheme '" + u.Scheme + "'")

	}
	opts.SetClientID(id)
	opts.SetUsername(u.User.Username())

	opts.SetConnectTimeout(time.Duration(timeout/1000) * time.Second)
	opts.SetAutoReconnect(true)
	if nr > 0 {
		opts.SetCleanSession(false)
	}
	opts.SetStore(NewMemoryStore(queueSizeprop))
	password, _ := u.User.Password()
	opts.SetPassword(password)
	cl := &client{nil, topic, nr, true, opts, timeout}
	opts.SetOnConnectHandler(cl.onConnect)
	opts.SetConnectionLostHandler(cl.onLost)
	opts.SetMaxReconnectInterval(30 * time.Second)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(2 * time.Second)
	cl.opts = opts
	mclient := mqtt.NewClient(opts)

	cl.mclient = mclient
	if u.Scheme == "mqtt" {
		if token := mclient.Connect(); token.WaitTimeout(time.Duration(timeout/1000)*time.Second) && token.Error() == nil {

		} else {
			lg.Error("Connection failed")
			return cl, errors.New("Can't connect to mqtt at " + u.Host)
		}
	}
	return cl, err

}

func (cl *client) onLost(c mqtt.Client, err error) {
	lg.Error("Connection lost " + fmt.Sprint(err))

}

func (cl *client) onConnect(c mqtt.Client) {

	if c.IsConnectionOpen() {
		lg.Info("Connection established")
	}
}

func (cl *client) Publish(topic string, message interface{}) {
	if cl == nil {
		return
	}

	//if !cl.mclient.IsConnectionOpen() {
	//we don't need this but maybe implement a que
	//	return
	//}

	t := cl.topic
	if topic != "" {
		if cl.topic != "" {
			t = t + "/" + topic
		} else {
			t = topic
		}
	}

	/*if !cl.mclient.IsConnected() {
		lg.Error("disconnected from broker")
		if token := cl.mclient.Connect(); token.Wait() && token.Error() != nil {
			//lg.Error(token.Error.Error())
			return
		}
	}*/
	//avoid blocking
	cl.mclient.Publish(t, byte(cl.qos), false, message)
	//if token := cl.mclient.Publish(t, byte(cl.qos), false, message); token.Wait() && token.Error() != nil {
	//lg.Error(token.Error())
	//	lg.Error("There is an Error")
	//}
}
func (cl *client) Shutdown() {
	if cl != nil && cl.mclient.IsConnected() && cl.mclient.IsConnectionOpen() {
		cl.mclient.Disconnect(1)
	}
}
func (cl *client) SetTLS(id string) error {
	tlsconfig := newTLSConfig(dirname + "/certs/" + id)
	cl.opts.SetTLSConfig(tlsconfig)
	cl.mclient = mqtt.NewClient(cl.opts)
	if token := cl.mclient.Connect(); token.WaitTimeout(time.Duration(cl.timeout/1000)*time.Second) && token.Error() == nil {

	} else {
		lg.Error("Connection failed")
		return errors.New("Can't connect to mqtt ")
	}
	return nil
}
