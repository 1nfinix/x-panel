package system

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"x-panel/api/api_v1/common/apiData/commonApiStructs"
	"x-panel/api/api_v1/common/apiReturn"
	"x-panel/api/api_v1/common/base"
	"x-panel/global"
	"x-panel/lib/cmn"
	"x-panel/models"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type FileApi struct{}

// resolveStoredFileSrc accepts both the current path format and the older
// upload-directory-relative format. It rejects files outside sourcePath.
func resolveStoredFileSrc(sourcePath, src string) (diskPath, publicURL string, err error) {
	switch {
	case strings.HasPrefix(src, "./"):
		diskPath, publicURL = src, src[1:]
	case strings.HasPrefix(src, "/"):
		diskPath, publicURL = filepath.Join(sourcePath, src), "/uploads"+src
	case strings.HasPrefix(src, filepath.Clean(sourcePath)+string(filepath.Separator)):
		// Earlier favicon downloads stored the cleaned relative path without "./".
		diskPath, publicURL = src, "/"+filepath.ToSlash(src)
	default:
		return "", "", fmt.Errorf("unsupported file src: %q", src)
	}

	root, err := filepath.Abs(sourcePath)
	if err != nil {
		return "", "", err
	}
	target, err := filepath.Abs(diskPath)
	if err != nil {
		return "", "", err
	}
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("file src is outside source_path: %q", src)
	}
	return diskPath, publicURL, nil
}

func (a *FileApi) UploadImg(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	configUpload := global.Config.GetValueString("base", "source_path")
	f, err := c.FormFile("imgfile")
	if err != nil {
		apiReturn.ErrorByCode(c, 1300)
		return
	} else {
		fileExt := strings.ToLower(path.Ext(f.Filename))
		agreeExts := []string{
			".png",
			".jpg",
			".gif",
			".jpeg",
			".webp",
			".svg",
			".ico",
		}

		if !cmn.InArray(agreeExts, fileExt) {
			apiReturn.ErrorByCode(c, 1301)
			return
		}
		fileName := cmn.Md5(fmt.Sprintf("%s%s", f.Filename, time.Now().String()))
		fildDir := fmt.Sprintf("%s/%d/%d/%d/", configUpload, time.Now().Year(), time.Now().Month(), time.Now().Day())
		isExist, _ := cmn.PathExists(fildDir)
		if !isExist {
			os.MkdirAll(fildDir, os.ModePerm)
		}
		filepath := fmt.Sprintf("%s%s%s", fildDir, fileName, fileExt)
		c.SaveUploadedFile(f, filepath)

		// 像数据库添加记录
		mFile := models.File{}
		mFile.AddFile(userInfo.ID, f.Filename, fileExt, filepath)
		apiReturn.SuccessData(c, gin.H{
			"imageUrl": filepath[1:],
		})
	}
}

func (a *FileApi) UploadFiles(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	configUpload := global.Config.GetValueString("base", "source_path")

	form, err := c.MultipartForm()
	if err != nil {
		apiReturn.ErrorByCode(c, 1300)
		return
	}
	files := form.File["files[]"]
	errFiles := []string{}
	succMap := map[string]string{}
	for _, f := range files {
		fileExt := strings.ToLower(path.Ext(f.Filename))
		fileName := cmn.Md5(fmt.Sprintf("%s%s", f.Filename, time.Now().String()))
		fildDir := fmt.Sprintf("%s/%d/%d/%d/", configUpload, time.Now().Year(), time.Now().Month(), time.Now().Day())
		isExist, _ := cmn.PathExists(fildDir)
		if !isExist {
			os.MkdirAll(fildDir, os.ModePerm)
		}
		filepath := fmt.Sprintf("%s%s%s", fildDir, fileName, fileExt)
		if c.SaveUploadedFile(f, filepath) != nil {
			errFiles = append(errFiles, f.Filename)
		} else {
			// 成功
			// 像数据库添加记录
			mFile := models.File{}
			mFile.AddFile(userInfo.ID, f.Filename, fileExt, filepath)
			succMap[f.Filename] = filepath[1:]
		}
	}

	apiReturn.SuccessData(c, gin.H{
		"succMap":  succMap,
		"errFiles": errFiles,
	})
}

func (a *FileApi) GetList(c *gin.Context) {
	list := []models.File{}
	userInfo, _ := base.GetCurrentUserInfo(c)
	if err := global.Db.Order("created_at desc").Find(&list, "user_id=?", userInfo.ID).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	data := []map[string]interface{}{}
	sourcePath := global.Config.GetValueString("base", "source_path")
	for _, v := range list {
		diskPath, publicURL, err := resolveStoredFileSrc(sourcePath, v.Src)
		if err != nil {
			global.Logger.Warnf("skip invalid file src for id %d: %v", v.ID, err)
			continue
		}
		data = append(data, map[string]interface{}{
			"src":        publicURL,
			"fileName":   v.FileName,
			"id":         v.ID,
			"createTime": v.CreatedAt,
			"updateTime": v.UpdatedAt,
			"path":       diskPath,
		})
	}
	apiReturn.SuccessListData(c, data, int64(len(data)))
}

func (a *FileApi) Deletes(c *gin.Context) {
	req := commonApiStructs.RequestDeleteIds[uint]{}
	userInfo, _ := base.GetCurrentUserInfo(c)
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	sourcePath := global.Config.GetValueString("base", "source_path")
	if err := global.Db.Transaction(func(tx *gorm.DB) error {
		files := []models.File{}

		if err := tx.Order("created_at desc").Find(&files, "user_id=? AND id in ?", userInfo.ID, req.Ids).Error; err != nil {
			return err
		}

		paths := make([]string, 0, len(files))
		for _, v := range files {
			diskPath, _, err := resolveStoredFileSrc(sourcePath, v.Src)
			if err != nil {
				return err
			}
			paths = append(paths, diskPath)
		}
		for _, diskPath := range paths {
			if err := os.Remove(diskPath); err != nil && !os.IsNotExist(err) {
				return err
			}
		}

		if err := tx.Order("created_at desc").Delete(&files, "user_id=? AND id in ?", userInfo.ID, req.Ids).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	apiReturn.Success(c)

}
