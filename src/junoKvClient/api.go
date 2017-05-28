package junoKvClient

import (
	"errors"
	"github.com/serialx/hashring"
	"github.com/siddontang/goredis"
	"strconv"
)

// KVClient wrapper around goredis.Client with additional weight based consistent hashing ring
type KVClient struct {
	ring     *hashring.HashRing
	servers  map[string]int
	connects map[string]*goredis.Client
}

// NewClient use for initialize new client
func NewClient(servers map[string]int) *KVClient {
	c := new(KVClient)
	c.servers = servers
	c.ring = hashring.NewWithWeights(servers)
	c.connects = make(map[string]*goredis.Client)
	for addr := range c.servers {
		c.connects[addr] = goredis.NewClient(addr, "")
		c.connects[addr].SetMaxIdleConns(1)
	}
	return c
}

// Close closing of all connected clients
func (c *KVClient) Close() {
	for addr := range c.connects {
		c.connects[addr].Close()
	}
}

// RunCmd select connect from hash ring and execute command inside goredis
// @todo implement really async client
func (c *KVClient) RunCmd(cmd string, args ...interface{}) (interface{}, error) {
	key := args[0].(string)
	addr, ok := c.ring.GetNode(key)
	if ok {
		v, err := c.connects[addr].Do(cmd, args...)
		return v, err
	}
	return nil, errors.New("Can't find server from hashring by key")
}

// Expire set expire for selected key, key will be deleted on server after ttl seconds
func (c *KVClient) Expire(key string, ttl int64) (err error) {
	_, err = c.RunCmd("EXPIRE", key, ttl)
	return err
}

// Get get value for STRING_TYPE entry, return value, wrong type error or nil
func (c *KVClient) Get(key string) (string, error) {
	v, err := c.RunCmd("GET", key)
	if err == nil {
		return string(v.([]byte)), err
	} else {
		return "", err
	}

}

// Set set value for STRING_TYPE entry with ttl
func (c *KVClient) Set(key string, value interface{}, ttl int64) (error) {
	_, err := c.RunCmd("SET", key, value)
	if err == nil && ttl > 0 {
		return c.Expire(key, ttl)
	}
	return err
}

// Del delete one entry, always return nil, b
// server can delete multiple entry
// @todo anybody will read this code ?
func (c *KVClient) Del(key string) (err error) {
	_, err = c.RunCmd("DEL", key)
	return err
}

// ListPush create LIST_TYPE entry if not exists and append multiple values
// @todo how to skip conversion from ...interface to interface  ?
func (c *KVClient) ListPush(key string, values ...interface{}) (pushed int64, err error) {
	args := make([]interface{}, len(values) + 1)
	args[0] = key
	for i,v := range values {
		switch v:=v.(type) {
		case int:
			args[i+1] = strconv.Itoa(int(v))
		case string:
			args[i+1] = v
		case []byte:
			args[i+1] = string(v)
		}
	}
	p, err := c.RunCmd("LPUSH", args...)
	if err != nil {
		return 0, err
	}

	switch p := p.(type) {
	case string:
		return strconv.ParseInt(p, 10, 64)
	case []byte:
		return strconv.ParseInt(string(p), 10, 64)
	case int:
		return int64(p), err
	case int64:
		return p, err
	default:
		return 0, errors.New("Invalid LPUSH response")
	}
}


// ListRange return slice from LIST_TYPE key, return error if key not exists see https://redis.io/commands/lrange for other restrictions
// @todo maybe replace int64 to int16 or int32 ? we need strong defence from out of memory
func (c *KVClient) ListRange(key string, from int64, to int64) ([]string, error) {
	l, err := c.RunCmd("LRANGE", key, from, to)
	if err != nil {
		return nil, err
	}
	// @todo need best implementation for convert []interface{} to []string somebody please explain me WHAT this code do?
	switch l := l.(type) {
	case []interface{}:
		s := make([]string,len(l))
		for i, v := range l {
			switch v := v.(type) {
			case []byte:
				s[i] =  string(v)
			}
		}
		return s, err
	}
	return nil, errors.New("Invalid LRANGE response type")
}


// ListSet set value for item in LIST_TYPE key with index offset, return error if key not found or key has wrong type
// @todo maybe replace int64 to int16 or int32 ? we need strong defence from out of memory
func (c *KVClient) ListSet(key string, index int64, value interface{}) (err error) {
	_, err = c.RunCmd("LSET", key, index, value)
	return err
}

// ListDel delete item from LIST_TYPE key by value https://redis.io/commands/lrem for other restrictions
// @todo maybe replace int64 to int16 or int32 ? we need strong defence from out of memory
func (c *KVClient) ListDel(key string, count int64, value interface{}) (err error) {
	_, err = c.RunCmd("LREM", key, count, value)
	return err
}

// DictGet get item from MAP_TYPE by subKey return nil if subKey not exists
func (c *KVClient) DictGet(key string, subKey string) (interface{}, error) {
	return c.RunCmd("HGET", key, subKey)
}

// DictSet set item value for MAP_TYPE by subKey create entry with key if not exists
func (c *KVClient) DictSet(key string, subKey string, value interface{}) (err error) {
	_, err = c.RunCmd("HSET", key, subKey, value)
	return err
}

// DictDel delete item value from MAP_TYPE by subKey , return error if key have wrong type
func (c *KVClient) DictDel(key string, subKey string) (err error) {
	_, err = c.RunCmd("HDEL", key, subKey)
	return err
}
