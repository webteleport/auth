package auth

import (
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/btwiuse/rng"
	"golang.org/x/exp/slices"
)

//go:embed login.html
var LoginIndex string

func getAllFilenames(efs fs.FS) (files []string, err error) {
	if err := fs.WalkDir(efs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		files = append(files, path)

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

type LoginMiddleware struct {
	// set login password
	// empty value removes password check
	Password string
	// optional, fallback to "UFOSID" if empty
	SessionKey string
	// optional, fallback to "password" if empty
	PasswordKey string

	sessions map[string]struct{}
	mutex    *sync.RWMutex
	login    http.Handler
	files    []string
}

func (lm *LoginMiddleware) AddSessionId(id string) {
	lm.mutex.Lock()
	lm.sessions[id] = struct{}{}
	lm.mutex.Unlock()
}

func (lm *LoginMiddleware) HasSessionId(id string) bool {
	lm.mutex.RLock()
	_, ok := lm.sessions[id]
	lm.mutex.RUnlock()
	return ok
}

func (lm *LoginMiddleware) RedirectToLogin(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasSuffix(r.URL.Path, "/login/"):
		break
	case slices.Contains(lm.files, path.Base(r.URL.Path)):
		break
	default:
		// redirect to login page
		// after login, redirect back to original page preserving query strings
		originalUrl := r.URL.String()
		loginAndNext := fmt.Sprintf("./login/?next=%s", originalUrl)
		http.Redirect(w, r, loginAndNext, 302)
		return
	}
	lm.login.ServeHTTP(w, r)
}

func (lm *LoginMiddleware) SetCookiesAndRedirect(w http.ResponseWriter, r *http.Request) {
	sid := rng.NewUUID()
	lm.AddSessionId(sid)
	cookies := fmt.Sprintf(`%s="%s"; Max-Age=2592000; HttpOnly; SameSite=Lax`, lm.SessionKey, sid)
	w.Header().Set("Set-Cookie", cookies)

	// redirect to next page
	next := r.URL.Query().Get("next")
	if next != "" {
		http.Redirect(w, r, next, 302)
		return
	}

	http.Redirect(w, r, "/", 302)
}

func (lm *LoginMiddleware) IsValidLogin(r *http.Request) bool {
	switch r.Method {
	case http.MethodGet:
		return strings.HasSuffix(r.URL.Path, fmt.Sprintf("/login/%s", lm.Password))
	case http.MethodPost:
		return r.PostFormValue(lm.PasswordKey) == lm.Password
	}
	return false
}

func (lm *LoginMiddleware) IsValidSession(r *http.Request) bool {
	sid, err := r.Cookie(lm.SessionKey)
	if err != nil || !lm.HasSessionId(sid.Value) {
		return false
	}
	return true
}

func LoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, LoginIndex)
	})
}

func (lm *LoginMiddleware) initialize() {
	if lm.mutex == nil {
		lm.mutex = &sync.RWMutex{}
	}
	if lm.sessions == nil {
		lm.sessions = map[string]struct{}{}
	}
	if lm.login == nil {
		lm.login = LoginHandler()
	}
	if lm.PasswordKey == "" {
		lm.PasswordKey = "password"
	}
	if lm.SessionKey == "" {
		lm.SessionKey = fmt.Sprintf("SESSION%s", rng.NewDigits(8))
	}
}

func (lm *LoginMiddleware) IsLocalhost(r *http.Request) bool {
	hostonly, _, _ := strings.Cut(r.URL.Host, ":")
	return strings.HasSuffix(hostonly, "localhost")
}

// PrecheckAccessToken returns a bool that indicates whether the caller should continue
func (lm *LoginMiddleware) Wrap(next http.Handler) http.Handler {
	lm.initialize()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// skip login if no password
		case lm.Password == "":
			break
		// skip login if on localhost
		case lm.IsLocalhost(r):
			break
		// validate session id for all requests
		case !lm.IsValidSession(r):
			// validate password for login requests
			if lm.IsValidLogin(r) {
				lm.SetCookiesAndRedirect(w, r)
				return
			}
			// lm.RedirectToLogin(w, r)
			lm.login.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// WithPassword is a middleware that checks for a password
func WithPassword(handler http.Handler, secret string) http.Handler {
	if strings.Contains(secret, ":") {
		return BasicAuth(handler, secret)
	}
	lm := &LoginMiddleware{
		Password: secret,
	}
	return lm.Wrap(handler)
}
