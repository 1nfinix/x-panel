package panel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"x-panel/api/api_v1/common/apiData/commonApiStructs"
	"x-panel/api/api_v1/common/apiData/panelApiStructs"
	"x-panel/api/api_v1/common/apiReturn"
	"x-panel/api/api_v1/common/base"
	"x-panel/global"
	"x-panel/lib/siteFavicon"
	"x-panel/models"
	"x-panel/models/datatype"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type ItemIcon struct {
}

func (a *ItemIcon) Edit(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	req := models.ItemIcon{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	if req.ItemIconGroupId == 0 {
		// apiReturn.Error(c, "Group is mandatory")
		apiReturn.ErrorParamFomat(c, "Group is mandatory")
		return
	}
	var groupCount int64
	if err := global.Db.Model(&models.ItemIconGroup{}).Where("id=? AND user_id=?", req.ItemIconGroupId, userInfo.ID).Count(&groupCount).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	if groupCount == 0 {
		apiReturn.ErrorDataNotFound(c)
		return
	}

	req.UserId = userInfo.ID

	// json转字符串
	if j, err := json.Marshal(req.Icon); err == nil {
		req.IconJson = string(j)
	}

	if req.ID != 0 {
		// 修改
		var itemCount int64
		if err := global.Db.Model(&models.ItemIcon{}).Where("id=? AND user_id=?", req.ID, userInfo.ID).Count(&itemCount).Error; err != nil {
			apiReturn.ErrorDatabase(c, err.Error())
			return
		}
		if itemCount == 0 {
			apiReturn.ErrorDataNotFound(c)
			return
		}
		updateField := []string{"IconJson", "Title", "Url", "LanUrl", "Description", "OpenMethod", "ItemIconGroupId"}
		if req.Sort != 0 {
			updateField = append(updateField, "Sort")
		}
		result := global.Db.Model(&models.ItemIcon{}).
			Select(updateField).
			Where("id=? AND user_id=?", req.ID, userInfo.ID).Updates(&req)
		if result.Error != nil {
			apiReturn.ErrorDatabase(c, result.Error.Error())
			return
		}
	} else {
		req.Sort = 9999
		// 创建
		if err := global.Db.Create(&req).Error; err != nil {
			apiReturn.ErrorDatabase(c, err.Error())
			return
		}
	}

	apiReturn.SuccessData(c, req)
}

// 添加多个图标
func (a *ItemIcon) AddMultiple(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	// type Request
	req := []models.ItemIcon{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	for i := 0; i < len(req); i++ {
		if req[i].ItemIconGroupId == 0 {
			apiReturn.ErrorParamFomat(c, "Group is mandatory")
			return
		}
		var groupCount int64
		if err := global.Db.Model(&models.ItemIconGroup{}).Where("id=? AND user_id=?", req[i].ItemIconGroupId, userInfo.ID).Count(&groupCount).Error; err != nil {
			apiReturn.ErrorDatabase(c, err.Error())
			return
		}
		if groupCount == 0 {
			apiReturn.ErrorDataNotFound(c)
			return
		}
		req[i].UserId = userInfo.ID
		// json转字符串
		if j, err := json.Marshal(req[i].Icon); err == nil {
			req[i].IconJson = string(j)
		}
	}

	if err := global.Db.Create(&req).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	apiReturn.SuccessData(c, req)
}

// // 获取详情
// func (a *ItemIcon) GetInfo(c *gin.Context) {
// 	req := systemApiStructs.AiDrawGetInfoReq{}

// 	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
// 		apiReturn.ErrorParamFomat(c, err.Error())
// 		return
// 	}

// 	userInfo, _ := base.GetCurrentUserInfo(c)

// 	aiDraw := models.AiDraw{}
// 	aiDraw.ID = req.ID
// 	if err := aiDraw.GetInfo(global.Db); err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			apiReturn.Error(c, "不存在记录")
// 			return
// 		}
// 		apiReturn.ErrorDatabase(c, err.Error())
// 		return
// 	}

// 	if userInfo.ID != aiDraw.UserID {
// 		apiReturn.ErrorNoAccess(c)
// 		return
// 	}

// 	apiReturn.SuccessData(c, aiDraw)
// }

func (a *ItemIcon) GetListByGroupId(c *gin.Context) {
	req := models.ItemIcon{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	userInfo, _ := base.GetCurrentUserInfo(c)
	itemIcons := []models.ItemIcon{}

	if err := global.Db.Order("sort ,created_at").Find(&itemIcons, "item_icon_group_id = ? AND user_id=?", req.ItemIconGroupId, userInfo.ID).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	for k, v := range itemIcons {
		json.Unmarshal([]byte(v.IconJson), &itemIcons[k].Icon)
	}

	apiReturn.SuccessListData(c, itemIcons, 0)
}

func (a *ItemIcon) Deletes(c *gin.Context) {
	req := commonApiStructs.RequestDeleteIds[uint]{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	userInfo, _ := base.GetCurrentUserInfo(c)
	if err := global.Db.Delete(&models.ItemIcon{}, "id in ? AND user_id=?", req.Ids, userInfo.ID).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	apiReturn.Success(c)
}

// 保存排序
func (a *ItemIcon) SaveSort(c *gin.Context) {
	req := panelApiStructs.ItemIconSaveSortRequest{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	userInfo, _ := base.GetCurrentUserInfo(c)

	transactionErr := global.Db.Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		for _, v := range req.SortItems {
			if err := tx.Model(&models.ItemIcon{}).Where("user_id=? AND id=? AND item_icon_group_id=?", userInfo.ID, v.Id, req.ItemIconGroupId).Update("sort", v.Sort).Error; err != nil {
				// 返回任何错误都会回滚事务
				return err
			}
		}

		// 返回 nil 提交事务
		return nil
	})

	if transactionErr != nil {
		apiReturn.ErrorDatabase(c, transactionErr.Error())
		return
	}

	apiReturn.Success(c)
}

// 支持获取并直接下载对方网站图标到服务器
func (a *ItemIcon) GetSiteFavicon(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	req := panelApiStructs.ItemIconGetSiteFaviconReq{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	src, _, err := downloadFavicon(c.Request.Context(), userInfo.ID, req.Url)
	if err != nil {
		apiReturn.Error(c, "acquisition failed: "+err.Error())
		return
	}
	apiReturn.SuccessData(c, panelApiStructs.ItemIconGetSiteFaviconResp{IconUrl: src})
}

type faviconCandidate struct {
	ID    uint                      `json:"id"`
	Title string                    `json:"title"`
	URL   string                    `json:"url"`
	Icon  datatype.ItemIconIconInfo `json:"icon"`
}

type faviconFillResult struct {
	ID      uint   `json:"id"`
	Success bool   `json:"success"`
	Reason  string `json:"reason,omitempty"`
	IconURL string `json:"iconUrl,omitempty"`
}

func candidateIcon(item models.ItemIcon) (datatype.ItemIconIconInfo, bool) {
	icon := datatype.ItemIconIconInfo{}
	if json.Unmarshal([]byte(item.IconJson), &icon) != nil {
		return icon, false
	}
	return icon, icon.ItemType == 1 || (icon.ItemType == 2 && strings.TrimSpace(icon.Src) == "")
}

func (a *ItemIcon) FaviconCandidates(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	items := []models.ItemIcon{}
	if err := global.Db.Where("user_id=?", userInfo.ID).Order("sort, id").Find(&items).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	list := make([]faviconCandidate, 0)
	for _, item := range items {
		if icon, ok := candidateIcon(item); ok {
			list = append(list, faviconCandidate{ID: item.ID, Title: item.Title, URL: item.Url, Icon: icon})
		}
	}
	apiReturn.SuccessListData(c, list, int64(len(list)))
}

func (a *ItemIcon) FillFavicons(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	results := make([]faviconFillResult, 0, len(req.IDs))
	for _, id := range req.IDs {
		result := faviconFillResult{ID: id}
		item := models.ItemIcon{}
		if err := global.Db.Where("id=? AND user_id=?", id, userInfo.ID).First(&item).Error; err != nil {
			result.Reason = "item not found"
		} else if icon, ok := candidateIcon(item); !ok {
			result.Reason = "icon is already set or is not an eligible type"
		} else {
			src, file, err := downloadFavicon(c.Request.Context(), userInfo.ID, item.Url)
			if err != nil {
				result.Reason = err.Error()
			} else {
				icon.ItemType, icon.Src = 2, src
				updated, _ := json.Marshal(icon)
				dbResult := global.Db.Model(&models.ItemIcon{}).Where("id=? AND user_id=? AND icon_json=?", id, userInfo.ID, item.IconJson).Update("icon_json", string(updated))
				if dbResult.Error != nil || dbResult.RowsAffected == 0 {
					os.Remove(file.Src)
					global.Db.Delete(&file)
					result.Reason = "item changed during download"
					if dbResult.Error != nil {
						result.Reason = dbResult.Error.Error()
					}
				} else {
					result.Success, result.IconURL = true, src
				}
			}
		}
		results = append(results, result)
	}
	apiReturn.SuccessData(c, results)
}

func downloadFavicon(parent context.Context, userID uint, website string) (string, models.File, error) {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	iconURL, err := siteFavicon.GetOneFaviconURLContext(ctx, website)
	if err != nil {
		return "", models.File{}, err
	}
	sourcePath := global.Config.GetValueString("base", "source_path")
	if !strings.HasPrefix(sourcePath, "./") {
		return "", models.File{}, errors.New("source_path must start with ./")
	}
	now := time.Now()
	savePath := filepath.Join(sourcePath, fmt.Sprintf("%d/%d/%d", now.Year(), now.Month(), now.Day()))
	if err := os.MkdirAll(savePath, 0755); err != nil {
		return "", models.File{}, err
	}
	imageFile, err := siteFavicon.DownloadImageContext(ctx, iconURL, savePath, 1024*1024)
	if err != nil {
		return "", models.File{}, err
	}
	parsedWebsite, _ := url.Parse(website)
	storedPath := "./" + filepath.ToSlash(imageFile.Name())
	file, err := (&models.File{}).AddFile(userID, parsedWebsite.Host, path.Ext(imageFile.Name()), storedPath)
	if err != nil {
		os.Remove(imageFile.Name())
		return "", models.File{}, err
	}
	return storedPath[1:], file, nil
}
