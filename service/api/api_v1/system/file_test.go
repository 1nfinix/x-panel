package system

import (
	"path/filepath"
	"testing"
)

func TestResolveStoredFileSrc(t *testing.T) {
	sourcePath := "./conf/uploads"
	tests := []struct {
		name     string
		src      string
		wantDisk string
		wantURL  string
		wantErr  bool
	}{
		{
			name:     "current format",
			src:      "./conf/uploads/2026/1/26/new.svg",
			wantDisk: "./conf/uploads/2026/1/26/new.svg",
			wantURL:  "/conf/uploads/2026/1/26/new.svg",
		},
		{
			name:     "older format",
			src:      "/2026/1/26/old.svg",
			wantDisk: filepath.FromSlash("conf/uploads/2026/1/26/old.svg"),
			wantURL:  "/uploads/2026/1/26/old.svg",
		},
		{
			name:     "existing favicon download format",
			src:      "conf/uploads/2026/1/26/favicon.svg",
			wantDisk: "conf/uploads/2026/1/26/favicon.svg",
			wantURL:  "/conf/uploads/2026/1/26/favicon.svg",
		},
		{
			name:    "current format traversal",
			src:     "./conf/uploads/../../secret.svg",
			wantErr: true,
		},
		{
			name:    "older format traversal",
			src:     "/../../secret.svg",
			wantErr: true,
		},
		{
			name:    "favicon format traversal",
			src:     "conf/uploads/../../secret.svg",
			wantErr: true,
		},
		{
			name:    "current path outside upload directory",
			src:     "./conf/other/secret.svg",
			wantErr: true,
		},
		{
			name:    "upload directory itself",
			src:     "/",
			wantErr: true,
		},
		{
			name:    "unsupported format",
			src:     "2026/1/26/old.svg",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			disk, url, err := resolveStoredFileSrc(sourcePath, tt.src)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolveStoredFileSrc(%q) = (%q, %q), want error", tt.src, disk, url)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if disk != tt.wantDisk || url != tt.wantURL {
				t.Errorf("resolveStoredFileSrc(%q) = (%q, %q), want (%q, %q)", tt.src, disk, url, tt.wantDisk, tt.wantURL)
			}
		})
	}
}
