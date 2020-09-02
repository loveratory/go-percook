package percook

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"testing"
)

func panicParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func convertCookieMapToStringMap(a CookiesMap) map[string][]string {
	ret := make(map[string][]string)
	for k, vl := range a {
		kStr := k.String()
		for _, v := range vl {
			ret[kStr] = append(ret[kStr], v.String())
		}
	}
	return ret
}

func TestExportCookie(t *testing.T) {
	type SetCookie struct {
		Cookies []*http.Cookie
		URL     *url.URL
	}
	type TestCase struct {
		Name       string
		SetCookies []SetCookie
		Expected   CookiesMap
	}

	for i, tc := range []TestCase{
		{
			"without secure",
			[]SetCookie{
				{
					[]*http.Cookie{
						{
							Name:  "asdf",
							Value: "1234",
						},
					},
					panicParseURL("https://example.com/"),
				},
			},
			CookiesMap{
				panicParseURL("http://example.com/"): {
					{
						Name:  "asdf",
						Value: "1234",
						Path:  "/",
					},
				},
			},
		},
		{
			"different paths",
			[]SetCookie{
				{
					[]*http.Cookie{
						{
							Name:  "asdf",
							Value: "1234",
							Path:  "/pp",
						},
						{
							Name:  "zxcv",
							Value: "5678",
							Path:  "/",
						},
					},
					panicParseURL("https://example.com/"),
				},
			},
			CookiesMap{
				panicParseURL("http://example.com/pp"): {
					{
						Name:  "asdf",
						Value: "1234",
						Path:  "/pp",
					},
				},
				panicParseURL("http://example.com/"): {
					{
						Name:  "zxcv",
						Value: "5678",
						Path:  "/",
					},
				},
			},
		},
		{
			"one is secure",
			[]SetCookie{
				{
					[]*http.Cookie{
						{
							Name:   "asdf",
							Value:  "1234",
							Secure: true,
						},
						{
							Name:  "zxcv",
							Value: "5678",
						},
					},
					panicParseURL("https://example.com/"),
				},
			},
			CookiesMap{
				panicParseURL("https://example.com/"): {
					{
						Name:   "asdf",
						Value:  "1234",
						Secure: true,
						Path:   "/",
					},
				},
				panicParseURL("http://example.com/"): {
					{
						Name:  "zxcv",
						Value: "5678",
						Path:  "/",
					},
				},
			},
		},
		{
			"without set-cookie path, but url under subdirectory",
			[]SetCookie{
				{
					[]*http.Cookie{
						{
							Name:   "asdf",
							Value:  "1234",
							Secure: true,
						},
					},
					panicParseURL("https://example.com/test/1234"),
				},
			},
			CookiesMap{
				panicParseURL("https://example.com/test"): {
					{
						Name:   "asdf",
						Value:  "1234",
						Secure: true,
						Path:   "/test",
					},
				},
			},
		},
		{
			"same name & value with another domain",
			[]SetCookie{
				{
					[]*http.Cookie{
						{
							Name:   "asdf",
							Value:  "1234",
							Secure: true,
						},
					},
					panicParseURL("https://example.com/test/1234"),
				},
				{
					[]*http.Cookie{
						{
							Name:   "asdf",
							Value:  "1234",
							Secure: true,
						},
					},
					panicParseURL("https://example.jp"),
				},
			},
			CookiesMap{
				panicParseURL("https://example.com/test"): {
					{
						Name:   "asdf",
						Value:  "1234",
						Secure: true,
						Path:   "/test",
					},
				},
				panicParseURL("https://example.jp/"): {
					{
						Name:   "asdf",
						Value:  "1234",
						Secure: true,
						Path:   "/",
					},
				},
			},
		},
		{
			"updated value",
			[]SetCookie{
				{
					[]*http.Cookie{
						{
							Name:   "counter",
							Value:  "1",
							Secure: true,
						},
					},
					panicParseURL("https://example.com/test/1234"),
				},
				{
					[]*http.Cookie{
						{
							Name:   "counter",
							Value:  "2",
							Secure: true,
						},
					},
					panicParseURL("https://example.com/test/1234"),
				},
			},
			CookiesMap{
				panicParseURL("https://example.com/test"): {
					{
						Name:   "counter",
						Value:  "2",
						Secure: true,
						Path:   "/test",
					},
				},
			},
		},
		{
			"empty",
			[]SetCookie{},
			CookiesMap{},
		},
		{
			"subdomain uniqueness",
			[]SetCookie{
				{
					[]*http.Cookie{
						{
							Name:   "abcd",
							Value:  "1234",
							Secure: true,
							// including subdomains
							Domain: "example.com",
						},
					},
					panicParseURL("https://example.com/"),
				},
				{
					[]*http.Cookie{
						{
							Name:   "efgh",
							Value:  "1234",
							Secure: true,
						},
					},
					panicParseURL("https://sub.example.com/"),
				},
			},
			CookiesMap{
				panicParseURL("https://example.com/"): {
					{
						Name:   "abcd",
						Value:  "1234",
						Secure: true,
						Path:   "/",
						Domain: "example.com",
					},
				},
				panicParseURL("https://sub.example.com/"): {
					{
						Name:   "efgh",
						Value:  "1234",
						Secure: true,
						Path:   "/",
					},
				},
			},
		},
		{
			"subdomain uniqueness w/ path",
			[]SetCookie{
				{
					[]*http.Cookie{
						{
							Name:  "abcd",
							Value: "1234",
							Path:  "/subdir",
							// including subdomains
							Domain: "example.com",
						},
					},
					panicParseURL("https://example.com/"),
				},
				{
					[]*http.Cookie{
						{
							Name:   "efgh",
							Value:  "1234",
							Secure: true,
							Path:   "/subdir",
						},
					},
					panicParseURL("https://sub.example.com/"),
				},
			},
			CookiesMap{
				panicParseURL("http://example.com/subdir"): {
					{
						Name:   "abcd",
						Value:  "1234",
						Path:   "/subdir",
						Domain: "example.com",
					},
				},
				panicParseURL("https://sub.example.com/subdir"): {
					{
						Name:   "efgh",
						Value:  "1234",
						Secure: true,
						Path:   "/subdir",
					},
				},
			},
		},
		{
			"with Domain=",
			[]SetCookie{
				{
					[]*http.Cookie{
						{
							Name:  "abcd",
							Value: "1234",
							// including subdomains
							Domain: "example.jp",
						},
					},
					panicParseURL("https://example.jp/sub"),
				},
			},
			CookiesMap{
				panicParseURL("http://example.jp/"): {
					{
						Name:   "abcd",
						Value:  "1234",
						Path:   "/",
						Domain: "example.jp",
					},
				},
			},
		},
	} {
		tc := tc
		t.Run(fmt.Sprintf("%d-%s", i, tc.Name), func(t *testing.T) {
			t.Parallel()
			jar, err := cookiejar.New(&cookiejar.Options{})
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			pjar := New(jar)
			for _, setCookie := range tc.SetCookies {
				pjar.SetCookies(setCookie.URL, setCookie.Cookies)
			}
			actualOriginal := pjar.AllCookies()
			expected := convertCookieMapToStringMap(tc.Expected)
			actual := convertCookieMapToStringMap(actualOriginal)
			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("allCookies failed:\n%#v expected, but\n%#v returned", expected, actual)
				t.FailNow()
			}

			// check restoring
			restoreJar, err := cookiejar.New(&cookiejar.Options{})
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			rpjar := New(restoreJar)
			for url, cookies := range actualOriginal {
				rpjar.SetCookies(url, cookies)
			}
			restoredActualOriginal := rpjar.AllCookies()
			restoredActual := convertCookieMapToStringMap(restoredActualOriginal)
			if !reflect.DeepEqual(expected, restoredActual) {
				t.Errorf("reverse restoring failed:\n%#v expected, but\n%#v returned", expected, actual)
			}
		})
	}
}
