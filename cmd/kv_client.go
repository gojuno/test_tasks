package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"github.com/peterh/liner"
	"github.com/siddontang/goredis"
	// @todo use juno_kv_client instead of goredis
	// "juno_kv_client"
	"log"
)

var ip = flag.String("h", "127.0.0.1", "server ip (default 127.0.0.1)")
var port = flag.Int("p", 8379, "server port (default 8379)")

var (
	line        *liner.State
	historyPath = "/tmp/.juno-kv-client-history"
)

func main() {
	flag.Parse()

	line = liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	loadHisotry()
	defer saveHisotry()

	var addr string
	addr = fmt.Sprintf("%s:%d", *ip, *port)

	c := goredis.NewClient(addr, "")
	c.SetMaxIdleConns(1)

	reg, err := regexp.Compile(`'.*?'|".*?"|\S+`)
	if err != nil {
		return
	}

	prompt := fmt.Sprintf("%s> ", addr)

	for {

		cmd, err := line.Prompt(prompt)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			return
		}

		cmds := reg.FindAllString(cmd, -1)
		if len(cmds) == 0 {
			continue
		} else {
			line.AppendHistory(cmd)

			args := make([]interface{}, len(cmds[1:]))

			for i := range args {
				args[i] = strings.Trim(string(cmds[1+i]), "\"'")
			}

			r, err := c.Do(cmds[0], args...)
			if err != nil {
				fmt.Printf("%s", err.Error())
			} else {
				printReply(0, r)
				fmt.Printf("\n")
			}

		}
	}
}

func printReply(level int, reply interface{}) {
	switch reply := reply.(type) {
	case int64:
		fmt.Printf("(integer) %d", reply)
	case string:
		fmt.Printf("(string) %s", reply)
	case []byte:
		fmt.Printf("(bytes) %q", reply)
	case nil:
		fmt.Print("(nil)")
	case goredis.Error:
		fmt.Printf("%s", string(reply))
	case []interface{}:
		fmt.Println("[]interface{}")
		for i, v := range reply {
			if i != 0 {
				fmt.Printf("%s", strings.Repeat(" ", level*4))
			}

			s := fmt.Sprintf("%d) ", i+1)
			fmt.Printf("%-4s", s)

			printReply(level+1, v)
			if i != len(reply)-1 {
				fmt.Printf("\n")
			}
		}
	default:
		fmt.Println("invalid server reply")
	}
}

func loadHisotry() {
	if f, err := os.Open(historyPath); err == nil {
		_, err = line.ReadHistory(f)
		if err != nil {
			log.Println("loadHistory error:", err)
		}
		if err = f.Close(); err != nil {
			log.Println("loadHistory error:", err)
		}
	}
}

func saveHisotry() {
	if f, err := os.Create(historyPath); err != nil {
		fmt.Println("Error writing history file: ", err)
	} else {
		_, err := line.WriteHistory(f)
		if err != nil {
			log.Println("loadHistory error:", err)
		}
		if err = f.Close(); err != nil {
			log.Println("loadHistory error:", err)
		}
	}
}
