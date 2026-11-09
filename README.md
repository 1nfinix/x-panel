# X-Panel

个人网址导航面板，基于 [Sun-Panel](https://github.com/hslr-s/sun-panel) v1.3.0（hslr-s，MIT 许可）改造。

## 与 Sun-Panel 1.3.0 的主要区别

- 首页改为分组列表布局：顶部分组标签带数量，`/` 快捷键过滤，按分组瀑布流展示
- 添加网址时自动抓取网站图标，已有网址可一键补齐
- 图标背景色生效，默认白色
- 登录图形验证码
- 每次打开页面随机切换一张 Bing 壁纸
- X 造型 favicon 与 PWA 图标
- 兼容 Sun-Panel 1.8.1 的数据库和上传文件路径（`/uploads/...`）
- 监听地址读取 `[base] http_address`，可只监听 `127.0.0.1`

## 构建

前端为 Vue 3 + Vite + Naive UI，后端为 Go + Gin + GORM（sqlite，需要 CGO）。

linux/amd64 发布包（需要 Docker buildx）：

```bash
scripts/build-linux-amd64.sh
```

产物为 `release/x-panel_v<版本>_linux_amd64.tar.gz`，内含 `x-panel` 二进制和 `web/` 目录。

## 许可

MIT。原项目版权归 Sun-Panel 作者所有，见 [LICENSE](./LICENSE)。
