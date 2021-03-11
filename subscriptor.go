package transport

import (
	"errors"
	"strings"
)

//Subscriptor subscriptor structure
type Subscriptor struct {
	ID           string            `json:"id,omitempty"`
	Enable       bool              `json:"enable"`
	Name         string            `json:"name,omitempty"`
	Path         string            `json:"path,omitempty"`
	SubscriberID string            `json:"subscriberId,omitempty"`
	Properties   map[string]string `json:"properties"`
}

func (subscriptor Subscriptor) valid() error {
	if subscriptor.Name == "" || strings.TrimSpace(subscriptor.Name) == "" {
		return errors.New("Subscriptor must have a name")
	}
	if subscriptor.SubscriberID == "" || strings.TrimSpace(subscriptor.SubscriberID) == "" {
		return errors.New("Subscriptor must have a subscriber")
	}
	_, ok := config.Subscribers[subscriptor.SubscriberID]
	if !ok {
		return errors.New("Subscriptor subscriber does not exist")
	}

	return nil
}

//SendReport --send a report
func (subscriptor Subscriptor) SendReport(report interface{}) {
	config.mu.RLock()
	s, ok := config.Subscribers[subscriptor.SubscriberID]
	config.mu.RUnlock()
	if !ok {
		lg.Error("Subscriptor subscriber does not exist")
		return
	}
	if s.Provider == nil {
		return
	}
	s.Provider.Publish(subscriptor.Path, report)
}
