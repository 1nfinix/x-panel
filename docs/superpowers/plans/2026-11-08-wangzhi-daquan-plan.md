# 网址大全改造实施计划

- 日期：2026-11-08（北京时间）
- 设计：[2026-11-08-wangzhi-daquan-design.md](../specs/2026-11-08-wangzhi-daquan-design.md)（已批准，提交 `cf530e6`）
- 分支：集成分支 `wangzhi-daquan`（从 `master` 拉出）；代码分支 `feat/wangzhi-daquan`，放在独立的 worktree 里
- 分工：Codex 写代码（遵循 `codex-dispatch` 技能）；Claude 写派活说明、审查、亲自重跑验收；本地截图验收和登录验收由用户做；上线的每一步由用户单独授权

## 0. 开工前要你决定的事

| # | 事项 | 建议 | 原因 |
|---|---|---|---|
| D1 | 本机安装 Go（`brew install go`，1.21 或以上） | 装 | 后端有 CGO sqlite。没有 Go 的话，Codex 写完后端代码在本地编译不了、也测不了，只能把每次验证都交给 Docker 来跑，Codex 的沙箱大概率连不上 Docker。属于安装操作，需要你授权 |
| D2 | 启动 Docker Desktop | 需要 | 产出 linux/amd64 二进制、做回滚兼容性实测都要用。Docker 现在没在运行（`docker.sock` 不存在） |

Node 和 pnpm 的风险：本机是 Node 24 和 pnpm 12，lock 文件是 `lockfileVersion: '6.0'`，对应 pnpm 8。为避免 lock 被改写，统一用 `npx pnpm@8` 安装依赖。如果 Vite 4 或 `vue-tsc` 1.x 在 Node 24 下报错，记录下来再决定，不擅自升级框架（设计第 3 节规定）。

## 1. 阶段与验收

### 阶段 A：数据快照（Claude 做，只读）

1. 在服务器上执行 `sqlite3 -readonly … .dump`，通过 ssh 流式传回本地，在临时目录里还原成 `database.db`。不在服务器上写任何文件。
2. 用 `rsync` 把 `conf/uploads/` 拉回本地，再拉 `conf.ini`，**只拉路径相关的配置项**，用本地生成的配置代替原文件。
3. 把 `/opt/sun-panel/sun-panel`（1.8.1）和 `web/` 拉回本地，留给回滚实测用。
4. 存放位置：会话临时目录 `…/scratchpad/prod-snapshot-20261108/`，不进仓库。
5. 核对：网址 64、分组 9、文件 98，并执行 `PRAGMA integrity_check` 返回 ok。

### 阶段 B：构建链路和基线（Codex 第 1 轮）

范围：

- `service/router/A_ENTER.go`：`web/` 目录改为相对于可执行文件所在目录解析，找不到时回退到 `./web`。
- `service/assets/version` 改为 `11|1.3.0-x.1`。
- 新增 `scripts/build-linux-amd64.sh` 和配套的 `Dockerfile.build`。
  - 构建环境：`golang:1.21-alpine3.18` 加 musl，静态链接，`--platform linux/amd64`。
  - 前端用 `npx pnpm@8` 构建。
  - 产出 `release/sun-panel_v1.3.0-x.1_linux_amd64.tar.gz`，内含 `sun-panel` 和 `web/`。
- `.gitignore` 加上 `release/`。

验收（Claude 重跑）：

- 本机 `go build` 和 `go vet ./...` 通过（如果 D1 已装 Go）。
- 构建脚本能产出 tar 包，`file sun-panel` 显示为 `ELF 64-bit x86-64, statically linked`。
- 在 Docker linux/amd64 中，用阶段 A 的数据副本启动改造版（这时界面还没改）：
  - 首页返回 200；
  - 接口返回 64 个网址；
  - 再查一次库，网址条数仍是 64，并且 `file` 表多出了 `method` 列（这是预期结果）。
- **用户登录**：用户在本地副本上用线上原账号登录一次。登不上就执行 `-password-reset`，并把这一步写进上线步骤。
- **回滚实测**：用 1.8.1 二进制打开“被 1.3.0 迁移过的库”的副本，看它能否正常启动并列出网址。结果写进本计划第 3 节。

### 阶段 C：后端功能（Codex 第 2 轮）

范围仅限 `service/`：

1. **登录验证码**（设计 8.4）
   - 注册免登录的出图接口 `/api/captcha/getImage`。
   - 登录接口在 `LoginCaptcha` 开启时先校验验证码；新增配置项默认开启。
   - 验证码一次性有效，有效期不超过 5 分钟；验证码错误要返回独立的错误码。
2. **favicon**（设计 8.1）
   - `getSiteFavicon` 取尺寸最大的图标；先找 `apple-touch-icon`，再找 `icon`，最后用 `/favicon.ico`。
   - 单个网站超时不超过 10 秒，大小上限 1MB。
   - 新增“补齐候选列表”接口：返回类型 1，以及类型 2 但 src 为空的网址；不包含类型 3。
   - 新增“按 id 批量补齐”接口，返回逐条的成功或失败。
   - 两个接口都只作用于当前登录用户。
3. **壁纸池**（设计 8.3）
   - 后台任务启动时跑一次，之后每 24 小时跑一次，拉取必应最近 8 张壁纸，存到 `source_path/wallpaper/`。
   - 去重；只保留最近约 30 张，超出的删除。
   - 网络失败只记录日志，不影响服务启动。
   - 新增一个免登录接口，返回壁纸池里的图片地址列表。
4. 补 Go 单元测试：
   - favicon HTML 解析：从多个 `link` 中选出最大的那个；
   - 壁纸池去重和数量上限；
   - 验证码一次性校验。
   - 网络请求用 `httptest` 模拟，不访问外网。

验收（Claude 重跑）：

- `go test ./...` 和 `go vet ./...` 全部通过。
- 本地用数据副本启动，用 curl 依次验证：
  - 不带验证码登录返回验证码错误；
  - 带错误的验证码登录返回验证码错误；
  - 带正确验证码登录成功，可以在测试模式下从内存 store 取到答案，或者由用户手动输入；
  - 同一个验证码第二次使用失败；
  - 补齐候选只有 6551 和 Inspora 两个；
  - 壁纸接口能返回列表；断开网络后服务照常启动。

### 阶段 D：前端改版（Codex 第 3 轮）

范围仅限 `src/`、`public/`、`index.html`，可以新增依赖 `@goopui/core`（设计已经批准）：

1. **主题**
   - 深色基调，主色用 goopUI 的绿色系，写成 CSS 变量。
   - 通过 Naive UI 的 `themeOverrides` 统一后台风格。
   - 卡片和图标圆角不超过 8px，标签做成胶囊形。
   - 标题用粗圆字体，选系统字体或自托管字体，不走外网 CDN。
2. **首页 E3 布局**
   - 顶部从左到右：logo 文字、“全部”、各分组标签。
   - 默认显示第一个分组。
   - 网格使用约 46px 的大图标；“全部”页按分组分段显示，带小标题。
   - 每个分组标签末尾有“添加”按钮，单个分组内可以拖拽排序；“全部”页只读。
   - 右键菜单保留。
   - 手机上标签可以横向滑动，每行 4 到 5 个图标。
   - 时钟和搜索栏默认关闭，代码保留。
3. **黏液效果**
   - 应用在三处：标签滑块、图标悬停和按下、弹窗展开。
   - 优先用 `@goopui/core`；接不上就用自己写的 SVG 滤镜。
   - 用户系统开启 `prefers-reduced-motion` 时关闭动效。
   - 不做背景漂浮球。
4. **壁纸**
   - 每次打开页面，从壁纸池里随机选一张；加载完之前显示纯深色底，加载完再淡入。
   - 壁纸池为空时，回退到用户上传的壁纸。
   - 面板设置里加“自动壁纸”开关，默认开启。
   - 登录页也使用同一个壁纸池。
5. **favicon**
   - 编辑弹窗里，网址输入框失焦时自动抓取图标，仅在类型为 2 且 src 为空时触发。
   - 设置里加“一键补齐”，流程为：列出候选 → 用户勾选 → 执行 → 显示结果。
6. **站点图标**
   - 新的 `public/favicon.svg`（一滴黏液，里面带一个“X”），配套生成 `favicon.ico`。
   - 更新 `index.html` 的引用。
7. **登录页**
   - 启用验证码组件，深色样式。
   - 点击图片刷新验证码；登录失败后自动刷新。

验收（Claude 重跑）：

- `npx pnpm@8 run type-check`、`lint`、`build` 全部通过。
- 本地用数据副本启动后，Claude 在内置浏览器里逐项检查设计第 2 节和 8.1 到 8.4 节：
  - 截图：桌面 1440 宽、手机 375 宽、登录页；
  - 打开系统“减少动态效果”，确认动效关闭；
  - 64 个网址都能找到。
- 截图交给用户确认界面。

### 阶段 E：合并与打包（Claude 做，需要你授权）

1. 审查通过后，把 `feat/wangzhi-daquan` 合并进 `wangzhi-daquan`。**合并前会先问你。**
2. 重新跑一遍阶段 B 的构建脚本，产出最终的 tar 包，记下 sha256。
3. 是否推送到 `1nfinix/sun-panel`、是否开 PR、是否合并回 `master`，每一项单独问你。

### 阶段 F：上线（每一步需要你单独授权；遵循 tencent-cvm-ops 托管约定）

| 步 | 动作 | 影响 | 恢复 |
|---|---|---|---|
| F1 | 上传 tar 包到 `/opt/sun-panel/releases/v1.3.0-x.1/`；把现有 1.8.1 复制（不是移动）到 `releases/v1.8.1/`；建立 `current → releases/v1.8.1` | 不影响运行中的服务 | 删除新增的目录 |
| F2 | 备份 unit 文件：`/etc/systemd/system/sun-panel.service.bak-20261108-x1`；把 `ExecStart` 改为 `/opt/sun-panel/current/sun-panel`；执行 `daemon-reload` 并重启（此时 current 仍指向 1.8.1） | 服务重启约 5 秒 | 用备份覆盖 unit 文件，执行 `daemon-reload` 并重启 |
| F3 | 停止服务；`sqlite3 database.db ".backup database.db.bak-20261108-pre-x1"`；执行 `PRAGMA integrity_check`；备份 `conf.ini.bak-20261108-pre-x1` | 服务停止 | 启动服务 |
| F4 | 把 `current` 切到 `v1.3.0-x.1`；按阶段 B 的结论决定是否执行 `-password-reset`；启动服务 | 正式切换到新版 | 见 F6 |
| F5 | 验证：执行 `systemctl show -p Result -p ExecMainStatus -p ActiveState`；查看 `journalctl -u sun-panel --since -10min`；`https://119helpme.icu` 返回 200；网址数为 64；壁纸池已拉到图片；用户实际登录一次 | — | — |
| F6 | 回滚：停止服务，`current` 切回 `v1.8.1`；如果阶段 B 实测 1.8.1 不能读取迁移后的库，同时恢复 F3 的数据库备份；然后启动服务，按 F5 验证 | 回到 1.8.1 | — |

旧的 `/opt/sun-panel/sun-panel` 和 `/opt/sun-panel/web` 在 F2 之后不会再被使用。观察一段时间、没有问题后，**另外征求你同意**再删除。

上线完成后，更新 tencent-cvm-ops 技能里的 `references/server-inventory.md`：sun-panel 一行改为新版本和新的 `ExecStart` 路径。按技能规定，这类长期事实写在技能里。

## 2. Codex 派活要点

- worktree：在 `wangzhi-daquan` 分支上执行 `git worktree add -b feat/wangzhi-daquan ../sun-panel-wt-x wangzhi-daquan`。
- 阶段 B、C、D 在同一个 worktree 里按顺序执行，用 `resume` 续跑同一个会话。每轮结束后 Claude 审查，有问题就续跑修改，审查通过才开始下一轮。
- 每轮的派活说明要写明：先读设计文档和本计划对应的阶段；只改范围内的文件；只提交到当前分支，不推送；不改生产环境，不连接服务器；数据副本的路径作为参数传入，只读使用。
- Codex 额度用完或不可用时，先问你是等待还是另作安排，不改由 Claude 自己写代码。

## 3. 记录（实施过程中填写）

- 阶段 A 快照核对结果：2026-11-08 完成，integrity_check ok；网址 64、分组 9、文件 98；uploads 33M；1.8.1 二进制已拉回（scratchpad/prod-snapshot-20261108）。
- 阶段 B 登录结果，以及是否需要 `-password-reset`：2026-11-08 用户用线上原账号在本地副本登录成功，**不需要** `-password-reset`。基线首页 200、网址 64，迁移只新增 `file.method` 列。
- 阶段 B 回滚实测（1.8.1 能否读取迁移后的库）：2026-11-08 通过。1.8.1 在被 1.3.0-x.1 迁移过的库副本上正常启动、无报错、integrity_check ok，用户登录确认网址都在。回滚只切软链即可，不必恢复备份。例外：改造版上线后新上传的文件地址是 `/conf/uploads/...`，1.8.1 不提供这个路径，回滚后这些新图标会失效，需要手动处理。
- 阶段 D 实测发现的兼容问题：1.8.1 写入的上传地址是 `/uploads/...`（file.src 为 `/2026/...`），1.3.0 只提供 `/conf/uploads`。已派 Codex 增加 `/uploads` 路由，并让文件管理兼容两种格式。
- 最终 tar 包的 sha256：`73b94abfabf0e0504951d5a90cd7a564b9da42d29c7b9718a7d92487feb21e76`（`sun-panel_v1.3.0-x.1_linux_amd64.tar.gz`，构建自 `5b44215`，存放在仓库 `release/`，已被 gitignore）。2026-11-08 在 Ubuntu 24.04 amd64 容器中冒烟通过：首页、旧版 `/uploads` 图标、站点图标、验证码、壁纸池均正常，网址 64 条，integrity_check ok。
