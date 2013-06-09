package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/nictuku/myservers/login"
)

var (
	port = flag.Int("port", 80, "Port on which to run the web server.")
)

const (
	username = "nictuku" // XXX
)

const indexTemplate = `
<html>
<head><title>Mothership</title>
</head>
<body>
	<h1>Logged in as Yves Junqueira</h1>
	<h2>List of servers</h2>
	<table>
	<th>hostname</th><th>IP</th><th>Last seen</th>
	{{range .}}
	<tr><td><a href="ssh://{{ if .Username }}{{.Username}}@{{ end }}{{.IP}}:{{.SSHPort}}">{{.Hostname}}</a></td><td>{{.IP}}</td><td>{{with .Since}}{{.}} ago. {{end}}</td></tr>
	{{end}}
	</table>
</body>
</html>

`

var index *template.Template

type ServerInfo struct {
	Hostname    string
	Username    string
	SSHPort     int
	IP          string
	LastContact time.Time
}

func (s ServerInfo) Since() time.Duration {
	return time.Since(s.LastContact)
}

type serversInfo struct {
	sync.RWMutex
	Info map[string]ServerInfo
}

var servers *serversInfo

func indexHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		log.Printf("unknown path %v", req.URL.Path)
		http.NotFound(w, req)
		return
	}
	servers.Lock()
	defer servers.Unlock()
	hostname := req.FormValue("hostname")

	// TODO: Make a separate handler for the heartbeats.
	if hostname != "" {
		s := ServerInfo{
			Hostname:    hostname,
			Username:    username,
			LastContact: time.Now(),
		}
		sshPort, err := strconv.Atoi(req.FormValue("sshPort"))
		if sshPort == 0 || err != nil {
			sshPort = -1
		}
		s.SSHPort = sshPort

		if ip, _, err := net.SplitHostPort(req.RemoteAddr); err == nil && ip != "" {
			s.IP = ip
		}
		// TODO: ensure that a rogue HTTP client can't override legitimate entries.
		servers.Info[s.Hostname] = s
		log.Println("updated server", s.Hostname)
		io.WriteString(w, "ok")
		return
	}

	passport, err := login.CurrentPassport(req)
	if err != nil {
		log.Println("Redirecting to ghlogin, %q. Referrer was:", err, req.Referer())
		http.Redirect(w, req, "/ghlogin", http.StatusFound)
		return
	}
	if passport.Email != "yves.junqueira@gmail.com" {
		http.Error(w, "Nope.", http.StatusForbidden)
		return
	}

	err = index.Execute(w, servers.Info)
	if err != nil {
		log.Println(err)
	}
}

func wwwDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = string(filepath.Separator)
	}
	dir := filepath.Join(home, "www")
	return dir
}

func main() {
	flag.Parse()
	index = template.Must(template.New("index").Parse(indexTemplate))
	servers = &serversInfo{Info: map[string]ServerInfo{}}
	http.HandleFunc("/", indexHandler)
	log.Println("Serving mothership index at /")

	dir := wwwDir()
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))
	log.Printf("Serving static files from %v at /static", dir)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Println("Error starting www server:", err)
		// os.IsPermission(err) doesn't work.
		if *port == 80 {
			log.Printf("Try: sudo setcap 'cap_net_bind_service=+ep' %v", os.Args[0])
		}
	}
}
