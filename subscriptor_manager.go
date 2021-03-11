package transport

import (
	"errors"

	guuid "github.com/google/uuid"
)

var subs map[string]map[string]string
var subEnabled map[string]map[string]string
var subscriptors map[string]*Subscriptor

//AddSubscriptor  add a subscriptor
func AddSubscriptor(sub Subscriptor) error {
	err := sub.valid()
	if err != nil {
		return err
	}

	subscriptors[sub.ID] = &sub
	if subs[sub.SubscriberID] == nil {
		subs[sub.SubscriberID] = make(map[string]string)
		subEnabled[sub.SubscriberID] = make(map[string]string)
	}
	subs[sub.SubscriberID][sub.ID] = sub.ID
	if sub.Enable {
		subEnabled[sub.SubscriberID][sub.ID] = sub.ID
	}

	return nil

}

//DefineSubscriptor define a subscriptor
func DefineSubscriptor(sub Subscriptor) (string, error) {
	if sub.ID != "" {
		return "", errors.New("Subscriptor ID must not be set ")
	}
	err := sub.valid()
	if err != nil {

		return "", err
	}

	var id guuid.UUID
	for {
		id = guuid.New()
		_, ok := subscriptors[id.String()]
		if !ok {
			sub.ID = id.String()
			break
		}

	}
	subscriptors[sub.ID] = &sub
	if subs[sub.SubscriberID] == nil {
		subs[sub.SubscriberID] = make(map[string]string)
		subEnabled[sub.SubscriberID] = make(map[string]string)
	}
	subs[sub.SubscriberID][sub.ID] = sub.ID
	if sub.Enable {
		subEnabled[sub.SubscriberID][sub.ID] = sub.ID
	}

	return id.String(), nil

}

//UpdateSubscriptor update a subscriptor
func UpdateSubscriptor(sub Subscriptor) error {
	s, ok := subscriptors[sub.ID]
	if !ok {
		return errors.New("Subscriptor with Id " + sub.ID + " has not been initialized")
	}
	if s.Enable && sub.Enable {
		return errors.New("Can't update an used subscriptor")
	}
	err := sub.valid()
	if err != nil {
		lg.Error(err)
		return err
	}
	if s.SubscriberID != sub.SubscriberID {
		delete(subs[s.SubscriberID], s.ID)
		delete(subEnabled[s.SubscriberID], s.ID)
		if subs[sub.SubscriberID] == nil {
			subs[sub.SubscriberID] = make(map[string]string)
			subEnabled[sub.SubscriberID] = make(map[string]string)
		}
		subs[sub.SubscriberID][sub.ID] = sub.ID
	}

	if sub.Enable {
		subEnabled[sub.SubscriberID][sub.ID] = sub.ID
	} else {

		delete(subEnabled[sub.SubscriberID], sub.ID)

	}
	subscriptors[sub.ID] = &sub
	return nil
}

//DeleteSubscriptor deletes a subscriptor
func DeleteSubscriptor(id string) error {
	s, ok := subscriptors[id]
	if !ok {
		return errors.New("Subscriptor with Id " + id + " has not been initialized")
	}
	delete(subs[s.SubscriberID], id)
	delete(subEnabled[s.SubscriberID], id)
	delete(subscriptors, id)
	return nil
}
