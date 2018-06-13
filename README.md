Mothership
=========

A zero configuration web service that holds recent status updates from my servers.

It has a single web page that lists servers, gives me a link to their SSH port and indicates if a server isn't running anymore.

Installation:

```
go get github.com/nictuku/mothership
mkdir ~/.mothership
cp "$(go env GOPATH)/src/github.com/nictuku/mothership/mothership.json" ~/.mothership
```

Configuration:

Edit ~/.mothership/mothership.json to add your personal pushover key.


Agent installation:

```
curl -sL https://raw.github.com/nictuku/mothership/master/agent/install.sh | bash -x
```
