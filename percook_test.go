package percook

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"strconv"
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
		SetCookies []SetCookie
		Expected   CookiesMap
	}

	for i, tc := range []TestCase{
		{
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
				// no secure
				panicParseURL("http://example.com/"): {
					{
						Name:  "asdf",
						Value: "1234",
					},
				},
			},
		},
		{
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
					},
				},
				panicParseURL("http://example.com/"): {
					{
						Name:  "zxcv",
						Value: "5678",
					},
				},
			},
		},
		{
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
						Name:  "asdf",
						Value: "1234",
					},
				},
				panicParseURL("http://example.com/"): {
					{
						Name:  "zxcv",
						Value: "5678",
					},
				},
			},
		},
		{
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
				panicParseURL("https://example.com/test/"): {
					{
						Name:  "asdf",
						Value: "1234",
					},
				},
			},
		},
		{
			[]SetCookie{},
			CookiesMap{},
		},
	} {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
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
				t.Errorf("%#v expected, but %#v returned", expected, actual)
			}
		})
	}
}
