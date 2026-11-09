package system

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"testing"
)

func TestValidUploadedImage(t *testing.T) {
	for _, tt := range []struct {
		name  string
		data  []byte
		valid bool
	}{
		{"fake.png", []byte("<html><script>alert(1)</script></html>"), false},
		{"icon.svg", []byte("<svg></svg>"), false},
		{"real.png", []byte("\x89PNG\r\n\x1a\nrest"), true},
	} {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("imgfile", tt.name)
		if err != nil {
			t.Fatal(err)
		}
		part.Write(tt.data)
		writer.Close()
		req := httptest.NewRequest("POST", "/", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		if err := req.ParseMultipartForm(1024); err != nil {
			t.Fatal(err)
		}
		got := validUploadedImage(req.MultipartForm.File["imgfile"][0])
		if got != tt.valid {
			t.Errorf("%s valid = %v", tt.name, got)
		}
	}
}
