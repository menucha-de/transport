package transport

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	guuid "github.com/google/uuid"
)

//SubscriberConfiguration subscriber configuration
type SubscriberConfiguration struct {
	Subscribers map[string]*Subscriber `json:"subscribers,omitempty"`
	mu          sync.RWMutex
}

func (c *SubscriberConfiguration) add(sub Subscriber) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Subscribers == nil {
		c.Subscribers = make(map[string]*Subscriber)
	}
	if sub.ID != "" {
		return "", errors.New("Subscriber ID must not be set ")
	}
	if sub.URI == "" {
		return "", errors.New("Subscriber Path must  be set ")
	}
	var id guuid.UUID
	for {
		id = guuid.New()
		_, ok := c.Subscribers[id.String()]
		if !ok {
			sub.ID = id.String()
			break
		}

	}

	if sub.Enable {
		_, err := sub.newProvider()
		if err != nil {
			return "", err
		}
	}
	c.Subscribers[id.String()] = &sub
	c.serialize()
	return id.String(), nil
}
func (c *SubscriberConfiguration) update(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	old, ok := c.Subscribers[id]
	if !ok {
		return errors.New("Subscriber with ID " + id + " does not exist")
	}
	if old.Enable {
		if old.Provider != nil {
			err := old.Provider.SetTLS(id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (c *SubscriberConfiguration) set(sub Subscriber) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	old, ok := c.Subscribers[sub.ID]
	if len(subEnabled[sub.ID]) > 0 {
		return errors.New("Subscriber is locked")
	}
	if !ok {
		return errors.New("Subscriber with ID " + sub.ID + " does not exist")

	}

	if sub.Enable {
		if old.Provider != nil {
			old.Provider.Shutdown()
		}
		flag, err := sub.newProvider()
		if err != nil {
			return err
		}
		if flag {
			err := sub.Provider.SetTLS(sub.ID)
			if err != nil {
				return err
			}
		}
	}
	c.Subscribers[sub.ID] = &sub
	c.serialize()
	return nil
}
func (c *SubscriberConfiguration) delete(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	sub, ok := c.Subscribers[id]
	if len(subs[id]) > 0 {
		return errors.New("Subscriber is in use")
	}
	if !ok {
		return errors.New("Subscriber with ID " + id + " does not exist")

	}
	if sub.Provider != nil {
		sub.Provider.Shutdown()
	}
	delete(c.Subscribers, id)
	err := os.RemoveAll(certFolder + "/" + id)
	if err != nil {
		lg.WithError(err).Warning("Faied to delete certificate folder")
	}
	//stop connection
	c.serialize()
	return nil
}
func (c *SubscriberConfiguration) serialize() {
	f, err := os.Create(filename)
	if err != nil {
		lg.WithError(err).Error("Failed to create or open configuration file")
	} else {
		enc := json.NewEncoder(f)
		enc.SetIndent("", " ")
		enc.Encode(c)
	}
	defer f.Close()
}
