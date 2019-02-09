package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/fanatic/robot-game/server"
	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/require"
	"github.com/yudai/gojsondiff"
	"gopkg.in/yudai/gojsondiff.v1/formatter"
)

var TestRouter http.Handler
var TestGame *server.Game

func setup(t *testing.T) {
	var err error

	TestGame, err = server.NewGame("unit-test-api.db")
	require.NoError(t, err)

	TestRouter, err = server.New(TestGame)
	require.NoError(t, err)
}

func teardown() {
	TestGame.Close()
	os.Remove("unit-test-api.db")
}

func newAPI(t *testing.T) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		// prepend this url to all requests
		BaseURL: "https://example.com",

		// use http.Client with a cookie jar and timeout
		Client: &http.Client{
			Jar:       httpexpect.NewJar(),
			Timeout:   time.Second * 2,
			Transport: httpexpect.NewBinder(TestRouter),

			// do *not* follow redirects. UGH!
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},

		// use fatal failures
		Reporter: httpexpect.NewRequireReporter(t),

		// use verbose logging
		// Printers: []httpexpect.Printer{
		// 	httpexpect.NewCurlPrinter(t),
		// 	httpexpect.NewDebugPrinter(t, true),
		// },
	})
}

func sanitize(a interface{}, keepIDFields bool) {
	switch b := a.(type) {
	case map[string]interface{}:
		for k, v := range b {
			switch v := v.(type) {
			case []interface{}:
				sanitize(v, keepIDFields)
				continue
			case map[string]interface{}:
				sanitize(v, keepIDFields)
				continue
			default:
				if k == "id" && !keepIDFields {
					delete(b, k)
				} else if k == "created_at" || k == "resource_id" || k == "updated_at" {
					delete(b, k)
				}
			}
		}

		for _, providerKey := range []string{"plan_id", "vendor_addon_id", "vendor_app_id"} {
			if v, exist := b[providerKey]; exist && v != nil {
				b[providerKey] = "__PRESENT__"
			}
		}
		if _, ok := b["message"]; ok {
			// error
			delete(b, "url")
		}
		if v, ok := b["api_key"]; ok && v != nil {
			b["api_key"] = "__PRESENT__"
		}
		if _, ok := b["config"]; ok {
			if config, ok := b["config"].(map[string]interface{}); ok {
				newConfig := map[string]interface{}{}
				for k, v := range config {
					if v != nil {
						newConfig[k] = "__PRESENT__"
					}
				}
				b["config"] = newConfig
			}
		}

	case []interface{}:
		for _, v := range b {
			switch v := v.(type) {
			case map[string]interface{}:
				delete(v, "app_ids")
			}
			sanitize(v, keepIDFields)
		}
	}
}

func getJSON(str string) interface{} {
	var value interface{}
	if err := json.Unmarshal([]byte(str), &value); err != nil {
		return nil
	}
	return value
}

func assertEqualJSON(t *testing.T, actual interface{}, expected string) {
	exp := getJSON(expected)
	sanitize(actual, false)
	if !reflect.DeepEqual(exp, actual) {
		t.Errorf("\nexpected value equal to:\n%s\n\nbut got:\n%s\n\ndiff:\n%s",
			dumpValue(exp),
			dumpValue(actual),
			diffValues(exp, actual))
	}
}

func assertEqualErrorJSON(t *testing.T, actual interface{}, expected string) {
	exp := getJSON(expected)
	sanitize(actual, true)
	if !reflect.DeepEqual(exp, actual) {
		t.Errorf("\nexpected value equal to:\n%s\n\nbut got:\n%s\n\ndiff:\n%s",
			dumpValue(exp),
			dumpValue(actual),
			diffValues(exp, actual))
	}
}

func dumpValue(value interface{}) string {
	b, err := json.MarshalIndent(value, " ", "  ")
	if err != nil {
		return " " + fmt.Sprintf("%#v", value)
	}
	return " " + string(b)
}

func diffValues(expected, actual interface{}) string {
	differ := gojsondiff.New()

	var diff gojsondiff.Diff

	if ve, ok := expected.(map[string]interface{}); ok {
		if va, ok := actual.(map[string]interface{}); ok {
			diff = differ.CompareObjects(ve, va)
		} else {
			return " (unavailable)"
		}
	} else if ve, ok := expected.([]interface{}); ok {
		if va, ok := actual.([]interface{}); ok {
			diff = differ.CompareArrays(ve, va)
		} else {
			return " (unavailable)"
		}
	} else {
		return " (unavailable)"
	}

	config := formatter.AsciiFormatterConfig{
		ShowArrayIndex: true,
	}
	formatter := formatter.NewAsciiFormatter(expected, config)

	str, err := formatter.Format(diff)
	if err != nil {
		return " (unavailable)"
	}

	return "--- expected\n+++ actual\n" + str
}

// returns id field if it exists in response
func assertResponse(t *testing.T, r *httpexpect.Request, respBody string, status int) string {
	resp := r.Expect().Status(status).JSON().Raw()

	// return id field if response is a single map[string]interface with an ID field
	var id string
	if rm, ok := resp.(map[string]interface{}); ok {
		// ignore failure to map or cast - id will just remain blank
		id, _ = rm["id"].(string)
	}
	assertEqualJSON(t, resp, respBody)
	return id
}

func POST(t *testing.T, path, reqBody string) *httpexpect.Request {
	u, err := url.Parse(path)
	require.Nil(t, err)
	return newAPI(t).
		POST(u.Path).
		WithQueryString(u.RawQuery).
		WithText(reqBody)
}

// automatically handles query params in the path string
func GET(t *testing.T, path string) *httpexpect.Request {
	u, err := url.Parse(path)
	require.Nil(t, err)
	return newAPI(t).
		GET(u.Path).
		WithQueryString(u.RawQuery)
}

func PATCH(t *testing.T, path, reqBody string) *httpexpect.Request {
	u, err := url.Parse(path)
	require.Nil(t, err)
	return newAPI(t).
		PATCH(u.Path).
		WithQueryString(u.RawQuery).
		WithJSON(getJSON(reqBody))
}

func PUT(t *testing.T, path, reqBody string) *httpexpect.Request {
	u, err := url.Parse(path)
	require.Nil(t, err)
	return newAPI(t).
		PUT(u.Path).
		WithQueryString(u.RawQuery).
		WithJSON(getJSON(reqBody))
}

func DELETE(t *testing.T, path string) *httpexpect.Request {
	u, err := url.Parse(path)
	require.Nil(t, err)
	return newAPI(t).
		DELETE(u.Path).
		WithQueryString(u.RawQuery)
}

// TODO(jw): move all these into the better structure using above funcs
func assertBadBody(t *testing.T, method, path string) {
	text := newAPI(t).
		Request(method, path).
		WithHeader("Content-Type", "application/json; charset=utf-8").
		WithBytes([]byte(`{asdf}`)).
		Expect().
		Status(http.StatusBadRequest).
		Text().Raw() // TODO(jp): Errors should be application/json
	actual := getJSON(text)
	assertEqualErrorJSON(t, actual, `
		{
			"id":"bad_request",
			"message":"Bad parameter: invalid character 'a' looking for beginning of object key string"
		}
	`)
}

func assertEmptyBody(t *testing.T, method, path string) {
	text := newAPI(t).
		Request(method, path).
		WithHeader("Content-Type", "application/json; charset=utf-8").
		Expect().
		Status(http.StatusBadRequest).
		Text().Raw() // TODO(jp): Errors should be application/json
	actual := getJSON(text)
	assertEqualErrorJSON(t, actual, `
		{
			"id":"bad_request",
			"message":"Bad parameter: Missing request body"
		}
	`)
}

func assertIDNotFound(t *testing.T, method, path, typ string) {
	text := newAPI(t).
		Request(method, path).
		WithJSON(getJSON(`{"test":"nothing"}`)). //Pass bogus BODY for PUT/POST requests that expect bad ID
		Expect().
		Status(http.StatusNotFound).
		Text().Raw() // TODO(jp): Errors should be application/json
	actual := getJSON(text)
	assertEqualErrorJSON(t, actual, fmt.Sprintf(`{"id":"not_found","message":"No such %s exists."}`, typ))
}

func String(s string) *string {
	return &s
}
