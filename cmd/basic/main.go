// 1. visit /
// 2. enter user:pass
// 3. see directory content

package main

import (
	"net/http"

	"github.com/webteleport/auth"
	"github.com/webteleport/wtf"
)

func main() {
	wtf.Serve("https://ufo.k0s.io", auth.WithPassword(http.FileServer(http.Dir(".")), "user:pass"))
}
