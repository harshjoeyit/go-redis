package core

import (
	"time"
)

type Obj struct {
	value     any
	expiresAt int64
}

func NewObj(value any, expDurationMs int64) *Obj {
	var expiresAt int64 = -1 // default expiry = infinity
	if expDurationMs > 0 {
		expiresAt = time.Now().UnixMilli() + expDurationMs
	}
	return &Obj{
		value:     value,
		expiresAt: expiresAt,
	}
}

type DataStore interface {
	Get(key string) (any, bool)
	Set(key string, value any, expDurationMs int64)
	TTL(key string) int64
}

type InMemDataStore map[string]*Obj

func NewInMemDataStore() InMemDataStore {
	return make(map[string]*Obj)
}

func (s InMemDataStore) Get(key string) (any, bool) {
	obj, ok := s[key]
	if !ok {
		return nil, false
	}
	if obj.expiresAt < time.Now().UnixMilli() {
		return nil, true
	}
	return obj.value, true
}

func (s InMemDataStore) Set(key string, value any, expDurationMs int64) {
	s[key] = NewObj(value, expDurationMs)
}

func (s InMemDataStore) TTL(key string) int64 {
	obj, ok := s[key]
	if !ok {
		return -1 // key not found
	}
	ttlMs := obj.expiresAt - time.Now().UnixMilli()
	if ttlMs < 0 {
		return -2 // expired
	}
	return ttlMs / 1000
}

func (s InMemDataStore) Snapshot() DataStore {
	ss := make(InMemDataStore, len(s))
	for k, v := range s {
		ss[k] = &Obj{
			value:     v.value,
			expiresAt: v.expiresAt,
		}
	}
	return ss
}
