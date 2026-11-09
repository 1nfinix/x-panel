package system

import (
	"encoding/base64"
	"net/http"
	"strings"
	"x-panel/lib/captcha"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (l LoginApi) CaptchaImage(c *gin.Context) {
	id := uuid.NewString()
	encoded, err := captcha.GenerateCaptchaHandler(id, 160, 48)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	parts := strings.SplitN(encoded, ",", 2)
	if len(parts) != 2 {
		c.Status(http.StatusInternalServerError)
		return
	}
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("CaptchaId", id, 300, "/api", "", c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https", true)
	c.Data(http.StatusOK, "image/png", data)
}
