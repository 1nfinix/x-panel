package siteFavicon

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetOneFaviconURLSelectsLargest(t *testing.T) {
	allowPrivateIPs = true
	t.Cleanup(func() { allowPrivateIPs = false })
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
	allowPrivateIPs = true
	t.Cleanup(func() { allowPrivateIPs = false })
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
	allowPrivateIPs = true
	t.Cleanup(func() { allowPrivateIPs = false })
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

func TestRejectNonPublicAddresses(t *testing.T) {
	for _, raw := range []string{"127.0.0.1", "10.1.2.3", "169.254.0.23", "100.64.1.1", "::1", "fc00::1", "fe80::1"} {
		if publicIP(netip.MustParseAddr(raw)) {
			t.Errorf("accepted %s", raw)
		}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "ok") }))
	defer server.Close()
	if _, err := GetOneFaviconURL(server.URL); err == nil {
		t.Fatal("default client accepted loopback")
	}
}

func TestRedirectAndImageContent(t *testing.T) {
	png, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/lQAAAABJRU5ErkJggg==")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect":
			allowPrivateIPs = false
			http.Redirect(w, r, "http://127.0.0.1:12345/private", http.StatusFound)
		case "/html":
			fmt.Fprint(w, "<html><script>alert(1)</script></html>")
		case "/svg":
			fmt.Fprint(w, "<svg xmlns=\"http://www.w3.org/2000/svg\"></svg>")
		default:
			w.Write(png)
		}
	}))
	defer server.Close()
	allowPrivateIPs = true
	dir := t.TempDir()
	for _, path := range []string{"/html", "/svg"} {
		if _, err := DownloadImageContext(context.Background(), server.URL+path, dir, 1024); err == nil {
			t.Fatalf("accepted %s", path)
		}
	}
	file, err := DownloadImageContext(context.Background(), server.URL+"/ok.svg", dir, 1024)
	if err != nil || filepath.Ext(file.Name()) != ".png" {
		t.Fatalf("png detection: %v, %v", file, err)
	}
	allowPrivateIPs = true
	if _, err := GetOneFaviconURLContext(context.Background(), server.URL+"/redirect"); err == nil {
		t.Fatal("accepted redirect to loopback")
	}
}
