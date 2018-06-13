// package login provides a simplified login mechanism that uses GitHub authentication.
// See the documentation for CurrentPassword for how to use.
package login

// TODO: clean this up.

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

// OAuth2CallBack is the URL path used by GitHub to redirect users after an
// authentication attempt. It must match the path used in the GitHub API client
// registration.
const OAuth2CallBack = "/oauth2callback"

func init() {
	http.HandleFunc("/ghlogin", handleAuthorize)
	http.HandleFunc(OAuth2CallBack, handleAuthToken)
}

// TODO: key rotation based on time, for stability after server restarts.
var cookies = sessions.NewCookieStore(securecookie.GenerateRandomKey(64))

type Passport struct {
	Authorized bool
	Login      string `login`
}

// CurrentPassword inspects cookies and finds if the user has been authenticated already. If so, the
// user details are returned in Password - otherwise an error is provided. After receiving an error,
// callers should confirm that the referrer is not their own application, then redirect the user to
// /ghlogin which will show a login form.
//
// The page generating the /ghlogin redirect can store a cookie called "ref" with the current URL
// path as the value. After login is authenticated, we'll inspect the ref cookie and direct the user
// back there. Make sure the path stored in the cookie makes it readable from the /ghlogin handler.
func CurrentPassport(r *http.Request) (*Passport, error) {
	session, _ := cookies.Get(r, "userauth")
	login, ok := session.Values["login"].(string)
	if !ok {
		return nil, fmt.Errorf("passport cookie not found")
	}
	return &Passport{Authorized: true, Login: login}, nil
}

// Service should be overridden by the client library. It can't be used
// concurrently therefore it should be set during program initialization, or
// sometime before it's used.
var Service *oauth2.Config

func handleAuthorize(w http.ResponseWriter, r *http.Request) {
	if Service == nil {
		http.Error(w, "Passport service not configured", http.StatusInternalServerError)
		return
	}
	url := Service.AuthCodeURL("")
	http.Redirect(w, r, url, http.StatusFound)
}

const (
	userApiEndPoint = "https://api.github.com/user"
)

func handleAuthToken(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if Service == nil {
		http.Error(w, "Passport service not configured", http.StatusInternalServerError)
		return
	}
	token, err := Service.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Printf("service token err %v, code %v", err, code)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	github := Service.Client(oauth2.NoContext, token)
	githubUserData, err := github.Get(userApiEndPoint)
	if err != nil {
		log.Println("github get", err)
		http.Error(w, "github userApiEndPoint get failure", http.StatusInternalServerError)
		return
	}
	defer githubUserData.Body.Close()

	passport, err := parsePassport(githubUserData.Body)
	if err != nil {
		log.Println("parsePassport err:", err)
	} else {
		session, _ := cookies.Get(r, "userauth")
		session.Values["login"] = passport.Login
		session.Save(r, w)
	}

	redirectTo := "/"
	cookie, err := r.Cookie("ref")
	if err != nil {
		log.Printf("could not fetch ref cookie: %v", err)
	} else {
		redirectTo = cookie.Value
	}
	log.Printf("Redirecting back to %v", redirectTo)
	http.Redirect(w, r, redirectTo, http.StatusFound)

	return
}

func parsePassport(body io.ReadCloser) (*Passport, error) {
	b, _ := ioutil.ReadAll(body)
	var passport Passport
	if err := json.Unmarshal(b, &passport); err != nil {
		log.Println("passport json decoding:", err)
		return nil, err
	}
	if len(passport.Login) > 0 {
		passport.Authorized = true
	}
	return &passport, nil
}
