package systemSettingCache

import (
	"x-panel/global"
	"x-panel/lib/cmn/systemSetting"
	"time"
)

func InItSystemSettingCache() *systemSetting.SystemSettingCache {
	return &systemSetting.SystemSettingCache{
		Cache: global.NewCache[interface{}](5*time.Hour, -1, "systemSettingCache"),
	}
}
