package router

import (
	"os"
	"path/filepath"
	"strings"

	"x-panel/global"
	"x-panel/lib/wallpaper"
	// "x-panel/router/admin"
	"x-panel/router/openness"
	"x-panel/router/panel"
	"x-panel/router/system"

	"github.com/gin-gonic/gin"
)

// 初始化总路由
func InitRouters(addr string) error {
	router := gin.Default()
	rootRouter := router.Group("/")
	routerGroup := rootRouter.Group("api")

	// 接口
	system.Init(routerGroup)
	panel.Init(routerGroup)
	openness.Init(routerGroup)

	// WEB文件服务
	webPath := resolveWebPath()
	{
		router.StaticFile("/", filepath.Join(webPath, "index.html"))
		router.Static("/assets", filepath.Join(webPath, "assets"))
		router.Static("/custom", filepath.Join(webPath, "custom"))
	}

	// 上传的文件
	sourcePath := global.Config.GetValueString("base", "source_path")
	uploadURLPath := sourcePath[1:]
	registerUploadStaticRoutes(router, uploadURLPath, sourcePath)
	wallpaper.Start(sourcePath)

	files, err := webRootFiles(webPath)
	if err != nil {
		global.Logger.Warn("Cannot read web root files: ", err)
	} else {
		reserved := make(map[string]bool)
		for _, route := range router.Routes() {
			name, _, _ := strings.Cut(strings.TrimPrefix(route.Path, "/"), "/")
			if name != "" {
				reserved[name] = true
			}
		}
		for _, name := range files {
			if reserved[name] {
				global.Logger.Warn("Skipping web root file with conflicting route: /", name)
				continue
			}
			router.StaticFile("/"+name, filepath.Join(webPath, name))
		}
	}

	global.Logger.Info("X-Panel is Started.  Listening and serving HTTP on ", addr)
	return router.Run(addr)
}

func webRootFiles(webPath string) ([]string, error) {
	entries, err := os.ReadDir(webPath)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.Type().IsRegular() && name != "index.html" && !strings.HasPrefix(name, ".") {
			files = append(files, name)
		}
	}
	return files, nil
}

func resolveWebPath() string {
	executable, err := os.Executable()
	if err != nil {
		return "./web"
	}
	return webPathForExecutable(executable)
}

func webPathForExecutable(executable string) string {
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "./web"
	}
	webPath := filepath.Join(filepath.Dir(resolved), "web")
	info, err := os.Stat(webPath)
	if err != nil || !info.IsDir() {
		return "./web"
	}
	return webPath
}

func registerUploadStaticRoutes(router *gin.Engine, uploadURLPath, sourcePath string) {
	serve := func(route string) {
		group := router.Group(route, func(c *gin.Context) {
			c.Header("X-Content-Type-Options", "nosniff")
			c.Header("Content-Security-Policy", "default-src 'none'; img-src 'self' data:; style-src 'unsafe-inline'; sandbox")
			c.Next()
		})
		group.Static("/", sourcePath)
	}
	serve(uploadURLPath)
	if uploadURLPath != "/uploads" {
		serve("/uploads")
	}
}
