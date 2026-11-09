package wallpaper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestSyncDeduplicates(t *testing.T) {
	var imageRequests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/feed" {
			json.NewEncoder(w).Encode(feed{Images: []imageInfo{
				{URL: "/image.jpg", URLBase: "same-image"},
				{URL: "/image.jpg", URLBase: "same-image"},
			}})
			return
		}
		atomic.AddInt32(&imageRequests, 1)
		w.Header().Set("Content-Type", "image/jpeg")
		fmt.Fprint(w, "image")
	}))
	defer server.Close()
	dir := t.TempDir()
	for i := 0; i < 2; i++ {
		if err := syncFromURL(context.Background(), server.Client(), server.URL+"/feed", dir); err != nil {
			t.Fatal(err)
		}
	}
	files, err := os.ReadDir(filepath.Join(dir, "wallpaper"))
	if err != nil || len(files) != 1 || atomic.LoadInt32(&imageRequests) != 1 {
		t.Fatalf("files=%d requests=%d err=%v", len(files), imageRequests, err)
	}
}

func TestPruneKeepsThirtyNewest(t *testing.T) {
	dir := t.TempDir()
	base := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 31; i++ {
		name := fmt.Sprintf("%02d.jpg", i)
		filename := filepath.Join(dir, name)
		if err := os.WriteFile(filename, []byte("image"), 0644); err != nil {
			t.Fatal(err)
		}
		stamp := base.Add(time.Duration(i) * time.Hour)
		if err := os.Chtimes(filename, stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}
	if err := prune(dir, 30); err != nil {
		t.Fatal(err)
	}
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 30 {
		t.Fatalf("files=%d err=%v", len(files), err)
	}
	if _, err := os.Stat(filepath.Join(dir, "00.jpg")); !os.IsNotExist(err) {
		t.Fatal("oldest wallpaper was not removed")
	}
	if !strings.HasSuffix(files[len(files)-1].Name(), ".jpg") {
		t.Fatal("unexpected file in wallpaper pool")
	}
}
