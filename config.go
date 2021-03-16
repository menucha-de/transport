package transport

import (
	"encoding/json"
	"os"

	loglib "github.com/menucha-de/logging"
)

const filename = "./conf/transport/subscribers.json"
const dirname = "./conf/transport"

var lg *loglib.Logger = loglib.GetLogger("transport")
var config *SubscriberConfiguration
var prefix = "Transporter."
var queueSize string = prefix + "ResendQueueSize"

func init() {

	subs = make(map[string]map[string]string)
	subEnabled = make(map[string]map[string]string)
	subscriptors = make(map[string]*Subscriptor)
	secKeys = make(map[string]string)
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		os.MkdirAll(dirname, 0700)

	}
	f, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err == nil {
		dec := json.NewDecoder(f)
		err = dec.Decode(&config)

		if err != nil {
			lg.Warning("Failed to parse config")
			config = initConfiguration()
		}
		for _, v := range config.Subscribers {
			if v.Enable {
				flag, err := v.newProvider()

				if err != nil {
					lg.WithError(err).Warning("Failed to create provider")
				}
				if flag {
					err = v.Provider.SetTLS(v.ID)
					if err != nil {
						lg.WithError(err).Warning("Failed to create provider")
					}
				}
			}
		}

	} else {
		lg.WithError(err).Debug("Failed to read config")
		config = initConfiguration()
	}
	defer f.Close()

}

func initConfiguration() *SubscriberConfiguration {
	subscribers := make(map[string]*Subscriber)
	config := SubscriberConfiguration{Subscribers: subscribers}
	return &config
}
