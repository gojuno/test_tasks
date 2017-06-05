package junoKvServer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"bytes"
	"github.com/gobwas/glob"
)

type commandHandler func(c *client) error

var cmdList = map[string]commandHandler{}

func addCmd(name string, f commandHandler) {
	if _, ok := cmdList[strings.ToLower(name)]; ok {
		panic(fmt.Sprintf("%s already registered", name))
	}

	cmdList[name] = f
}

func prepareKey(key string, checkedType int, server *KVServer, createIfNotExists bool, createValue interface{}) (k *kvEntry, err error) {
	d, ok := server.data.Get(key)
	if ok && d.t != checkedType {
		return nil, errors.New("Wrong key type")
	} else if !ok && !createIfNotExists {
		return nil, errors.New("Key not exists")
	} else if !ok && createIfNotExists {
		d, err = NewKVEntry(checkedType, createValue)
		if err != nil {
			return nil, err
		}
		return d, nil
	}
	return d, nil
}

func getCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return errCmdParams
	}

	if v, ok := c.server.data.Get(string(args[0])); !ok {
		c.respWriter.writeBulk(nil)
	} else if v.t != itemTypeString {
		return errors.New("Invalid key type")
	} else {
		if v == nil {
			c.respWriter.writeBulk(nil)
		} else {
			c.respWriter.writeBulk(v.Data().([]byte))
		}
	}
	return nil
}

func setCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return errCmdParams
	}

	if err := c.server.data.Set(string(args[0]), itemTypeString, args[1]); err != nil {
		return err
	}
	c.respWriter.writeStatus(okStatus)

	return nil
}

/** @todo #1:60m/DEV need strong refactoring for Expire, what happening when we create too much timer? */
func expireCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return errCmdParams
	}
	key := string(args[0])
	duration, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return errValue
	}

	if v, ok := c.server.data.Get(key); !ok {
		c.respWriter.writeInteger(0)
	} else {
		if v.timer != nil {
			v.timer.Stop()
		}
		v.timer = time.NewTimer(time.Second * time.Duration(duration))
		go func() {
			<-v.timer.C
			c.server.data.Del(key)
		}()
		c.respWriter.writeInteger(1)
	}

	return nil
}

func delCommand(c *client) error {
	args := c.args
	if len(args) == 0 {
		return errCmdParams
	}
	var deleted int64
	for _, key := range args {
		key := string(key)
		if v, ok := c.server.data.Get(key); ok {
			if v.timer != nil {
				v.timer.Stop()
			}
			c.server.data.Del(key)
			deleted++
		}
	}
	c.respWriter.writeInteger(deleted)

	return nil
}

/** @todo #1:60m/DEV memory management need be more efficient */
func lpushCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return errCmdParams
	}
	key := string(args[0])
	entry, err := prepareKey(key, itemTypeList, c.server, true, [][]byte{})
	if err != nil {
		return err
	}
	entry.l = append(entry.l, args[1:]...)
	/** @todo #1:15m/DEV,ARCH Are we need copy over cmap.Set or append is enough? */
	c.server.data.Set(key, itemTypeList, entry.l)

	c.respWriter.writeInteger(int64(len(args) - 1))

	return nil
}

func lrangeCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return errCmdParams
	}

	var start int64
	var stop int64
	var err error

	start, err = strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return errValue
	}

	stop, err = strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return errValue
	}

	entry, err := prepareKey(string(args[0]), itemTypeList, c.server, false, nil)
	if err != nil {
		return err
	}
	llen := int64(len(entry.l))

	if start < 0 {
		start = llen + start
	}
	if stop < 0 {
		stop = llen + stop
	}
	if start < 0 {
		start = 0
	}

	if start > stop || start >= llen {
		c.respWriter.writeSliceArray([][]byte{})
		return nil
	}

	if stop >= llen {
		stop = llen - 1
	}
	limit := (stop - start) + 1
	c.respWriter.writeSliceArray(entry.l[start:start+limit])
	return nil

}

func lsetCommand(c *client) error {
	var lst *kvEntry
	var err error
	var index int64

	args := c.args
	if len(args) != 2 {
		return errCmdParams
	}
	key := string(args[0])
	lst, err = prepareKey(key, itemTypeList, c.server, false, nil)
	if err != nil {
		return err
	}

	index, err = strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return errValue
	}
	llen := int64(len(lst.l))
	if index < 0 {
		index = llen + index
	}

	if index > llen {
		return errors.New("List index out of range")
	}
	lst.l[index] = args[2]
	/** @todo #1:15m/DEV,ARCH Are we need copy over cmap.Set or append is enough? */
	c.server.data.Set(key, itemTypeList, lst.l)
	return nil
}

/** @todo #1:60m/DEV,ARCH think about refactoring in LREM without full copy of list, for low memory footprint */
func lremCommand(c *client) error {
	var entry *kvEntry
	var err error
	var count int64

	args := c.args
	if len(args) != 3 {
		return errCmdParams
	}
	key := string(args[0])
	entry, err = prepareKey(key, itemTypeList, c.server, false, nil)
	if err != nil {
		return err
	}

	count, err = strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return errValue
	}
	value := args[2]
	llen := int64(len(entry.l))

	var step int64 = 1
	var begin int64
	var deleted int64

	if count < 0 {
		begin = llen
		step = -1
		count = -count
	}

	for i := begin; i >= 0 && i < llen; i += step {
		if bytes.Equal(entry.l[i], value) {
			entry.l[i] = nil
			deleted++
			if deleted == count {
				break
			}
		}
	}

	compressedList := make([][]byte, llen-deleted)
	i := 0
	for _, v := range entry.l {
		if v != nil {
			compressedList[i] = v
			i++
		}
	}

	c.server.data.Set(key, itemTypeList, compressedList)
	c.respWriter.writeInteger(deleted)
	return nil
}

func hsetCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return errCmdParams
	}
	key := string(args[0])
	subKey := string(args[1])
	value := args[2]
	entry, err := prepareKey(key, itemTypeMap, c.server, true, map[string][]byte{})
	if err != nil {
		return err
	}
	/** @todo #1:15m/DEV need check for value size */
	entry.m[subKey] = value
	c.server.data.Set(key, itemTypeMap, entry.m)
	c.respWriter.writeInteger(1)
	return nil
}

func hgetCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return errCmdParams
	}

	key := string(args[0])
	subKey := string(args[1])
	entry, err := prepareKey(key, itemTypeMap, c.server, false, nil)
	if err != nil {
		return err
	}
	if v, exists := entry.m[subKey]; !exists {
		c.respWriter.writeBulk(nil)
	} else {
		c.respWriter.writeBulk(v)
	}

	return nil
}

func hdelCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return errCmdParams
	}

	key := string(args[0])
	entry, err := prepareKey(key, itemTypeMap, c.server, false, nil)
	if err != nil {
		return err
	}

	var deleted int64
	for _, subKey := range args[1:] {
		if _, exists := entry.m[string(subKey)]; exists {
			delete(entry.m, string(subKey))
			deleted++
		}
	}

	c.respWriter.writeInteger(deleted)

	return nil
}

func keysCommand(c *client) error {
	args := c.args
	if len(args) > 1 {
		return errCmdParams
	}
	pattern := string(args[0])
	g, err := glob.Compile(pattern)
	if err != nil {
		return err
	}

	keys := make([]interface{}, 0)

	c.server.data.RangeOverCallback(func(key string, v interface{}) {
		if g.Match(key) {
			keys = append(keys, key)
		}
	})

	if len(keys) == 0 {
		c.respWriter.writeBulk(nil)
	}

	c.respWriter.writeArray(keys)
	return nil
}

func init() {
	addCmd("expire", expireCommand)
	addCmd("get", getCommand)
	addCmd("set", setCommand)
	addCmd("del", delCommand)
	addCmd("lpush", lpushCommand)
	addCmd("lrange", lrangeCommand)
	addCmd("lset", lsetCommand)
	addCmd("lrem", lremCommand)
	addCmd("hget", hgetCommand)
	addCmd("hset", hsetCommand)
	addCmd("hdel", hdelCommand)
	addCmd("keys", keysCommand)
}
