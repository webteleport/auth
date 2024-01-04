// 1. visit /dir/?a
// 2. be redirected to /login/?next=/dir/?a
// 3. enter password 123
// 4. be redirected to /dir/?a

package main

import (
	"net/http"

	"github.com/webteleport/auth"
	"github.com/webteleport/wtf"
)

func main() {
	wtf.Serve("https://ufo.k0s.io", auth.WithPassword(http.FileServer(http.Dir(".")), "123"))
}
