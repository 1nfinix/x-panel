package openness

import (
	"x-panel/api/api_v1/common/apiReturn"
	"x-panel/global"
	"x-panel/lib/wallpaper"

	"github.com/gin-gonic/gin"
)

func (a *Openness) Wallpapers(c *gin.Context) {
	urls, err := wallpaper.PublicURLs(global.Config.GetValueString("base", "source_path"))
	if err != nil {
		apiReturn.Error(c, err.Error())
		return
	}
	apiReturn.SuccessData(c, urls)
}
