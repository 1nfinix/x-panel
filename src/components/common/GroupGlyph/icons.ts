export const groupIconNames = [
  'home', 'folder', 'star', 'bookmark', 'message-circle', 'bot',
  'sparkles', 'cpu', 'code', 'terminal', 'git-branch', 'cloud',
  'server', 'database', 'shield', 'lock', 'mail', 'credit-card',
  'wallet', 'newspaper', 'rss', 'book-open', 'file-text', 'palette',
  'image', 'video', 'music', 'shopping-cart', 'gamepad-2',
  'graduation-cap', 'briefcase', 'globe', 'wrench', 'layout-grid',
  'compass', 'layers', 'hash', 'box',
] as const

export const groupIcons = groupIconNames.map(name => `lucide-${name}`)

// Earliest keyword in the title wins; this table breaks ties (e.g. 图片 -> image).
export const groupIconRules = [
  { icon: 'bot', keywords: ['ai', '对话', '聊天', 'chat', 'gpt'] },
  { icon: 'cpu', keywords: ['模型', 'api', 'llm', 'model'] },
  { icon: 'sparkles', keywords: ['skill', 'agent', '智能体', '自动化'] },
  { icon: 'newspaper', keywords: ['资讯', '新闻', '社区', '论坛', 'news', 'rss'] },
  { icon: 'book-open', keywords: ['博客', '阅读', '文章', 'book'] },
  { icon: 'image', keywords: ['图片', '摄影', 'photo'] },
  { icon: 'palette', keywords: ['设计', '素材', '灵感', 'design', 'ui', '图片'] },
  { icon: 'video', keywords: ['视频', '影视', 'video'] },
  { icon: 'music', keywords: ['音乐', 'music'] },
  { icon: 'file-text', keywords: ['办公', '文档', '笔记', 'doc', 'office'] },
  { icon: 'cloud', keywords: ['云', '运维', '服务器', 'cloud', 'devops', 'vps'] },
  { icon: 'database', keywords: ['数据库', 'db'] },
  { icon: 'server', keywords: ['服务', 'server', '主机'] },
  { icon: 'shield', keywords: ['安全', 'security', '密码'] },
  { icon: 'mail', keywords: ['邮箱', 'mail'] },
  { icon: 'wallet', keywords: ['支付', '钱包', '银行', 'pay', 'finance', '投资'] },
  { icon: 'code', keywords: ['开发', '代码', '编程', 'dev', 'code', 'github'] },
  { icon: 'terminal', keywords: ['终端', '命令行', 'shell'] },
  { icon: 'wrench', keywords: ['工具', 'tool', '实用'] },
  { icon: 'folder', keywords: ['项目', 'project'] },
  { icon: 'bookmark', keywords: ['书签', '收藏', 'bookmark'] },
  { icon: 'home', keywords: ['常用', '首页', 'home'] },
  { icon: 'graduation-cap', keywords: ['学习', '教程', '课程', 'learn'] },
  { icon: 'briefcase', keywords: ['工作', '职业', 'job'] },
  { icon: 'shopping-cart', keywords: ['购物', '电商', 'shop'] },
  { icon: 'gamepad-2', keywords: ['游戏', 'game'] },
  { icon: 'globe', keywords: ['网站', '导航', 'web', 'global'] },
] as const

const fallbackIcons = [
  'star', 'bookmark', 'folder', 'globe', 'sparkles',
  'layout-grid', 'compass', 'layers', 'hash', 'box',
] as const
const spriteNames = new Set(Object.keys(import.meta.glob('../../../assets/svg-icons/*.svg'))
  .map(path => path.split('/').pop()?.replace(/\.svg$/, '')))
const legacyIcons = new Set([
  'material-symbols:ad-group-outline-rounded',
  'material-symbols:ad-group-outline',
])

function keywordIndex(title: string, keyword: string) {
  let index = title.indexOf(keyword)
  while (index !== -1) {
    const shortToken = ['ai', 'ui', 'db'].includes(keyword)
    const embeddedBookmark = keyword === 'book' && title.startsWith('bookmark', index)
    if (!embeddedBookmark && (!shortToken
      || (!/[a-z]/.test(title[index - 1] ?? '') && !/[a-z]/.test(title[index + keyword.length] ?? ''))))
      return index
    index = title.indexOf(keyword, index + 1)
  }
  return -1
}

export function autoGroupIcon(title = ''): string {
  const normalized = title.normalize('NFKC').toLocaleLowerCase()
  let earliest = Number.POSITIVE_INFINITY
  let match = ''
  for (const rule of groupIconRules) {
    for (const keyword of rule.keywords) {
      const index = keywordIndex(normalized, keyword)
      if (index >= 0 && index < earliest) {
        earliest = index
        match = rule.icon
      }
    }
  }
  if (match)
    return `lucide-${match}`

  // FNV-1a: independent of session, browser and group ordering.
  let hash = 2166136261
  for (const character of normalized)
    hash = Math.imul(hash ^ (character.codePointAt(0) ?? 0), 16777619) >>> 0
  return `lucide-${fallbackIcons[hash % fallbackIcons.length]}`
}

export function isValidGroupIcon(icon?: string): boolean {
  const spriteName = icon?.replace(/:/g, '-')
  return !!(icon && spriteName && !legacyIcons.has(icon) && spriteNames.has(spriteName))
}

export function resolveGroupIcon(icon?: string, title?: string): string {
  return isValidGroupIcon(icon) ? icon! : autoGroupIcon(title)
}
