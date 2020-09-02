package percook

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
)

type CookieJar struct {
	jar http.CookieJar
	// [cookieKey]struct{}
	seenKeys sync.Map
}

func New(jar http.CookieJar) *CookieJar {
	// return pointer because jar & seenKeys must be shared
	return &CookieJar{jar: jar}
}

func toKey(u *url.URL, c *http.Cookie) string {
	schema := "http://"
	if c.Secure {
		schema = "https://"
	}
	domain := stringCoalesceWithDefault(c.Domain, u.Host)
	if domain[0] == '.' {
		domain = domain[1:]
	}
	path := c.Path
	if len(path) == 0 {
		// > If the server omits the Path attribute, the user
		// > agent will use the "directory" of the request-uri's path component as
		// > the default value.
		// https://tools.ietf.org/html/rfc6265#section-4.1.2.4
		// see also: https://tools.ietf.org/html/rfc6265#section-5.1.4
		cu, err := url.Parse("./")
		if err != nil {
			panic(fmt.Errorf("Unexpected, unrecovable parse error: %s", err))
		}
		path = stringCoalesceWithDefault("/", u.ResolveReference(cu).Path)
	}
	return schema + domain + path
}

func (pjar *CookieJar) storeKey(u *url.URL, c *http.Cookie) {
	// avoid locking, load first
	pjar.seenKeys.LoadOrStore(toKey(u, c), struct{}{})
}

func (pjar *CookieJar) keys() []*url.URL {
	var urls []*url.URL
	pjar.seenKeys.Range(func(key, _ interface{}) bool {
		urlString, ok := key.(string)
		if !ok {
			panic(fmt.Errorf("Unexpected, non sting key is stored in seenKeys. key=%#v", key))
		}
		url, err := url.Parse(urlString)
		if err != nil {
			panic(fmt.Errorf("Unexpected error, non url-format key is stored in seenKeys. key=%#v, err=%s", key, err))
		}
		urls = append(urls, url)

		// please don't stop
		return true
	})
	return urls
}

type CookiesMap map[*url.URL][]*http.Cookie

func (pjar *CookieJar) AllCookies() CookiesMap {
	// same cookie must appear at once
	cookieByHostPlusCookieStr := make(map[string]*http.Cookie)
	keysByHostPlusCookieStr := make(map[string][]string)
	for _, key := range pjar.keys() {
		keyString := key.String()
		for _, cookie := range pjar.jar.Cookies(key) {
			cookie := cookie
			hpcs := key.Host + cookie.String()
			cookieByHostPlusCookieStr[hpcs] = cookie
			keysByHostPlusCookieStr[hpcs] = append(keysByHostPlusCookieStr[hpcs], keyString)
		}
	}
	reversedMap := make(CookiesMap)
	for hpcs, keys := range keysByHostPlusCookieStr {
		// shortest key = shortest scope
		u, _ := url.Parse(stringMin(keys...))
		copy := *cookieByHostPlusCookieStr[hpcs]
		copy.Secure = u.Scheme == "https"
		reversedMap[u] = append(reversedMap[u], &copy)
	}
	return reversedMap
}

func (pjar *CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	for _, c := range cookies {
		pjar.storeKey(u, c)
	}
	pjar.jar.SetCookies(u, cookies)
}
func (pjar *CookieJar) Cookies(u *url.URL) []*http.Cookie {
	return pjar.jar.Cookies(u)
}
