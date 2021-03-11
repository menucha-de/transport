package transport

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/amenzhinsky/iothub/iotdevice"
	iotmqtt "github.com/amenzhinsky/iothub/iotdevice/transport/mqtt"
)

type azureclient struct {
	client  *iotdevice.Client
	timeout int
}

var azureTimeoutProperty string = prefix + "Azure.Timeout"
var azureOnDemand string = prefix + "Azure.OnDemand"

func newAzureProvider(s Subscriber) (*azureclient, error) {

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
	connectDirectly := true
	connectionString := u.Host
	timeout := defaultTimeout

	if s.Properties != nil {
		for key, property := range s.Properties {
			if key != "" && strings.HasPrefix(key, prefix+"Azure") {
				switch key {
				case azureTimeoutProperty:
					timeout, err := strconv.Atoi(property)
					if err != nil {
						lg.Error("Invalid timeout value '" + property + "'")
						return nil, errors.New("Invalid timeout value '" + property + "'")

					}

					if timeout < 0 {
						lg.Error("Invalid timeout value '" + property + "'")
						return nil, errors.New("Invalid timeout value '" + property + "'")

					}
				case azureOnDemand:
					b, err := strconv.ParseBool(property)
					if err != nil {
						lg.Error("Invalid " + azureOnDemand + " '" + property + "'")
						return nil, errors.New("Invalid " + azureOnDemand + " '" + property + "'")

					}
					connectDirectly = b

				default:
					lg.Error("Unknown property key '" + key + "'")
					return nil, errors.New("Unknown property key '" + key + "'")
				}

			}
		}
	}
	c, err := iotdevice.NewFromConnectionString(
		iotmqtt.New(), connectionString,
	)
	if err != nil {
		lg.Error(err.Error())
		return nil, err
	}
	if connectDirectly {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
		defer cancel()
		if err = c.Connect(ctx); err != nil {
			lg.Error("Failed to connect to azure subscriptor " + connectionString)
			return nil, err
		}

	}

	return &azureclient{client: c, timeout: timeout}, nil
}
func (c *azureclient) Publish(topic string, message interface{}) {
	if err := c.client.Connect(context.Background()); err != nil {
		lg.Error("Failed to connect to azure subscriptor ")

		return
	}
	str := fmt.Sprintf("%v", message)
	// send a device-to-cloud message
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeout)*time.Millisecond)
	defer cancel()

	if err := c.client.SendEvent(ctx, []byte(str),
		iotdevice.WithSendMessageID(genID()),
	); err != nil {
		lg.Error(err.Error())
	}
	//iotservice.WithSendExpiryTime(time.Now().Add(c.timeout*time.Second))
	c.client.Close()
}
func (c *azureclient) Shutdown() {
	if c != nil {
		c.client.Close()
	}
}
func (c *azureclient) SetTLS(id string) error {
	return nil
}
func genID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
