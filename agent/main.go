// agent is a program that frequently streams their status to a central
// server. 
//
// My goal is to have a web page that lists all my servers, gives me a
// link to their SSH port and indicates if a server isn't running anymore.
// This is the small agent that runs on each server and routinely contacts the
// mothership.
package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	heartBeatURL    = "http://1.a.magnets.im/" // XXX
	heartBeatPeriod = time.Second * 5          // XXX
	waitTime        = time.Second * 30
	debug           = true
)

func heartBeat() {
	s := newServerInfo()
	resp, err := http.PostForm(heartBeatURL, s.Values())
	if err != nil {
		log.Printf("Request to %v err %v", heartBeatURL, err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Printf("Response from %v: err %v", heartBeatURL, err)
		return
	}
	if resp.StatusCode != 200 {
		log.Printf("Response from %v: %v", heartBeatURL, resp.Status)
		if debug {
			log.Println(string(body))
		}
		return
	}
	log.Printf("Response from %v: %v", heartBeatURL, resp.Status)
}

type serverInfo struct {
	hostname string
	sshPort  int
}

func (s *serverInfo) Values() url.Values {
	param := make(url.Values)
	param.Set("hostname", s.hostname)
	param.Set("sshPort", strconv.Itoa(s.sshPort))
	return param
}

func newServerInfo() *serverInfo {
	h, _ := os.Hostname()
	if h == "" {
		h = "unknown"
	}
	return &serverInfo{
		hostname: h,
		sshPort:  22,
	}

}

func main() {
	tick := time.Tick(heartBeatPeriod)
	for {
		<-tick
		heartBeat()
	}
}
