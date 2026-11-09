# X-Panel 视觉改版规范（2026-11-09）

## 来源与边界

视觉方向参考 `/Users/vincent/.workbuddy-config-backups/20261031-shared-slim/skills/claude-design-card/DESIGN.md`，用于 X-Panel 的 Workbench 式布局、暖色表面和字体层级。产品始终称 X-Panel；界面与图标不使用 Anthropic 的放射星形标志、名称或其他品牌资产。

用户于 2026-11-09 知情选定此规范。与外脑 `DESIGN.md` 的“圆角不超过 8px”相比，本页允许卡片 12px 圆角；其余元素依旧优先使用紧凑、清晰的层级。品牌标题字距按后续首页概览带规范调整为 0。

## 颜色令牌

`src/styles/global.less` 是 CSS 令牌来源。`src/hooks/useTheme.ts` 的 Naive UI 对应色值必须同步维护。

| 令牌 | 浅色 | 深色 |
| --- | --- | --- |
| canvas | `#faf9f5` | `#181715` |
| surface-soft | `#f5f0e8` | `#1f1e1b` |
| surface-card | `#efe9de` | `#252320` |
| surface-strong | `#e8e0d2` | `#2e2b27` |
| hairline | `#e6dfd8` | `#34312c` |
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

- `category-tab`：透明未选中状态、muted 文本；选中为 surface-card、ink、8px 圆角。
- `text-input`：canvas 底、hairline 边、40px 高、8px 圆角；焦点为 primary 边及 3px focus-ring。
- `secondary` 按钮：canvas 底、hairline 边、ink 字。`primary` 按钮：primary 底、on-primary 字。
- `connector-tile`：canvas 底、hairline 边、12px 圆角、16px 内边距；网址行悬停使用 surface-soft。
- 壁纸开关：默认关闭，首页和登录页显示纯色 canvas；打开后使用已有壁纸与暖黑遮罩，设置面板才显示壁纸参数。

## 删除的旧代码

删除未被调用的首页 Result、AppIcon，通用 HoverButton、JsonImportExport、Verification 组件，测试路由及页面，旧 Markdown/代码高亮样式与其专用依赖，以及未使用的 `vuedraggable`。仍使用的 `src/utils/jsonImportExport` 与 `vue-draggable-plus` 保留。

## 首页概览带与品牌字体（2026-11-09）

- 顶栏保留品牌、筛选和整理；概览带按时钟、搜索、系统状态顺序排列，与分组区同宽。三个模块全关时不渲染概览带；任意组合不保留空列。
- 宽度至少 1024px 时一行布局；640–1023px 时钟与系统状态在第一行、搜索独占第二行；不足 640px 时纵向堆叠。顶栏到概览带、概览带到分组标签均为 24px，标签到分组区为 16px。
- 时钟使用 44px 表格式数字与 14px 日期，手机缩为 30px 并与日期同行。系统状态采用标签、百分比和 3px 进度条，超过 85% 用 error 色；编辑、排序与右键操作保留。分组总览采用约 320px 列宽的 CSS columns，分组卡避免跨列断开；独立 VueDraggable 列表和空组放置区保留。
- 品牌字与登录、About 标题使用自托管 Noto Serif SC Variable，字重 600，字距 0。字体按 unicode-range 分片加载。品牌桌面 28px，手机 24px。
- 网址行悬停 120ms 背景过渡，箭头位移 2px；筛选框支持方向键聚焦可见网址、Esc 清空并失焦。首屏卡片依次以最多 400ms 的淡入与 6px 上移动画出现；切换分组时仅做 150ms 淡入。减少动态效果设置时关闭这些动画。

## 分组侧栏（Icon Rail，2026-11-09）

- 桌面从 640px 起使用固定在视口左侧的 64px 全高图标栏，canvas 背景与右侧 hairline；壁纸开启时背景为 88% canvas 混合透明。主内容在剩余宽度中按原 maxWidth 居中。品牌图 28px、圆角 8px，固定在顶部，点按回到“全部”并滚动到顶部；其下间距 16px。分组区独立纵向滚动，隐藏滚动条。
- “全部”使用 Lucide layout-grid，后接 24px hairline，再按排序展示分组。按钮 40×40px、圆角 8px、图标 20px、间距 4px；默认 muted，悬停 surface-soft/ink，选中 surface-card/ink。最左侧珊瑚色 3×20px 指示条随选中项以 transform 移动 180ms；按钮状态过渡 120ms。减少动态效果时关闭过渡。
- 桌面 tooltip 在按钮右侧 12px，使用 ink 背景、canvas 字色、13px 字号、6px 圆角与 6×10px 内边距，无阴影；数量使用淡化的反色与表格式数字。首次悬停延迟 300ms，栏内相邻按钮立即切换，离开栏 100ms 后关闭，淡入及向右 4px 位移 120ms；浮层挂到 body，避免滚动裁切。焦点进入时立即展示。按钮提供可读名称及 `aria-current`，↑/↓、Home/End 移动 roving 焦点，Enter/Space 原生激活。
- 小于 640px 时图标栏固定底部，高 56px 加底部安全区；横向滚动并隐藏滚动条，选中按钮自动滚入可见范围。指示条变为顶部 20×3px；触屏不显示 tooltip。内容底部预留栏高，悬浮按钮和回顶按钮上移。单分组视图在顶栏、概览带之下显示 18px 分组标题与数量，桌面同样展示；“全部”不展示标题。
- 分组图标为空、为历史默认广告分组图标、或本地 sprite 不存在时，用标题首字符作 15px、字重 600 的思源宋体字标；英文转大写，不加额外底色。新建默认首字，编辑可在 8 列网格中选首字或本地图标。图标来自 Iconify 的 Lucide SVG 集，保存在 `src/assets/svg-icons/lucide-*.svg`，使用 `currentColor` 和 24×24 viewBox；Lucide 使用 ISC 许可，许可文本见同目录 `LICENSE.lucide`。
