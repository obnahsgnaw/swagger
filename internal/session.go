package internal

import (
	"github.com/gorilla/sessions"
	"net/http"
)

var store = sessions.NewCookieStore([]byte("cptbtptp1122334455667788"))

func init() {
	store.MaxAge(60 * 60)
}

func GetSession(r *http.Request) *sessions.Session {
	session, _ := store.Get(r, "SWGSESSID")
	return session
}
