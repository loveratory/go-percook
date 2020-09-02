package percook

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"golang.org/x/net/publicsuffix"
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
		cu, err := url.Parse(".")
		if err != nil {
			panic(fmt.Errorf("Unexpected, unrecovable parse error: %s", err))
		}
		path = stringCoalesceWithDefault("/", u.ResolveReference(cu).Path)
		if strings.HasSuffix(path, "/") {
			// remove trailing slash to reproduce same behaviour as above RFC
			path = path[:len(path)-1]
		}
	}
	return schema + domain + path
}

func (pjar *CookieJar) storeKey(u *url.URL, c *http.Cookie) {
	// avoid locking, load first
	pjar.seenKeys.LoadOrStore(toKey(u, c), struct{}{})
}

func (pjar *CookieJar) keys() ([]*url.URL, error) {
	var urls []*url.URL
	var err error
	pjar.seenKeys.Range(func(key, _ interface{}) bool {
		urlString, ok := key.(string)
		if !ok {
			err = fmt.Errorf("Unexpected, non sting key is stored in seenKeys. key=%#v", key)
			// = break
			return false
		}
		url, uerr := url.Parse(urlString)
		if uerr != nil {
			err = fmt.Errorf("Unexpected error, non url-format key is stored in seenKeys. key=%#v, err=%s", key, err)
			return false
		}
		urls = append(urls, url)
		// continue
		return true
	})

	if err != nil {
		return nil, err
	}

	return urls, nil
}

type CookiesMap map[*url.URL][]*http.Cookie

func (pjar *CookieJar) AllCookies() CookiesMap {
	keys, err := pjar.keys()
	if err != nil {
		panic(err)
	}

	type intermidiate = struct {
		URLString  *string
		URL        *url.URL
		TLDPlusOne *string
		Cookie     *http.Cookie
	}
	intsByCookieNameAndValue := make(map[string][]intermidiate)

	for _, u := range keys {
		u := u
		if u.Path == "" {
			// for consistent path
			u.Path = "/"
		}
		urlString := u.String()
		tldPlusOne, err := publicsuffix.EffectiveTLDPlusOne(u.Host)
		if err != nil {
			// fallback
			tldPlusOne = u.Host
		}

		// to check Domain= attribute, check with sloppy subdomain
		su := *u
		su.Host = "sloppy." + su.Host

		for _, u := range []*url.URL{u, &su} {
			u := u
			for _, cookie := range pjar.jar.Cookies(u) {
				cookie := cookie

				// domain, secure and path are saved in url,
				// only name & value needed
				nameAndValue := cookie.Name + cookie.Value
				intsByCookieNameAndValue[nameAndValue] = append(intsByCookieNameAndValue[nameAndValue], intermidiate{
					URL:        u,
					URLString:  &urlString,
					TLDPlusOne: &tldPlusOne,
					Cookie:     cookie,
				})
			}
		}
	}

	cmap := make(CookiesMap)

	// check uniqueness and assign to cmap
	// same cookie must appear at once
	for _, ints := range intsByCookieNameAndValue {
		groupIntsByTLDPlusOne := make(map[string][]*intermidiate)
		for _, inte := range ints {
			inte := inte
			groupIntsByTLDPlusOne[*inte.TLDPlusOne] = append(groupIntsByTLDPlusOne[*inte.TLDPlusOne], &inte)
		}

		for _, group := range groupIntsByTLDPlusOne {
			domains := make(map[string]struct{})
			shortest := group[0]
			shortestLen := len(*group[0].URLString)
			for _, item := range group {
				domains[item.URL.Host] = struct{}{}
				item := item
				if shortestLen > len(*item.URLString) {
					shortest = item
					shortestLen = len(*item.URLString)
				}
			}

			// fill cookie fields
			shortest.Cookie.Secure = shortest.URL.Scheme == "https"
			shortest.Cookie.Path = shortest.URL.Path
			if len(domains) > 1 {
				// appear multiple times = cookie is with Domain
				shortest.Cookie.Domain = shortest.URL.Host
			}

			cmap[shortest.URL] = append(cmap[shortest.URL], shortest.Cookie)
		}
	}

	return cmap
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
