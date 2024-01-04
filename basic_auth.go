package auth

import (
	"net/http"
	"strings"

	"github.com/kataras/basicauth"
)

// BasicAuth assumes the userpass strings contains :
func BasicAuth(next http.Handler, userpass string) http.Handler {
	splitResult := strings.Split(userpass, ":")
	user := splitResult[0]
	pass := splitResult[1]

	auth := basicauth.Default(map[string]string{
		user: pass,
	})

	return auth(next)
}
