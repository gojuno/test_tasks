package junoKvServer

import (
	"sync"
	"testing"
	"time"
	"strings"
	"junoKvClient"
)

var testServerOnce sync.Once
var testKVServer *KVServer
var testKVClient *junoKvClient.KVClient

var testKey string = time.Now().UTC().Format("2006-01-02__15-04-05.999999999")

func newTestClient() {
	testKVClient = junoKvClient.NewClient(map[string]int{"127.0.0.1:18379": 1})
}

func startTestServer() {
	f := func() {

		var err error
		testKVServer, err = NewKVServer("127.0.0.1:18379")
		if err != nil {
			println(err.Error())
			panic(err)
		}
		go testKVServer.Start()
		newTestClient()
	}

	testServerOnce.Do(f)
}

func TestServer(t *testing.T) {
	startTestServer()
}

func BenchmarkSETGETParallel(b *testing.B) {
	startTestServer()
	newTestClient()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			value := time.Now().UTC().Format("2006-01-02__15-04-05.999999999")
			if err := testKVClient.Set(testKey, value, 0); err != nil {
				b.Fatal(err)
			}
			v, err := testKVClient.Get(testKey)
			if err != nil {
				b.Fatal(err)
			}
			if v != "" {
				if strings.Index(v,"__") == -1 {
					b.Fatalf("__ not found in '%s'", v)
				}
			}
		}
	})
	testKVClient.Close()
}

func BenchmarkLPUSHLRANGEParallel(b *testing.B) {
	startTestServer()
	newTestClient()
	key := "test_list"
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if pushed, err := testKVClient.ListPush(key, "1"); err != nil {
				b.Fatal(err)
			} else if pushed != 1 {
				b.Fatalf("ListPush return %v instead of 1 rows", pushed)
			}

			v, err := testKVClient.ListRange(key, 0, -1)
			if err != nil {
				b.Fatal(err)
			}
			if len(v) == 0 {
				b.Fatal("Invalid length of test_list")
			}
		}
	})
	testKVClient.Close()
}
