package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestUnmarshalConf(t *testing.T) {
	tests := []struct {
		in  string
		out Config
	}{{`
{
  "PushoverKey": "pushkey",
  "Users": [
    {
      "Email": "yves.junqueira@gmail.com",
      "PushoverDestination": "pushoverdestin"
    }
  ]
}
`,
		Config{PushoverKey: "pushkey", Users: []User{{"yves.junqueira@gmail.com", "pushoverdestin"}}},
	}}
	for _, cfg := range tests {
		dec := json.NewDecoder(strings.NewReader(cfg.in))
		var c Config
		if err := dec.Decode(&c); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(c, cfg.out) {
			t.Fatalf("Config marsing wanted %v, got %v", cfg.out, c)

		}
		fmt.Printf("%s: %s\n", c.PushoverKey, c.Users)
	}
}
