package cmn

import "testing"

func TestParseVersionInfo(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    Version_Info
	}{
		{
			name:    "trailing newline",
			content: "11|1.3.0-x.1\n",
			want:    Version_Info{Version_code: 11, Version: "1.3.0-x.1"},
		},
		{
			name:    "surrounding whitespace",
			content: " 11|1.3.0-x.1 \r\n",
			want:    Version_Info{Version_code: 11, Version: "1.3.0-x.1"},
		},
		{
			name:    "missing separator",
			content: "11",
			want:    Version_Info{},
		},
		{
			name:    "empty content",
			content: "",
			want:    Version_Info{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseVersionInfo(tt.content); got != tt.want {
				t.Fatalf("parseVersionInfo(%q) = %+v, want %+v", tt.content, got, tt.want)
			}
		})
	}
}
