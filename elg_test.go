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

func TestAPIErrorKeepsStatusAndRetryAfter(t *testing.T) {
	b := &trackedBody{Reader: strings.NewReader("not JSON")}
	err := handleResponseError(&http.Response{Body: b, StatusCode: 429, Header: http.Header{"Retry-After": []string{"1"}}})
	var api *APIError
	if !errors.As(err, &api) || api.HTTPStatus != 429 || api.RetryAfter != "1" || !b.closed {
		t.Fatalf("structured error: %v", err)
	}
	b = &trackedBody{Reader: strings.NewReader(`{"response":{"status":{"code":409,"text":" conflict "}}}`)}
	err = handleResponseError(&http.Response{Body: b, StatusCode: 409, Header: http.Header{}})
	if !errors.As(err, &api) || api.ProviderStatus != 409 || api.Message != "conflict" {
		t.Fatalf("provider status: %v", err)
	}
}

func TestPatches(t *testing.T) {
	var bodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		bodies = append(bodies, body)
		_, _ = io.WriteString(w, `{"response":{"status":{"code":200},"emailaccount":{"emailaccount":"a@example.com"}}}`)
	}))
	defer server.Close()
	c := NewClient("project", "key", "test")
	_ = c.SetBaseURL(server.URL)
	zero, empty, no := 0, "", "no"
	if _, err := c.EmailDomains.EditAccountPatch(context.Background(), "a@example.com", EmailAccountPatch{AntiSpamLevel: &zero, AutoRespond: &no, AutoRespondMessage: &empty}); err != nil {
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
	if _, err := c.EmailDomains.CreateAccountWithSettings(context.Background(), "a@example.com", EmailAccountPatch{}); err != nil {
		t.Fatal(err)
	}
	if _, ok := bodies[1]["password"]; ok {
		t.Fatal("undocumented password sent")
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
