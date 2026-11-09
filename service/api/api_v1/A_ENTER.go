package api_v1

import (
	"x-panel/api/api_v1/openness"
	"x-panel/api/api_v1/panel"
	"x-panel/api/api_v1/system"
)

type ApiGroup struct {
	ApiSystem system.ApiSystem // 系统功能api
	ApiOpen   openness.ApiPpenness
	ApiPanel  panel.ApiPanel
}

var ApiGroupApp = new(ApiGroup)
