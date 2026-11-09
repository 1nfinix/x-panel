package imagecheck

import (
	"errors"
	"net/http"
	"strings"
)

// Extension returns the extension implied by image bytes, not by a supplied filename.
func Extension(data []byte) (string, error) {
	if len(data) >= 4 && data[0] == 0 && data[1] == 0 && data[2] == 1 && data[3] == 0 {
		return ".ico", nil
	}
	contentType := http.DetectContentType(data)
	switch strings.Split(contentType, ";")[0] {
	case "image/png":
		return ".png", nil
	case "image/jpeg":
		return ".jpg", nil
	case "image/gif":
		return ".gif", nil
	case "image/webp":
		return ".webp", nil
	case "image/x-icon", "image/vnd.microsoft.icon":
		return ".ico", nil
	default:
		return "", errors.New("unsupported image content")
	}
}

func ValidExtension(filenameExt, detectedExt string) bool {
	if filenameExt == ".jpeg" && detectedExt == ".jpg" {
		return true
	}
	return filenameExt == detectedExt
}
