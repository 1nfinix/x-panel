package system

import (
	"bytes"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"x-panel/lib/captcha"

	"github.com/gin-gonic/gin"
)

func TestCaptchaImageCookieAndSingleUse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/captcha/getImage", (LoginApi{}).CaptchaImage)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/captcha/getImage", nil))
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("image response status=%d type=%q", response.Code, response.Header().Get("Content-Type"))
	}
	decoded, err := png.Decode(bytes.NewReader(response.Body.Bytes()))
	if err != nil {
		t.Fatalf("captcha PNG cannot be decoded: %v", err)
	}
	if bounds := decoded.Bounds(); bounds.Dx() != 160 || bounds.Dy() != 48 {
		t.Fatalf("captcha dimensions = %dx%d, want 160x48", bounds.Dx(), bounds.Dy())
	}
	var id string
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == "CaptchaId" {
			id = cookie.Value
			if !cookie.HttpOnly || cookie.Path != "/api" {
				t.Fatal("captcha cookie must be HttpOnly on /api")
			}
		}
	}
	if id == "" {
		t.Fatal("missing captcha ID cookie")
	}
	answer := captcha.Store.Get(id, false)
	if answer == "" || !captcha.CaptchaVerifyHandle(id, strings.ToLower(answer)) || captcha.CaptchaVerifyHandle(id, answer) {
		t.Fatal("captcha must validate once, ignoring case")
	}
}
