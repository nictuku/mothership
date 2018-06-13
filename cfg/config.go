package cfg

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"strings"
)

func makeConfDir() string {
	dir := "/var/run/mothership"
	env := os.Environ()
	for _, e := range env {
		if strings.HasPrefix(e, "HOME=") {
			dir = strings.SplitN(e, "=", 2)[1]
			dir = path.Join(dir, ".mothership")
		}
	}
	// Ignore errors.
	os.MkdirAll(dir, 0750)

	if s, err := os.Stat(dir); err != nil {
		log.Fatal("stat config dir", err)
	} else if !s.IsDir() {
		log.Fatalf("Dir %v expected directory, got %v", dir, s)
	}
	return dir
}

func confPath() string {
	dir := makeConfDir()
	return path.Join(dir, "mothership.json")
}

type Config struct {
	PushoverKey string
	Users       []User
}

type User struct {
	Login               string
	PushoverDestination string
}

// ReadConfig reads the mothership configuration from $HOME/.mothership/mothership.json and returns
// the parsed Config.
func ReadConf() (cfg Config, err error) {
	file, err := os.Open(confPath())
	if err != nil {
		return cfg, err
	}
	decoder := json.NewDecoder(file)
	return cfg, decoder.Decode(&cfg)
}
