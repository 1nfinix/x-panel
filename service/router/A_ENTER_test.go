package router

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWebPathForExecutable(t *testing.T) {
	root := t.TempDir()
	release := filepath.Join(root, "releases", "v1.3.0-x.1")
	if err := os.MkdirAll(filepath.Join(release, "web"), 0755); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(release, "x-panel")
	if err := os.WriteFile(executable, nil, 0755); err != nil {
		t.Fatal(err)
	}
	current := filepath.Join(root, "current")
	if err := os.Symlink(release, current); err != nil {
		t.Fatal(err)
	}

	resolvedRelease, err := filepath.EvalSymlinks(release)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := webPathForExecutable(filepath.Join(current, "x-panel")), filepath.Join(resolvedRelease, "web"); got != want {
		t.Fatalf("web path through current symlink = %q, want %q", got, want)
	}
	if err := os.Remove(filepath.Join(release, "web")); err != nil {
		t.Fatal(err)
	}
	if got := webPathForExecutable(filepath.Join(current, "x-panel")); got != "./web" {
		t.Fatalf("web path without executable-side directory = %q, want ./web", got)
	}
}

func TestWebRootFiles(t *testing.T) {
	webPath := t.TempDir()
	for _, name := range []string{"index.html", "favicon.ico", "apple-touch-icon.png", ".hidden"} {
		if err := os.WriteFile(filepath.Join(webPath, name), nil, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(webPath, "assets"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(webPath, "favicon.ico"), filepath.Join(webPath, "linked.ico")); err != nil {
		t.Fatal(err)
	}

	files, err := webRootFiles(webPath)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"apple-touch-icon.png", "favicon.ico"}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("web root files = %q, want %q", files, want)
	}
}
