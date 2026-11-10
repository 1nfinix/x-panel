# X-Panel Linear 浅色视觉体系（2026-11-09）

## 来源与边界

本规范以已定稿的 `xnav.css`、`Main.dc.html`、`Mobile.dc.html`、`Tokens.dc.html` 为设计依据。`src/styles/global.less` 中的 `--x-*` 是运行时变量唯一来源，Naive UI 和 Tailwind 均引用同一套变量。浅色为默认主题；`html.dark` 与 `[data-theme="dark"]` 使用同一组暗色值。2026-11-10 soft-tones 与首页索引化调整后，以下运行时变量值与定稿 `xnav.css` 不同。旧版搜索栏配置字段可留在已保存 JSON 中，前端不再使用。

用户在 2026-11-09 批准原方案的 32px 卡片圆角；2026-11-10 调整为 20px，仍是外脑 `DESIGN.md` 的“圆角 ≤8px”硬约束的项目例外。页面不使用渐变、霓虹、发光、玻璃拟态、emoji 图标，也不以单边框装饰圆角卡片。

## 变量表

以下数值按当前 `global.less` 更新；空白暗色列表示该尺寸或字体继承浅色值。

| 变量 | 浅色 | 深色 |
| --- | --- | --- |
| `--x-bg0` | `#F1F3F6` | `#1B1C1F` |
| `--x-bg1` | `#F9FAFB` | `#242529` |
| `--x-bg2` | `#FFFFFF` | `#2F3035` |
| `--x-hover` | `#EBEEF2` | `#3A3B41` |
| `--x-ph` | `#E3E7EC` | `#3A3B41` |
| `--x-ph-text` | `#3A3A3C` | `#EBEBF0` |
| `--x-ph-sub` | `#5B5B60` | `#AEAEB2` |
| `--x-text1` | `#1D1D1F` | `#F5F5F7` |
| `--x-text2` | `#636368` | `#ADADB2` |
| `--x-icon3` | `#8E8E93` | `#8E8E93` |
| `--x-sep` | `#E3E7EC` | `#3A3B41` |
| `--x-disabled-bg` | `#E5E8EC` | `#3A3A3C` |
| `--x-disabled-fg` | `#8E8E93` | `#8E8E93` |
| `--x-primary` | `#0071E3` | `#0071E3` |
| `--x-primary-hover` | `#0062C4` | `#0062C4` |
| `--x-primary-active` | `#00539F` | `#00539F` |
| `--x-on-primary` | `#FFFFFF` | `#FFFFFF` |
| `--x-link` | `#0062C4` | `#6CB4FF` |
| `--x-focus` | `rgba(0, 113, 227, .35)` | `rgba(10, 132, 255, .55)` |
| `--x-inverse-bg` | `#1D1D1F` | `#F5F5F7` |
| `--x-inverse-fg` | `#FFFFFF` | `#1D1D1F` |
| `--x-inverse-sub` | `#A1A1A6` | `#6E6E73` |
| `--x-inverse-btn` | `#3A3A3C` | `#E5E5EA` |
| `--x-danger-hover` | `#F9D5D9` | `#552B2A` |
| `--x-blue-fg` | `#0062C4` | `#6CB4FF` |
| `--x-blue-tint` | `#E6F1FC` | `#18314B` |
| `--x-blue-sys` | `#007AFF` | `#0A84FF` |
| `--x-green-fg` | `#1B7330` | `#4CD964` |
| `--x-green-tint` | `#E8F4EB` | `#20402A` |
| `--x-green-sys` | `#34C759` | `#30D158` |
| `--x-orange-fg` | `#B83000` | `#FFB340` |
| `--x-orange-tint` | `#FBEBE5` | `#49361A` |
| `--x-orange-sys` | `#FF9500` | `#FF9F0A` |
| `--x-red-fg` | `#C40012` | `#FF8A84` |
| `--x-red-tint` | `#FCE6E8` | `#492424` |
| `--x-red-sys` | `#FF3B30` | `#FF453A` |
| `--x-indigo-fg` | `#3634A3` | `#A5A3FF` |
| `--x-indigo-tint` | `#ECECF8` | `#292946` |
| `--x-indigo-sys` | `#5856D6` | `#5E5CE6` |
| `--x-purple-fg` | `#8944AB` | `#DA8FFF` |
| `--x-purple-tint` | `#F4ECF7` | `#3D2848` |
| `--x-purple-sys` | `#AF52DE` | `#BF5AF2` |
| `--x-pink-fg` | `#C4103F` | `#FF8FA6` |
| `--x-pink-tint` | `#FCE8EE` | `#49212B` |
| `--x-pink-sys` | `#FF2D55` | `#FF375F` |
| `--x-teal-fg` | `#0070A0` | `#70D7EA` |
| `--x-teal-tint` | `#E5F1F6` | `#233E45` |
| `--x-teal-sys` | `#30B0C7` | `#40C8E0` |
| `--x-mint-fg` | `#0C7A74` | `#7EEBE6` |
| `--x-mint-tint` | `#E4F3F2` | `#2A4445` |
| `--x-mint-sys` | `#00C7BE` | `#63E6E2` |
| `--x-brown-fg` | `#7F6545` | `#CBAE88` |
| `--x-brown-tint` | `#F3EFEA` | `#39332D` |
| `--x-brown-sys` | `#A2845E` | `#AC8E68` |
| `--x-r-card` | `20px` | `—` |
| `--x-r-small` | `12px` | `—` |
| `--x-r-pill` | `999px` | `—` |
| `--x-r-icon` | `12px` | `—` |
| `--x-r-glyph` | `8px` | `—` |
| `--x-s1` | `8px` | `—` |
| `--x-s2` | `16px` | `—` |
| `--x-s3` | `24px` | `—` |
| `--x-s4` | `32px` | `—` |
| `--x-s5` | `40px` | `—` |
| `--x-s6` | `48px` | `—` |
| `--x-s7` | `56px` | `—` |
| `--x-control` | `44px` | `—` |
| `--x-font` | `Onest, -apple-system, "PingFang SC", "Helvetica Neue", "Microsoft YaHei", sans-serif` | `—` |
| `--x-ease` | `cubic-bezier(.2, .8, .2, 1)` | `—` |
| `--x-shadow-overlay` | `0 24px 60px -12px rgba(0,0,0,.18), 0 2px 8px rgba(0,0,0,.06)` | `0 24px 60px -12px rgba(0,0,0,.6), 0 2px 8px rgba(0,0,0,.4)` |

分组色相依次循环 blue、indigo、purple、pink、orange、mint、brown、teal、green、red；每色包含 `-fg`、`-tint`、`-sys`。“全部”固定使用 Apple Blue，不占色相。配色只依赖分组排序，不存入后端。

## 分层、尺寸和字体

L0 页面使用 `bg0`，分组与系统指标横条使用一层 `bg1`；总览网址行透明，单组网址行使用 `bg1`。输入框、选择器圆角 12px，按钮和标签保留胶囊；卡片无边框、无阴影，弹窗、对话框、菜单、Popover 和底部抽屉使用 `--x-shadow-overlay`。2026-11-10 首页索引化调整后，分组卡圆角 20px，小卡 12px，图标块 12px，行内字标 8px；字号以 12、13、15、17、22、28px 为主，时钟 44px。分组间距 24px，细部按 8px 基准。字体使用本地 `@fontsource-variable/onest`，中文回退苹方等系统字体；数字使用表格式数字。

按钮和链接的 `focus-visible` 为 3px 主色 35% 环并外扩 2px；输入框外层 `focus-within` 为 2px 主色及 5px 光晕。减少动态效果设置下关闭旋转和位移动画，状态文字仍显示。

## 首页与分组侧栏

桌面宽度从 768px 起：左侧 72px sticky 侧栏，外距 16px，圆角 20px、`bg1`。未选中分组图标为 `text2`，背景透明；悬停为 `hover` 底和 `text1` 图标；当前分组才显示 `tint` 底和 `fg` 图标。“全部”当前项使用蓝色 tint/fg。组名和数量不常驻侧栏。Tooltip 挂到 body；键盘使用 roving tabindex，方向键、Home/End 移焦，Enter/Space 原生激活。

小于 768px 时使用底部 64px 浮动栏，未选中项透明且为 `text2`，当前项才显示 tint/fg；“更多”打开底部抽屉列出全部分组。底部另铺 112px `bg0` 底带，内容预留 136px；网络、设置、登录按钮在标题行右侧并随内容滚动，回顶控件仍在导航之上。

顶栏为 28px 品牌名、日期与数量次行、44px 胶囊筛选框和次要操作按钮；整理模式的“完成”使用主色。概览区不铺外层卡，仅含时钟和系统指标：宽度至少 768px 时一行，时钟在左、指标在右，垂直居中并两端对齐；小于 768px 时各占一行，系统指标内部排两列。任一项关闭时另一项占满，全部关闭时不渲染概览区。顶栏至概览区 24px，概览区至分组区 40px；无概览区时顶栏至分组区 40px。系统状态为一层 `bg1` 横条。分组网格最小列宽 420px、间距 24px；分组卡一层 `bg1`、20px 圆角、内边距 20px/12px。总览网址行透明、高 52px，悬停和聚焦时显示 `bg2`；单组网址行为 `bg1`、高 56px。总览每组超过 9 个网址才折叠为前 8 个并显示“查看全部”。无 favicon 时显示首字符字标。壁纸开启时标题用白字，卡片和侧栏保持实色；壁纸背景及相关设置保留，概览区不显示壁纸缩略图。

## 异步操作三态

| 操作 | Loading | Success | Error |
| --- | --- | --- | --- |
| 保存整理排序 | 卡右上角“保存中” | “已保存”保留 1.5 秒 | 恢复远端顺序并提示失败 |
| 删除网址或分组 | 确认按钮 loading、阻止重复提交 | 列表刷新；网址删除可提供撤销窗口 | 弹窗内显示原因，可重试 |
| 编辑分组或网址 | 保存按钮 loading、表单保持可见 | 关闭弹窗并刷新 | 字段旁显示原因，保留输入 |
| 登录 | 主按钮 loading，防重复提交 | 导航到首页 | 对应字段下红字和图标 |
| 系统状态 | 骨架或“—”占位 | 状态标签与数值 | 红色“无法获取”、数值“—”，30 秒后重试 |
