package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const reply = `{"id":"msg_1","type":"message","role":"assistant","model":"claude-haiku-5-5",
"content":[{"type":"text","text":"Hi there, nice to meet."}],"stop_reason":"end_turn","stop_sequence":null,
"usage":{"input_tokens":1000,"output_tokens":200,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}`

func fakeAPI(t *testing.T, gotKey, gotBody *string, calls *int) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*calls++
		*gotKey = r.Header.Get("X-Api-Key")
		b, _ := io.ReadAll(r.Body)
		*gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, reply)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestRunUsesOnlyTheExplicitKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "must-not-be-used")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "must-not-be-used")
	var key, body string
	var calls int
	srv := fakeAPI(t, &key, &body, &calls)
	env := map[string]string{KeyEnv: "explicit-key"}
	var out bytes.Buffer
	if err := run(context.Background(), []string{"hello"}, func(k string) string { return env[k] }, &out, srv.URL); err != nil {
		t.Fatal(err)
	}
	if key != "explicit-key" {
		t.Fatalf("sent key %q", key)
	}
	// 1000 in × $0.10 + 200 out × $0.50, per million tokens.
	if !strings.Contains(out.String(), "estimated cost: $0.000200") || !strings.Contains(out.String(), "Hi there") {
		t.Fatalf("output:\n%s", out.String())
	}
	for _, bad := range []string{"temperature", "top_p", "top_k", "fallbacks"} {
		if strings.Contains(body, bad) {
			t.Fatalf("request sent %s: %s", bad, body)
		}
	}
	if strings.Contains(out.String(), "explicit-key") {
		t.Fatal("key printed")
	}
}

func TestRunNeedsKey(t *testing.T) {
	err := run(context.Background(), []string{"hi"}, func(string) string { return "" }, io.Discard, "")
	if err == nil || !strings.Contains(err.Error(), KeyEnv) {
		t.Fatalf("got %v", err)
	}
}

func TestRunNoSDKRetries(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(529)
		_, _ = io.WriteString(w, `{"type":"error","error":{"type":"overloaded_error","message":"x"}}`)
	}))
	defer srv.Close()
	env := map[string]string{KeyEnv: "k"}
	err := run(context.Background(), []string{"hi"}, func(k string) string { return env[k] }, io.Discard, srv.URL)
	if err == nil || !strings.Contains(err.Error(), "529") || calls != 1 {
		t.Fatalf("err %v calls %d", err, calls)
	}
}
