# 网址大全改造设计（Sun-Panel 1.3.0 fork）

- 日期：2026-11-08（北京时间）
- 仓库：`1nfinix/sun-panel`（fork 自 `hslr-s/sun-panel`），本地 `~/Documents/code/sun-panel`
- 线上：`https://119helpme.icu`，tencent-cvm 上的 `sun-panel.service`，目前运行官方闭源版 1.8.1
- 状态：设计稿，待用户批准

## 1. 为什么做

用户每天用 119helpme.icu 作为自己的网址导航页（1 个用户、9 个分组、64 个网址），有两点不满意：

1. **图标要手动处理**：新加的网址要自己去找图标、上传。
2. **一屏显示的内容太少**：现在是 1200px 宽度居中、带描述的详情卡片，一屏大约只能看到 16 个网址。

官方 1.8.1 是闭源的，没法改。所以基于开源的最后一个版本 v1.3.0 自己维护一个分支，以后想怎么改都可以。用户已经明确接受由此带来的代价：失去 1.4 到 1.8 新增的功能，以后自己维护。

## 2. 目标

1. **Favicon**
   - 新增网址时自动抓取网站图标，保存到服务器。
   - 设置里提供“一键补齐”，给已有网址补上图标。
   - 网址大全自己的浏览器标签页图标改为新设计：一滴黏液，里面带一个“X”。
2. **界面改版（布局 E3）**
   - 顶部是分组导航：左边是 logo 文字，后面依次是“全部”和各个分组标签。
   - 点一个分组，下面只显示这个分组的网址（单页切换，不刷新，不滚动定位）。
   - 打开页面时**默认显示第一个分组**（按分组排序的第一个）。
   - 内容区是大图标网格：图标约 46px，下面是名字，不显示描述。电脑上一行大约 10 个。
   - “全部”标签按分组依次排列，每组前面有一个小标题。
   - 去掉时钟和搜索栏（用现有的 `clockShow`、`searchBoxShow` 开关关闭，相关代码保留）。
3. **整体换成 goopUI 风格，但不换组件库**
   - 参考：https://goop.timmcmasters.work/
   - 首页：深色底，一个主色，标题用粗圆字体。分组标签的选中块像液体一样滑动。网址图标悬停时有黏液鼓起或滴落的效果，按下时压扁再弹回。
   - 后台：设置面板、编辑弹窗、右键菜单仍然用 Naive UI，只换主题（深色、同一个主色、同样的圆角），让它和首页风格一致。
   - 系统开启“减少动态效果”（`prefers-reduced-motion`）时，关闭黏液动画，只保留静态样式。
4. **壁纸自动切换**
   - 来源是必应每日壁纸（公开接口，不需要 key）。
   - 服务器每天拉一次最近 8 天的壁纸，下载到 `uploads/`，组成一个壁纸池，最多保留约 30 张，超出时删掉最旧的。
   - 每次打开页面，从池子里随机挑一张。图片加载完之前先显示纯深色底，加载完淡入。
   - 壁纸上叠一层深色遮罩，现有的 `backgroundMaskNumber` 设置继续生效。
   - 设置里提供“自动壁纸”开关，关掉后用用户自己上传的壁纸（现在的行为）。
   - 壁纸池为空或拉取失败时，回退到用户上传的那张壁纸。

## 3. 非目标

- 不对接「工作台」，也不为它预留接口（用户还没想好，2026-11-08 决定）。
- 不升级 Vue、Vite、Tailwind，不引入 `@goopui/vue`、reka-ui、shadcn-vue。
- 不补回 1.4 到 1.8 的功能（Open API、按分组设置卡片样式、分组公开访问隐藏、更多卡片类型等）。
- 不改多用户、登录、MySQL 和 Redis 相关逻辑。
- 不做 arm、Windows 和 Docker 镜像，只产出 linux/amd64 二进制。

## 4. 反目标（即使能跑也算失败）

- 切换后丢了网址、分组、已上传的图标或背景图。
- 自动抓图标**覆盖**了用户手动设置的图标。
- 首页因为动效变卡，或者在 Safari 上明显掉帧。
- 外部壁纸源不可用时出现白屏或报错。
- 后台设置页因为换主题导致文字看不清（对比度不足）。

## 5. 约束

### 5.1 技术现状（已核对源码）

- 前端：Vue 3.2.47、Vite 4.2、Naive UI 2.36、Pinia、Tailwind 3.2.7，代码在 `src/`，执行 `pnpm build` 产出 `dist/`。
- 后端：Go 1.20、Gin、GORM，代码在 `service/`，默认用 sqlite（`mattn/go-sqlite3`，需要 CGO）。
  - `service/assets/` 由 `go-bindata-assetfs` 打包成 `assets/bindata.go`。
  - 编译时要带 `-ldflags "-X sun-panel/global.RUNCODE=release"`。
- 已有能力：
  - 单个网址抓图标：`POST /panel/itemIcon/getSiteFavicon`，实现在 `service/lib/siteFavicon/favico.go`，会下载到服务器。
  - 编辑弹窗里有“获取图标”按钮。
- 前端目录在 `service/router/A_ENTER.go:26` 中写死为 `./web`，按**工作目录**解析。

### 5.2 线上现状（2026-11-08 只读核对）

- systemd 配置：`WorkingDirectory=/var/lib/sun-panel`、`ExecStart=/opt/sun-panel/sun-panel`，使用 `DynamicUser`、`StateDirectory=sun-panel`、`ProtectSystem=strict`。
- `conf.ini` 中的路径：
  - `source_path=./conf/uploads`
  - sqlite 文件为 `./conf/database/database.db`
  - 缓存和队列都用 memory
  - 这些路径和 1.3.0 的约定一致。
- 数据量：
  - `uploads/` 共 32M。
  - 64 个网址中，62 个的 `icon_json.itemType=2`（图片图标）。
  - 没有图片图标的 2 个：`https://6551.io/mcp`、`https://www.inspora.design/`。
- 1.8.1 比 1.3.0 多出的表和列：
  - `item_icon`：`card_type`、`only_name`、`background_color`、`expand_param`
  - `item_icon_group`：`only_name`、`public_visit_hide`、`card_style`
  - `file`：`type`、`is_public_gallery`
  - 整张 `open_api_token` 表，里面有 1 条记录
  - 1.3.0 的 `file.method` 在线上库中不存在。
  - GORM 的 `AutoMigrate` 只加不删，所以推断 1.3.0 能打开这个库，**需要用副本实测确认**。
- 这些多出的字段实际上没有被用到：9 个分组的 `card_style` 完全一样，`public_visit_hide` 全为 0，`card_type` 全为 1。
- 面板设置在 `user_config.panel_json`：
  - 背景图：`/uploads/2026/1/27/…jpg`
  - `logoText`：“X 导航”
  - `maxWidth`：1200px
  - `iconStyle`：1
  - `clockShow`：true
  - `searchBoxShow`：true

### 5.3 部署约定（tencent-cvm-ops「项目托管约定」）

- 改造过的第三方程序放在 `/opt/sun-panel/releases/<版本>/`（目录里是二进制和 `web/`），通过 `/opt/sun-panel/current` 软链切换。
- 数据目录 `/var/lib/sun-panel/` 不动。
- 替换前用 `sqlite3 .backup` 备份数据库，并执行 `PRAGMA integrity_check`。备份放在原目录旁边，文件名为 `*.bak-YYYYMMDD-<用途>`。
- 改 systemd、重启服务、部署、回滚，每一步都需要用户明确授权。

### 5.4 分工

- 代码由 Codex 按 `codex-dispatch` 技能在独立 worktree 里编写。
- Claude 负责审查，并亲自重跑关键检查。
- 提交、推送、部署需要用户另外授权。

## 6. 已定的“怎么做”

| 项 | 决定 | 谁定的 |
|---|---|---|
| 基线 | 在 1.3.0 上自己改 | 用户 |
| 布局 | E3：顶部分组标签页加“全部”，大图标，默认第一个分组 | 用户 |
| 时钟和搜索栏 | 去掉 | 用户 |
| 风格 | 整体参考 goopUI，不换组件库 | 用户 |
| 背景 | 自动切换的网络公开壁纸，每次打开随机一张 | 用户 |
| 壁纸来源 | 必应每日壁纸，服务器缓存成池 | Claude 定，用户未反对 |
| 站点图标 | 一滴黏液，里面带一个“X” | 用户 |
| 登录图形验证码 | 每次登录都要填（见 8.4） | 用户要求；“每次都要填”由 Claude 定 |
| 黏液效果实现 | 优先用 `@goopui/core`（纯 TypeScript，无依赖）；如果在 Vue 3.2 下接不上，就自己用 SVG 滤镜实现（模糊后做 alpha 阈值） | Claude 定 |
| 黏液使用范围 | 只用于交互反馈（标签滑块、图标悬停和按下、弹窗展开），背景不加漂浮球 | Claude 定，依据第 7 节 |
| 主色 | goopUI 的绿色系，定义成 CSS 变量，可以替换 | Claude 定 |
| 前端目录解析 | 改为相对于可执行文件所在目录查找 `web/`，找不到时回退到 `./web` | Claude 定 |
| 版本号 | `service/assets/version` 改为 `11\|1.3.0-x.1`，发布目录为 `releases/v1.3.0-x.1/` | Claude 定 |

## 7. 设计规范例外说明

用户全局设计规范 `vault/notes/prompts/tools/DESIGN.md` 中有两条硬约束：禁止装饰性光球和漂浮装饰物；卡片圆角不超过 8px。goopUI 的原始风格与这两条冲突。本设计的处理方式：

- 不做背景漂浮黏液球。背景用壁纸加深色遮罩。
- 图标和卡片圆角不超过 8px。分组标签用胶囊形（规范允许 pill）。
- 黏液只表达交互状态，不做常驻装饰。

## 8. 功能细节

### 8.1 自动抓取 favicon

- **图标类型**（1.3.0）：1 是文字，2 是图片，3 是 iconify 图标库。新建网址默认类型 2、`src` 为空（`src/views/home/components/EditItem/IconEditor.vue:30`）。
- **触发时机**：在新增或编辑弹窗里，网址输入框失去焦点后，**仅当图标类型为 2 且 `src` 为空**时，自动调用已有的 `getSiteFavicon`。类型 1、类型 3、已有地址的类型 2 一律不动（Claude 定）。
- **抓取结果**：
  - 成功：填好图片地址。
  - 失败：静默处理，保存时退回文字图标（首字母加纯色底），不弹错误，也不阻止保存。
- **抓取顺序**：
  1. 页面 HTML 中的 `apple-touch-icon` 和 `icon`，取尺寸最大的一个。
  2. `/favicon.ico`。
  - 下载大小上限沿用现有的 1MB。
  - 单个网址抓取超时不超过 10 秒。
- **一键补齐**：
  - 在设置中。候选为类型 1（文字）和类型 2 但 `src` 为空的网址，**不含类型 3**。
  - 先列出候选清单（默认全选，可取消勾选），用户确认后逐个抓取（Claude 定，避免误覆盖特意选的文字图标）。
  - 完成后报告成功和失败数量，并列出失败的网址。
  - 可以重复运行。
  - 线上当前候选为 2 个，均为类型 1 文字图标：6551、Inspora（2026-11-08 只读核对）。

### 8.1b 新布局下的编辑操作（Claude 定）

- 每个分组标签页的网格末尾保留“添加”按钮；编辑模式下可在**单个分组标签页内**拖拽排序。
- “全部”标签只读展示，不支持拖拽，也不支持跨分组拖拽。
- 分组的新增、改名、排序继续走设置里已有的“分组管理”；顶部标签顺序跟随分组排序。
- 右键菜单（编辑、删除、打开内网/外网地址）保留。

### 8.2 站点图标

- 新增 `favicon.svg`，并生成对应的 `favicon.ico`。
- 首页 `index.html` 的 icon 和 apple-touch-icon 指向新图标。
- 图案在深色和浅色浏览器标签栏上都要能看清。

### 8.3 壁纸池

- **后台任务**：服务启动时执行一次，之后每 24 小时执行一次，从必应接口获取最近 8 天的壁纸。
  - 接口：`https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=8`
- **存储**：下载到 `source_path` 下的独立子目录，用 urlbase 或哈希去重。最多保留约 30 张，超出时按时间删除最旧的。
- **接口**：提供一个接口返回壁纸池列表（或者直接随机返回一张），前端在每次加载首页时随机选一张。
- **开关**：自动壁纸开关是面板设置中的一项，存在 `panel_json` 里，默认开启。
- **访问控制**：壁纸池接口只返回图片地址，可以免登录访问，登录页也能用同一套壁纸。线上没有配置公开访问用户（`system_setting` 只有 `clientId` 一项，2026-11-08 核对），公开访问模式不在本次验收范围。
- **网络**：服务器访问必应的超时和失败都不能影响主服务启动。

### 8.4 登录图形验证码（用户 2026-11-08 追加）

- **要求**：登录页**每次都要填**图形验证码，4 位，字母和数字混合，不区分大小写；看不清可以点击图片换一张（Claude 定：总是要填，不做“失败几次后才出现”）。
- **现状**（已核对源码）：1.3.0 有一套没接通的验证码设施，需要补上的就是把它接通。
  - `service/lib/captcha/captcha.go` 里有生成和校验函数，基于 `base64Captcha`，答案存在内存里。
  - `service/api/api_v1/common/base/base.go` 里有 `VerificationCheck`。
  - 系统设置里有 `LoginCaptcha` 字段，`openness` 接口会把它返回给前端。
  - 前端 `src/components/common/Captcha` 组件已经写好，但在登录页里被注释掉了（`src/views/login/index.vue:98`）。
  - 后端**没有注册**出图接口 `/api/captcha/getImage`，登录接口 `Login` 也**没有校验**验证码。
- **行为**：
  - 每张验证码只能用一次，校验过就作废。有效期不超过 5 分钟。
  - 验证码错了，或者账号密码错了，都要换一张新图。
  - 校验顺序：先验验证码，再查账号密码。验证码错误的提示要和“账号或密码错误”分开。
  - 出图接口不需要登录。
- **开关**：沿用 `LoginCaptcha` 设置，默认开启。以后忘记密码时，可以在服务器上改配置临时关掉。
- **样式**：验证码图片跟随深色主题，背景深色、字浅色，保证能看清。

## 9. 构建与发布

- 前端：执行 `pnpm install && pnpm build`，产出 `dist/`。
- 后端：使用 Docker `--platform linux/amd64`，参考仓库中 `Dockerfile` 的 `golang:1.21-alpine3.18` 加 musl 方案。
  - 先执行 `go-bindata-assetfs`。
  - 再执行 `go build -ldflags "-X sun-panel/global.RUNCODE=release -linkmode external -extldflags -static"`，产出静态二进制。
  - 本机没有安装 Go，不需要装，全部在容器里编译。
- 打包：`sun-panel_v1.3.0-x.1_linux_amd64.tar.gz`，内含 `sun-panel` 和 `web/`。
- 本地验收：
  - 用线上数据库和 `uploads/` 的**副本**（放在本机临时目录，不进仓库）运行改造版。
  - 确认 64 个网址、9 个分组、背景图和图标都在。
  - 验证码：不填、填错、填对各登录一次，结果符合 8.4 节；刷新图片后旧验证码失效。
  - **用户用线上原账号密码在本地副本上登录一次**。1.3.0 的密码是三次 MD5（`service/lib/cmn/base.go:217`），线上密码字段是 32 位十六进制，格式一致，但 1.8.1 的算法没有验证。登不上时，用 1.3.0 自带的 `-password-reset` 参数重置第一个用户的密码，上线时同样处理。
  - **回滚兼容性实测**：把服务器上的 1.8.1 二进制拉到本地，在 Docker linux/amd64 中对“被 1.3.0 迁移过的库副本”启动一次，看能否正常读取。
  - 确认第 2 节的各项功能可用。
  - 截图给用户确认界面。

## 10. 上线与回滚（每一步都需要用户另外授权）

1. 在服务器上把现有的 1.8.1 移到 `/opt/sun-panel/releases/v1.8.1/`，上传新版本到 `/opt/sun-panel/releases/v1.3.0-x.1/`，建立 `current` 软链。
2. 修改 systemd：
   - `ExecStart=/opt/sun-panel/current/sun-panel`
   - 修改前备份 unit 文件。
3. 停止服务：
   - 备份数据库：`sqlite3 database.db ".backup database.db.bak-YYYYMMDD-pre-x1"`
   - 执行 `PRAGMA integrity_check`
   - 再把 `conf.ini` 也备份一份。
4. 把 `current` 切到新版本，启动服务。
5. 验证：
   - `systemctl show -p Result -p ExecMainStatus -p ActiveState`
   - 查看 `journalctl` 最近 10 分钟的日志
   - `https://119helpme.icu` 返回 200
   - 网址数量为 64
6. **回滚**：
   - 停止服务，把 `current` 切回 `v1.8.1`。
   - 数据库按第 9 节“回滚兼容性实测”的结果处理（Claude 定）：
     - 实测 1.8.1 能读迁移后的库：只切软链，保留上线后新增的数据。
     - 实测不能读，或没测：**恢复第 3 步的数据库备份**，上线后新增或修改的网址会丢失。数据只有你一个人在改，量也小，丢失部分手动补回即可。
   - 启动服务并按第 5 步验证。

## 11. 留给实施者决定

- 黏液效果的具体参数（模糊半径、阈值、弹簧系数）、动效时长。
- 图标网格在各种屏幕宽度下的列数和间距（手机上标签栏可以横向滑动，一行 4 到 5 个图标）。
- Naive UI 主题变量的具体取值。
- 粗圆标题字体的选择：优先用系统自带或可以自托管的字体，不依赖外网字体 CDN。
- 壁纸池接口的具体形式（返回列表还是随机一张）。

## 12. 第二轮变更（2026-11-08 用户确认）

1. **首页改为列表风**（参考 bueno.fyi，草图已确认）：
   - 顶部：左边 logo 文字，右边筛选框（按 `/` 聚焦，按名称或域名跨所有分组即时过滤，不是搜索引擎框）。
   - 分组标签：每个标签带数量，包括“全部”，默认显示第一个分组。
   - 单个分组：三栏紧凑列表，每行依次为 20px 图标、名称、域名、↗。
   - “全部”：每个分组一个半透明块，按瀑布流排成三栏，块标题旁显示数量。
   - 手机端单栏，筛选框换行显示。
   - 名称或域名过长时截断，鼠标悬停显示全名。
   - 编辑、单分组内拖拽排序、右键菜单、内外网切换、设置入口保留。
   - 取代第 2 节第 2 条中“大图标网格”的写法。
2. **改名**：界面上所有可见的“Sun-Panel”改为“X-Panel”，包括页面标题、登录页、关于页、PWA 名称、启动横幅。代码模块名、仓库名、服务名、LICENSE 不改。
3. **站点图标**：改成“X”造型，取代第 8.2 节的“黏液滴里带 X”。
4. **关于页**：维护者写用户本人，GitHub 账号 `1nfinix`，仓库 `1nfinix/sun-panel`。显示版本号。去掉原作者的 QQ 群、捐赠、B 站、YouTube、Gitee、Docker 链接。保留一行“基于 Sun-Panel（hslr-s，MIT 许可）”并附上游链接，这是 MIT 许可的署名要求。
5. 版本号改为 `12|1.3.0-x.2`。
