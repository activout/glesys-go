package glesys

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error { b.closed = true; return nil }
func TestResponseLifecycleAndErrors(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
		result     any
		wantErr    bool
	}{{"nil result", `{"response":{}}`, 200, nil, false}, {"malformed", "oops", 200, nil, true}, {"empty", "", 200, nil, true}, {"provider error", `{"response":{"status":{"code":409,"text":"conflict"}}}`, 200, nil, true}, {"large", strings.Repeat(" ", maxResponseBytes+1), 200, nil, true}} {
		t.Run(tc.name, func(t *testing.T) {
			b := &trackedBody{Reader: strings.NewReader(tc.body)}
			err := parseResponseBody(&http.Response{Body: b, StatusCode: tc.status, Header: http.Header{}}, tc.result)
			if (err != nil) != tc.wantErr || !b.closed {
				t.Fatalf("err=%v closed=%v", err, b.closed)
			}
		})
	}
	b := &trackedBody{Reader: strings.NewReader("not JSON")}
	err := handleResponseError(&http.Response{Body: b, StatusCode: 429, Header: http.Header{"Retry-After": []string{"1"}}})
	var api *APIError
	if !errors.As(err, &api) || api.HTTPStatus != 429 || api.RetryAfter != "1" || !b.closed {
		t.Fatalf("structured error: %v", err)
	}
}
func TestPublicRequestsAndPatches(t *testing.T) {
	var paths []string
	var bodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.EscapedPath())
		var body map[string]any
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&body)
		}
		bodies = append(bodies, body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"response":{"status":{"code":200},"emailaccount":{"emailaccount":"a@example.com"},"record":{},"server":{},"id":9007199254740993123,"future":{"a":true}}}`)
	}))
	defer server.Close()
	c := NewClient("project", "key", "test")
	_ = c.SetBaseURL(server.URL)
	c.SetHTTPClient(server.Client())
	zero := 0
	empty := ""
	no := "no"
	_, err := c.EmailDomains.EditAccountPatch(context.Background(), "a@example.com", EmailAccountPatch{AntiSpamLevel: &zero, AutoRespond: &no, AutoRespondMessage: &empty})
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := bodies[0]["antispamlevel"]; !ok || v != float64(0) {
		t.Fatal("zero lost")
	}
	if bodies[0]["autorespond"] != "no" || bodies[0]["autorespondmessage"] != "" {
		t.Fatal("explicit values lost")
	}
	if _, ok := bodies[0]["quotaingib"]; ok {
		t.Fatal("omitted value sent")
	}
	raw, err := c.Request(context.Background(), "POST", "email/quota", map[string]any{"emailaccount": "a@example.com"})
	if err != nil || !strings.Contains(string(raw), "9007199254740993123") || !strings.Contains(string(raw), "future") {
		t.Fatalf("raw response: %s %v", raw, err)
	}
	_, err = c.Request(context.Background(), "GET", "email/list", map[string]any{"filter": "a/b@c", "domainname": "example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(paths[2], "a%2Fb@c") {
		t.Fatalf("GET escaping: %s", paths[2])
	}
	before := len(paths)
	for _, endpoint := range []string{"https://evil.test/x", "//evil.test/x", "../email/quota", "email/../quota", "email/quota?x=y", "email/%2e%2e", "email\\quota"} {
		if _, err = c.Request(context.Background(), "POST", endpoint, nil); err == nil {
			t.Fatalf("accepted %s", endpoint)
		}
	}
	if len(paths) != before {
		t.Fatal("invalid endpoint reached network")
	}
	_, err = c.EmailDomains.CreateAccountWithSettings(context.Background(), "a@example.com", EmailAccountPatch{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := bodies[len(bodies)-1]["password"]; ok {
		t.Fatal("undocumented password sent")
	}
}
func TestRedirectsNeverCarryCredentials(t *testing.T) {
	reached := false
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true }))
	defer other.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL, http.StatusTemporaryRedirect)
	}))
	defer source.Close()
	c := NewClient("project", "key", "test")
	_ = c.SetBaseURL(source.URL)
	_, err := c.Request(context.Background(), "GET", "email/costs", nil)
	if err == nil || reached {
		t.Fatalf("redirect followed: err=%v reached=%v", err, reached)
	}
}
func TestOverviewFilterEscaping(t *testing.T) {
	var path string
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.EscapedPath()
		_, _ = io.WriteString(w, `{"response":{"overview":{}}}`)
	}))
	defer s.Close()
	c := NewClient("p", "k", "test")
	_ = c.SetBaseURL(s.URL)
	_, err := c.EmailDomains.Overview(context.Background(), OverviewParams{Filter: "a/b?c#d", Page: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(path, "a%2Fb%3Fc%23d") {
		t.Fatalf("filter escaped incorrectly: %s", path)
	}
}

func TestCostDetailsPreserveNumericAndStringAmounts(t *testing.T) {
	for _, amount := range []string{`0.125`, `"0.125"`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, `{"response":{"costs":{},"pricelist":{"quota":{"amount":`+amount+`,"currency":"SEK","unit":"GiB"}}}}`)
		}))
		c := NewClient("p", "k", "test")
		_ = c.SetBaseURL(server.URL)
		v, err := c.EmailDomains.CostsDetails(context.Background())
		server.Close()
		if err != nil || !strings.Contains(string(v.PriceList), `"amount":`+amount) {
			t.Fatalf("amount changed: %s %v", v.PriceList, err)
		}
	}
}
