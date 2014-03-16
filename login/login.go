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

	"bitbucket.org/gosimple/oauth2"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
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
	Email      string `email`
}

// CurrentPassword inspects cookies and finds if the user has been authenticated already. If so, the
// user details are returned in Password - otherwise an error is provided. After receiving an error,
// callers should confirm that the referrer is not their own application, then redirect the user to
// /ghlogin which will show a login form.
func CurrentPassport(r *http.Request) (*Passport, error) {
	session, _ := cookies.Get(r, "userauth")
	email, ok := session.Values["email"].(string)
	if !ok {
		return nil, fmt.Errorf("passport cookie not found")
	}
	return &Passport{Authorized: true, Email: email}, nil
}

func handleAuthorize(w http.ResponseWriter, r *http.Request) {
	url := ghService.GetAuthorizeURL("")
	http.Redirect(w, r, url, http.StatusFound)
}

const (
	apiBaseURL  = "https://api.github.com/"
	apiEndPoint = "user"
)

func handleAuthToken(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	token, err := ghService.GetAccessToken(code)
	if err != nil {
		log.Printf("service token err %v, code %v", err, code)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	github := oauth2.Request(apiBaseURL, token.AccessToken)
	github.AccessTokenInHeader = true
	github.AccessTokenInHeaderScheme = "token"
	github.AccessTokenInURL = true

	githubUserData, err := github.Get(apiEndPoint)
	if err != nil {
		log.Println("github get", err)
		http.Error(w, "github apiEndPoint get failure", http.StatusInternalServerError)
		return
	}
	defer githubUserData.Body.Close()

	passport, err := parsePassport(githubUserData.Body)
	if err != nil {
		log.Println("parsePassport err:", err)
	} else {
		session, _ := cookies.Get(r, "userauth")
		session.Values["email"] = passport.Email
		session.Save(r, w)
	}

	// TODO: Store the target URL for the redirect in a cookie, set by /ghlogin.
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

func parsePassport(body io.ReadCloser) (*Passport, error) {
	b, _ := ioutil.ReadAll(body)
	var passport Passport
	if err := json.Unmarshal(b, &passport); err != nil {
		log.Println("passport json decoding:", err)
		return nil, err
	}
	if len(passport.Email) > 0 {
		passport.Authorized = true
	}
	return &passport, nil
}