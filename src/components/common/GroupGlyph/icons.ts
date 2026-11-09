export const groupIconNames = [
  'home', 'folder', 'star', 'bookmark', 'message-circle', 'bot',
  'sparkles', 'cpu', 'code', 'terminal', 'git-branch', 'cloud',
  'server', 'database', 'shield', 'lock', 'mail', 'credit-card',
  'wallet', 'newspaper', 'rss', 'book-open', 'file-text', 'palette',
  'image', 'video', 'music', 'shopping-cart', 'gamepad-2',
  'graduation-cap', 'briefcase', 'globe', 'wrench', 'layout-grid',
] as const

export const groupIcons = groupIconNames.map(name => `lucide-${name}`)
