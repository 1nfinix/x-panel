# X-Panel 视觉改版规范（2026-11-09）

## 来源与边界

视觉方向参考 `/Users/vincent/.workbuddy-config-backups/20261031-shared-slim/skills/claude-design-card/DESIGN.md`，用于 X-Panel 的 Workbench 式布局、暖色表面和字体层级。产品始终称 X-Panel；界面与图标不使用 Anthropic 的放射星形标志、名称或其他品牌资产。

用户于 2026-11-09 知情选定此规范。与外脑 `DESIGN.md` 的“圆角不超过 8px”相比，本页允许卡片 12px 圆角；其余元素依旧优先使用紧凑、清晰的层级。品牌标题字距按后续首页概览带规范调整为 0。

## 颜色令牌

`src/styles/global.less` 是 CSS 令牌来源。`src/hooks/useTheme.ts` 的 Naive UI 对应色值必须同步维护。

| 令牌 | 浅色 | 深色 |
| --- | --- | --- |
| canvas | `#f5f3ee` | `#151412` |
| surface | `#ffffff` | `#1e1d1a` |
| surface-soft | `#eeeae3` | `#24231f` |
| surface-card | `#eae7e0` | `#292824` |
| surface-strong | `#e4dfd6` | `#33312c` |
| hairline | `#e2ded7` | `#2e2c28` |
| ink | `#141413` | `#faf9f5` |
| body | `#3d3d3a` | `#e3dfd7` |
| muted | `#6c6a64` | `#a09d96` |
| muted-soft | `#8e8b82` | `#8e8b82` |
| primary | `#cc785c` | `#cc785c` |
| primary-active | `#a9583e` | `#d98d73` |
| on-primary | `#ffffff` | `#ffffff` |
| error | `#c64545` | `#e06c6c` |
| success | `#5db872` | `#5db872` |
| focus-ring | `rgba(204,120,92,.15)` | `rgba(204,120,92,.28)` |

圆角分别为 6/8/12px 和 pill；正文使用系统回退的 Inter 字体栈，标题使用下文的自托管 Noto Serif SC Variable 字体栈，等宽字体使用 JetBrains Mono 回退栈。不请求第三方字体服务。

## 组件映射

- `group-rail`：40px 按钮包裹 32px 色块图标；颜色按分组顺序取色相，见下文。
- `text-input`：surface 底、hairline 边、40px 高、8px 圆角；焦点为 primary 边及 3px focus-ring。
- `secondary` 按钮：surface 底、hairline 边、ink 字。`primary` 按钮：primary 底、on-primary 字。
- `connector-tile`：surface 底、hairline 边、12px 圆角、16px 内边距；网址行悬停使用该分组浅色底的一半强度。
- 壁纸开关：默认关闭，首页和登录页显示纯色 canvas；打开后使用已有壁纸与暖黑遮罩，设置面板才显示壁纸参数。

## 删除的旧代码

删除未被调用的首页 Result、AppIcon，通用 HoverButton、JsonImportExport、Verification 组件，测试路由及页面，旧 Markdown/代码高亮样式与其专用依赖，以及未使用的 `vuedraggable`。仍使用的 `src/utils/jsonImportExport` 与 `vue-draggable-plus` 保留。

## 首页概览带与品牌字体（2026-11-09）

- 顶栏保留品牌、筛选和整理；概览带按时钟、搜索、系统状态顺序排列，与分组区同宽。三个模块全关时不渲染概览带；任意组合不保留空列。
- 宽度至少 1024px 时一行布局；640–1023px 时钟与系统状态在第一行、搜索独占第二行；不足 640px 时纵向堆叠。顶栏到概览带、概览带到分组区均为 24px，单分组标题到内容为 16px。
- 时钟使用 44px 表格式数字与 14px 日期，手机缩为 30px 并与日期同行。系统状态采用标签、百分比和 3px 进度条，超过 85% 用 error 色；编辑、排序与右键操作保留。分组总览采用约 320px 列宽的 CSS columns，分组卡避免跨列断开；独立 VueDraggable 列表和空组放置区保留。
- 品牌字与登录、About 标题使用自托管 Noto Serif SC Variable，字重 600，字距 0。字体按 unicode-range 分片加载。品牌桌面 28px，手机 24px。
- 网址行悬停 120ms 背景过渡，箭头位移 2px；筛选框支持方向键聚焦可见网址、Esc 清空并失焦。首屏卡片依次以最多 400ms 的淡入与 6px 上移动画出现；切换分组时仅做 150ms 淡入。减少动态效果设置时关闭这些动画。

## 分组侧栏（Icon Rail，2026-11-09）

- 桌面从 640px 起使用固定视口左侧的 64px 全高图标栏，canvas 背景及右侧 hairline；品牌图 28px、圆角 8px 固定顶部，点击回到“全部”并回顶，分组区独立纵向滚动。主内容在余下宽度按原 maxWidth 居中。壁纸下侧栏仍用实色 canvas。
- “全部”固定中性 layout-grid，不占色相序列。分组按排序显示，每个按钮 40×40px，间距 4px，内含 32×32px、圆角 8px 的色块及 18px 图标；默认浅色底与对应前景色，悬停增加 6% 前景色，选中为实色底加白色图标。左缘 3×20px 指示条使用同组前景色，以 transform 移动 180ms；状态过渡 150ms。减少动态效果时关闭过渡。
- Tooltip 在桌面按钮右侧 12px，ink 背景、canvas 字色、13px 字号、6px 圆角与 6×10px 内边距，无阴影；分组名前放 6px 同色圆点，数量使用淡化反色和表格式数字。首次悬停延迟 300ms，相邻按钮立即切换，离开栏 100ms 后关闭；浮层挂到 body 避免裁切。焦点进入时立即展示。按钮提供 `aria-label`、`aria-current`，↑/↓、Home/End 移动 roving 焦点，Enter/Space 原生激活。
- 小于 640px 时图标栏固定底部，高 56px 加安全区，横向滚动并隐藏滚动条；选中按钮自动滚入可见范围，指示条改为顶部 20×3px。触屏不显示 Tooltip，单分组视图在概览带下展示 18px 分组标题与数量；桌面同样展示。内容底部预留栏高，悬浮按钮和回顶按钮上移。
- 有效的本地 sprite 图标优先。空值、旧默认广告分组图标或不存在的图标按标题中最早出现的中英文关键词匹配；同位置按 `GroupGlyph/icons.ts` 规则顺序裁决，“图片”优先 image。仍未命中时，对规范化标题计算稳定 FNV-1a 哈希，从十个通用 Lucide 图标中选一个；分组图标始终不显示文字。图标管理的“自动”选项存空串。图标从 Iconify 的 Lucide SVG 集下载，24×24 viewBox、`currentColor`；许可文本见 `src/assets/svg-icons/LICENSE.lucide`（ISC，部分源自 Feather 的图标附 MIT 条款）。

## 分组色彩系统

- “全部”用中性 ink。其余分组按当前排序以 `index % 10` 循环分配色相，不入库；拖拽排序后颜色随顺序变化。`--xp-hue-1` 至 `--xp-hue-10` 各有 `-fg`、`-bg`、`-bd` 三档。色相跨暖、绿、蓝、紫、红分布，保持相近明度与中低饱和；浅色前景为满足对比度而比示例稍深。浅色 `-bg` 是前景色 13% 与 canvas 混合得到的实色，深色为提亮前景色 20% 与 surface 混合；边框档分别以 42%/48% 混入 canvas。

| 序号 | 色相 | 浅色 `-fg` | 深色 `-fg` |
| --- | --- | --- | --- |
| 1 | 珊瑚 | `#bd6147` | `#d47758` |
| 2 | 琥珀 | `#9f702f` | `#b8883c` |
| 3 | 橄榄 | `#6a853e` | `#849b51` |
| 4 | 青绿 | `#2c8478` | `#40a094` |
| 5 | 湖蓝 | `#3f84b5` | `#5b98c1` |
| 6 | 靛蓝 | `#5b6bc9` | `#7987d2` |
| 7 | 紫 | `#8a63c2` | `#a17bcc` |
| 8 | 莓红 | `#c2577e` | `#cc7391` |
| 9 | 陶土 | `#b0704a` | `#bd8562` |
| 10 | 石板 | `#5f7385` | `#8296a6` |

- 前景图标对相应浅色底的 WCAG 对比度最低约 3.12:1（浅色），深色最低约 3.75:1；选中白色图标对实色底最低约 3.06:1。Rail、底栏、分组标题图标、网址行首字母占位及 Tooltip 圆点使用同一分组色；真实 favicon 保留原色。

## 卡片层次

- 浅色 canvas 为暖灰白 `#f5f3ee`，卡片、概览带系统状态、搜索框、筛选框为实色白 `#ffffff`，配 1px hairline 与极轻阴影 `0 1px 2px rgba(20,20,19,.04), 0 1px 1px rgba(20,20,19,.02)`。卡片悬停时边框变分组 `-bd`、阴影变 `0 4px 16px rgba(20,20,19,.06)`，180ms 过渡且不位移。
- 深色 canvas `#151412`，卡片 surface `#1e1d1a`、hairline `#2e2c28`，阴影为 `inset 0 1px 0 rgba(255,255,255,.03)`；悬停边框同样用分组色。Naive UI 的 body、card、modal、popover 等对应颜色与 CSS 令牌一致。珊瑚主按钮色仍为 `--xp-primary`。
- 分组卡标题 15px、600 字重，左侧 24×24 同色图标，数量 13px muted；单分组标题图标为 28×28。网址行最低 40px，高亮为该组浅色底的一半强度。无渐变、发光、玻璃拟态，减少动态效果设置关闭新增过渡。
