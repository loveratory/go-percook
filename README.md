percook: *per*sistent *cook*iejar
===
percook makes any [`http.Cookiejar`](https://pkg.go.dev/net/http?tab=doc#CookieJar) implementations persistable.

This library requires Go 1.9 because using `sync.Map`.

Usage
---

Here is example with [https://httpbin.org/](https://httpbin.org/)'s Cookies endpoints.

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"

	"github.com/otofune/go-percook"
)

func panicGet(c *http.Client, url string) {
	if _, err := c.Get(url); err != nil {
		panic(err)
	}
}

func main() {
	httpJar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		panic(err)
	}
	pjar := percook.New(httpJar)
	client := http.Client{Jar: pjar}

	panicGet(&client, "https://httpbin.org/cookies/set/abc/123")
	panicGet(&client, "https://httpbin.org/cookies/set/edf/123")
	panicGet(&client, "https://httpbin.org/cookies/set/will_be_removed/100")
	panicGet(&client, "https://httpbin.org/cookies/delete?will_be_removed=")

	for u, cookies := range pjar.AllCookies() {
		for _, cookie := range cookies {
			fmt.Printf("%s: %s\n", u.String(), cookie.String())
		}
	}
}
```

Output expected:

```sh
$ # output
$ go run ./examples/readme
http://httpbin.org/: abc=123
http://httpbin.org/: edf=123
$
```

LICENSE
---
This project licensed under [MIT](./LICENSE.txt) unless otherwise specified.
