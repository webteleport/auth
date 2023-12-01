// 1. visit /dir/
// 2. be redirected to /login/?next=/dir/
// 3. enter password 123
// 4. be redirected to /dir/

package main

import (
	"net/http"

	"github.com/webteleport/auth"
	"github.com/webteleport/webteleport/ufo"
)

func main() {
	ufo.Serve("https://ufo.k0s.io", auth.WithPassword(http.FileServer(http.Dir(".")), "123"))
}
