# X-Panel Linear 浅色视觉体系（2026-11-09）

## 来源与边界

本规范以已定稿的 `xnav.css`、`Main.dc.html`、`Mobile.dc.html`、`Tokens.dc.html` 为设计依据。`src/styles/global.less` 中的 `--x-*` 是运行时变量唯一来源，Naive UI 和 Tailwind 均引用同一套变量。浅色为默认主题；`html.dark` 与 `[data-theme="dark"]` 使用同一组暗色值。保留现有产品行为和 `panelConfig` 数据结构。2026-11-10 soft-tones 调整后，以下运行时变量值与定稿 `xnav.css` 不同。

用户在 2026-11-09 明确批准本项目采用 32px 卡片圆角，作为外脑 `DESIGN.md` 的“圆角 ≤8px”硬约束的项目例外。页面不使用渐变、霓虹、发光、玻璃拟态、emoji 图标，也不以单边框装饰圆角卡片。

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
| `--x-r-card` | `32px` | `—` |
| `--x-r-small` | `24px` | `—` |
| `--x-r-pill` | `999px` | `—` |
| `--x-r-icon` | `14px` | `—` |
| `--x-r-glyph` | `10px` | `—` |
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

分组色相依次循环 blue、indigo、purple、pink、orange、mint、brown、teal、green、red；每色包含 `-fg`、`-tint`、`-sys`。“全部”固定使用 Apple Blue，不占色相。配色只依赖分组排序，不存入后端。

## 分层、尺寸和字体

L0 页面使用 `bg0`，L1 大卡使用 `bg1`，L2 小卡和 L1 上的控件使用 `bg2`。相邻两层不使用同色。卡片不使用边框、阴影或渐变；点击控件统一高 44px。卡片圆角 32px，小卡 24px，胶囊 999px，44px 图标块 14px，32px 图标块 10px。区块间距 40–56px，细部按 8px 基准，以 4px、12px 辅助。字体使用本地 `@fontsource-variable/onest`，中文回退苹方等系统字体；数字使用表格式数字。

按钮和链接的 `focus-visible` 为 3px 主色 35% 环并外扩 2px；输入框外层 `focus-within` 为 2px 主色及 5px 光晕。减少动态效果设置下关闭旋转和位移动画，状态文字仍显示。

## 首页与分组侧栏

桌面宽度从 768px 起：左侧 88px 浮动 sticky 侧栏，外距 16px，内距 16px，圆角 32px、`bg1`。品牌块 44×44，双色 X，点击回“全部”并回顶。各分组按钮 44×44、圆角 14px，未选中为 `tint` 底 + `fg` 图标，悬停为 `bg2`，选中为 `fg` 实色底 + `bg0` 图标；“全部”选中用主色实底 + 白图标。组名和数量不常驻侧栏。Tooltip 挂到 body，首次悬停延迟 300ms，相邻按钮立即切换，离开栏 100ms 后关闭；使用反色胶囊、6px 色点和表格式数量。键盘使用 roving tabindex，方向键、Home/End 移焦，Enter/Space 原生激活。

小于 768px 时使用底部 64px 浮动胶囊栏，左右 16px、距底 24px，背景 `bg1`。最多显示“全部”、当前分组、最多两个相邻分组、“更多”；当前项展开显示图标和标题。“更多”打开底部抽屉列出全部分组。底部另铺 112px `bg0` 底带，内容预留 136px；悬浮按钮和回顶控件置于导航之上。

顶栏为 34px 品牌名、日期与数量次行、44px 筛选框、整理与导入书签按钮。概览卡 `bg1`、32px 圆角、32px 内边距、12 栏网格；时钟、搜索、系统状态和壁纸缩略图按开关自然收缩。分组网格最小宽 460px、间距 40px。分组卡 `bg1`、32px 圆角、24px 内边距；网址小卡 `bg2`、24px 圆角、高 64px。总览每组显示前 8 个网址和“查看全部”；单组显示全部。无 favicon 时显示所在分组 `tint` 底 + `globe` 图标。壁纸开启时标题用白字，卡片和侧栏保持实色。

## 异步操作三态

| 操作 | Loading | Success | Error |
| --- | --- | --- | --- |
| 保存整理排序 | 卡右上角“保存中” | “已保存”保留 1.5 秒 | 恢复远端顺序并提示失败 |
| 删除网址或分组 | 确认按钮 loading、阻止重复提交 | 列表刷新；网址删除可提供撤销窗口 | 弹窗内显示原因，可重试 |
| 编辑分组或网址 | 保存按钮 loading、表单保持可见 | 关闭弹窗并刷新 | 字段旁显示原因，保留输入 |
| 登录 | 主按钮 loading，防重复提交 | 导航到首页 | 对应字段下红字和图标 |
| 系统状态 | 骨架或“—”占位 | 状态标签与数值 | 红色“无法获取”、数值“—”，30 秒后重试 |
| 壁纸缩略图 | 纯色占位图 | 显示真实图 | 纯色占位图与尺寸 |

占位图统一使用 `--x-ph` 纯色底、居中说明和尺寸标注（例如“今日壁纸 / 320 × 180”），不使用模糊图、渐变或假数据。
