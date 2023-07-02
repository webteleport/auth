package main

import (
	"net/http"

	"github.com/webteleport/auth"
	"github.com/webteleport/webteleport/ufo"
)

func main() {
	pwd := http.FileServer(http.Dir("."))
	lm := auth.LoginMiddleware{
		Password: "123",
	}
	handler := lm.Wrap(pwd)
	ufo.Serve("https://ufo.k0s.io", handler)
}
