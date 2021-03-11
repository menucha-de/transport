/*
 * Copyright (c) 2013 IBM Corp.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v10.html
 *
 * Contributors:
 *    Seth Hoenig
 *    Allan Stockdill-Mander
 *    Mike Robertson
 */

package transport

import (
	"sync"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// MemoryStore implements the store interface to provide a "persistence"
// mechanism wholly stored in memory. This is only useful for
// as long as the client instance exists.
type MemoryStore struct {
	sync.RWMutex
	messages map[string]packets.ControlPacket
	opened   bool
	size     int
}

// NewMemoryStore returns a pointer to a new instance of
// MemoryStore, the instance is not initialized and ready to
// use until Open() has been called on it.
func NewMemoryStore(size int) *MemoryStore {
	store := &MemoryStore{
		messages: make(map[string]packets.ControlPacket),
		opened:   false,
		size:     size,
	}
	return store
}

// Open initializes a MemoryStore instance.
func (store *MemoryStore) Open() {
	store.Lock()
	defer store.Unlock()
	store.opened = true
	lg.Debug("memorystore initialized")
}

// Put takes a key and a pointer to a Message and stores the
// message.
func (store *MemoryStore) Put(key string, message packets.ControlPacket) {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		lg.Error("Trying to use memory store, but not open")
		return
	}
	if len(store.messages) >= store.size {
		lg.Debug("queue size excedeed")
		return
	}
	store.messages[key] = message
}

// Get takes a key and looks in the store for a matching Message
// returning either the Message pointer or nil.
func (store *MemoryStore) Get(key string) packets.ControlPacket {
	store.RLock()
	defer store.RUnlock()
	if !store.opened {
		lg.Error("Trying to use memory store, but not open")
		return nil
	}
	//mid := mIDFromKey(key)
	m := store.messages[key]
	if m == nil {
		lg.Warning("memorystore get: message not found")
	} else {
		lg.Debug("memorystore get: message found")
	}
	return m
}

// All returns a slice of strings containing all the keys currently
// in the MemoryStore.
func (store *MemoryStore) All() []string {
	store.RLock()
	defer store.RUnlock()
	if !store.opened {
		lg.Error("Trying to use memory store, but not open")
		return nil
	}
	var keys []string
	for k := range store.messages {
		keys = append(keys, k)
	}
	return keys
}

// Del takes a key, searches the MemoryStore and if the key is found
// deletes the Message pointer associated with it.
func (store *MemoryStore) Del(key string) {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		lg.Error("Trying to use memory store, but not open")
		return
	}
	//mid := mIDFromKey(key)
	m := store.messages[key]
	if m == nil {
		lg.Warning("memorystore del: message not found")
	} else {
		delete(store.messages, key)
		lg.Debug("memorystore del: message was deleted")
	}
}

// Close will disallow modifications to the state of the store.
func (store *MemoryStore) Close() {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		lg.Error("Trying to close memory store, but not open")
		return
	}
	store.opened = false
	lg.Debug("memorystore closed")
}

// Reset eliminates all persisted message data in the store.
func (store *MemoryStore) Reset() {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		lg.Error("Trying to reset memory store, but not open")
	}
	store.messages = make(map[string]packets.ControlPacket)
	lg.Debug("memorystore wiped")
}
