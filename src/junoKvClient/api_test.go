package junoKvClient

import (
	"github.com/alicebob/miniredis"
	"testing"
)

// @todo implements full key-value api test with miniredis
func TestJunoKVClientAPI(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	addr := s.Addr()

	c := NewClient(map[string]int{addr:1})
	defer c.Close()

	if err = c.Set("test", 1, 0); err != nil {
		t.Fatal(err)
	}

	if v, err := c.Get("test"); err != nil {
		t.Fatal(err)
	} else if v != "1" {
		t.Fatalf("GET test return %v", v)
	}

	if pushed, err := c.ListPush("test_list", 1, 2, 3); err != nil {
		t.Fatal(err)
	} else if pushed != 3 {
		t.Fatalf("ListPush return %v instead of 1 rows", pushed)
	}

	if v, err := c.ListRange("test_list", 0, -1); err != nil {
		t.Fatal(err)
	} else {
		if len(v) != 3 {
			t.Fatalf("LRANGE test_list 0 -1 return %s %T", v, v)
		}
	}


}

