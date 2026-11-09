package system

import (
	"x-panel/api/api_v1"
	"x-panel/api/api_v1/middleware"

	"github.com/gin-gonic/gin"
)

func InitLogin(router *gin.RouterGroup) {
	loginApi := api_v1.ApiGroupApp.ApiSystem.LoginApi

	router.POST("/login", loginApi.Login)
	router.GET("/captcha/getImage", loginApi.CaptchaImage)
	router.POST("/logout", middleware.LoginInterceptor, loginApi.Logout)

}
