package siteFavicon

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGetOneFaviconURLSelectsLargest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><head>
<link rel="icon" sizes="32x32" href="/small.png">
<link rel="apple-touch-icon" sizes="180x180" href="/apple.png">
<link rel="icon" sizes="256x256" href="/large.png">
</head></html>`)
	}))
	defer server.Close()
	got, err := GetOneFaviconURL(server.URL + "/page")
	if err != nil || got != server.URL+"/large.png" {
		t.Fatalf("largest icon = %q, err %v", got, err)
	}
}

func TestDownloadImageEnforcesLimitWithoutContentLength(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		fmt.Fprint(w, strings.Repeat("x", 1025))
	}))
	defer server.Close()
	dir := t.TempDir()
	if _, err := DownloadImageContext(context.Background(), server.URL+"/icon.png", dir, 1024); err == nil {
		t.Fatal("oversized icon was accepted")
	}
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 0 {
		t.Fatalf("partial download remained: %v, err=%v", files, err)
	}
}

func TestGetOneFaviconURLAppleAndFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/plain" {
			fmt.Fprint(w, `<link rel="icon" href="/icon.ico"><link rel="apple-touch-icon" href="/apple.png">`)
		} else {
			fmt.Fprint(w, `<html></html>`)
		}
	}))
	defer server.Close()
	for _, tt := range []struct{ page, want string }{
		{"/plain", "/apple.png"},
		{"/missing", "/favicon.ico"},
	} {
		got, err := GetOneFaviconURL(server.URL + tt.page)
		if err != nil || got != server.URL+tt.want {
			t.Fatalf("%s = %q, err %v", tt.page, got, err)
		}
	}
}
