export interface BookmarkGroup {
  title: string
  items: { title: string; url: string }[]
}

export interface BookmarkParseResult {
  groups: BookmarkGroup[]
  skippedInvalid: number
}

export function normalizeUrl(url: string): string {
  const trimmed = url.trim()
  try {
    const parsed = new URL(trimmed)
    const pathname = parsed.pathname.replace(/\/+$/, '')
    const credentials = parsed.username ? `${parsed.username}${parsed.password ? `:${parsed.password}` : ''}@` : ''
    return `${parsed.protocol}//${credentials}${parsed.host}${pathname}${parsed.search}`
  }
  catch {
    return trimmed
  }
}

function folderList(heading: Element): Element | null {
  const folder = heading.closest('dt') ?? heading.parentElement
  if (!folder)
    return null

  const child = Array.from(folder.children).find(element => element.tagName === 'DL')
  if (child)
    return child

  const sibling = folder.nextElementSibling
  if (sibling?.tagName === 'DL')
    return sibling
  if (sibling?.tagName === 'DD')
    return Array.from(sibling.children).find(element => element.tagName === 'DL') ?? null
  return null
}

export function parseBookmarksHtml(html: string): BookmarkParseResult {
  const document = new DOMParser().parseFromString(html, 'text/html')
  const folderNames = new Map<Element, string>()
  for (const heading of Array.from(document.querySelectorAll('h3'))) {
    const list = folderList(heading)
    if (list)
      folderNames.set(list, heading.textContent?.trim() || '未命名文件夹')
  }

  const groups: BookmarkGroup[] = []
  const byTitle = new Map<string, BookmarkGroup>()
  let skippedInvalid = 0
  for (const link of Array.from(document.querySelectorAll('a[href]'))) {
    const url = link.getAttribute('href')?.trim() ?? ''
    if (!/^https?:\/\//i.test(url)) {
      skippedInvalid++
      continue
    }

    const groupTitle = folderNames.get(link.closest('dl') as Element) ?? '书签栏'
    let group = byTitle.get(groupTitle)
    if (!group) {
      group = { title: groupTitle, items: [] }
      byTitle.set(groupTitle, group)
      groups.push(group)
    }

    let hostname = url
    try {
      hostname = new URL(url).hostname
    }
    catch {}
    const title = link.textContent?.trim() || hostname
    group.items.push({ title: Array.from(title).slice(0, 20).join(''), url })
  }
  return { groups, skippedInvalid }
}
