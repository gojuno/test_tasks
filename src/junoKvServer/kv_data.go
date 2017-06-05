package junoKvServer

import (
	"time"
	"errors"
	"github.com/streamrail/concurrent-map"
)

const (
	itemTypeString = iota
	itemTypeList   = iota
	itemTypeMap    = iota
)

/** @todo #1:60m/DEV,ARCH try use in future https://github.com/golang/sync/blob/master/syncmap/map.go instead of Mutex + Map ? */
type kvData struct {
	data cmap.ConcurrentMap
}

/** @todo #1:60m/DEV,ARCH it's really simple stupid global data map implementation (PHP like ;) with highly memory fragmentation, are we need more smarter implementation? */
type kvEntry struct {
	t     int
	s     []byte
	m     map[string][]byte
	l     [][]byte
	timer *time.Timer
}

func newKVData() *kvData {
	kv := new(kvData)
	kv.data = cmap.New()
	return kv
}

// NewKVEntry allocate memory for new Key-Value complex type entry
func NewKVEntry(t int, value interface{}) (*kvEntry, error) {
	d := new(kvEntry)
	d.t = t
	switch t {
	case itemTypeString:
		d.s = value.([]byte)
	case itemTypeMap:
		d.m = value.(map[string][]byte)
	case itemTypeList:
		d.l = value.([][]byte)
	default:
		return nil, errors.New("Invalid value type")
	}
	return d, nil
}

// Data return part of entry depends on current entry type
func (d *kvEntry) Data() (v interface{}) {
	switch d.t {
	case itemTypeString:
		return d.s
	case itemTypeMap:
		return d.m
	case itemTypeList:
		return d.l
	default:
		return nil
	}
}

// Set set value for key in concurrent map
func (kv *kvData) Set(key string, t int, value interface{}) error {
	d, err := NewKVEntry(t, value)
	if err != nil {
		return err
	}
	kv.data.Set(key, d)
	return nil
}

// Set get value by key from concurrent map
func (kv *kvData) Get(key string) (d *kvEntry, ok bool) {
	v, ok := kv.data.Get(key)
	if ok {
		d = v.(*kvEntry)
	} else {
		d = nil
	}
	return d, ok
}

// Set delete key from concurrent map
func (kv *kvData) Del(key string) {
	kv.data.Remove(key)
}

func (kv *kvData) RangeOverCallback(fn cmap.IterCb) {
	kv.data.IterCb(fn)
}
